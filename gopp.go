package main

import (
	"io"
	"os"
	"fmt"
	"bytes"
	"strings"
	"io/ioutil"
	"go/token"
	"go/scanner"
	"go/parser"
	"go/printer"
	"github.com/droundy/goopt"
)

type gopp struct {
	defines map[string]interface{}
	output chan string
	StripComments, ignoring bool
	Prefix string
}

func (g *gopp) DefineValue(key string, value interface{}) {
	g.defines[key] = value
}

func (g *gopp) Define(key string) {
	g.DefineValue(key, nil)
}

func (g *gopp) Undefine(key string) {
	delete(g.defines, key)
}

func (g *gopp) Parse(r io.Reader) error {
	src, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	file := fset.AddFile("<stdin>", fset.Base(), len(src))

	s := scanner.Scanner{}
	s.Init(file, src, nil, scanner.ScanComments)

	go func() {
		for {
			_, tok, str := s.Scan()
			if len(str) == 0 { str = tok.String() }
			if tok == token.EOF {
				break
			}

			if tok != token.COMMENT && !g.ignoring {
				val, ok := g.defines[str]
				if ok {
					g.output <- val.(string)
				} else {
					g.output <- str
				}
				continue
			}

			if !strings.HasPrefix(str, g.Prefix) {
				if !g.StripComments && !g.ignoring {
					g.output <- str + "\n"
				}
				continue
			}

			// Trim the prefix from the start.
			strTrim := strings.Replace(str, g.Prefix, "", 1)
			lnr := strings.SplitN(strTrim, " ", 2)
			if len(lnr) < 1 {
				fmt.Println("Invalid gopp comment:", str)
				continue
			}

			cmd := strings.ToLower(lnr[0])

			//fmt.Printf("%q %q %s %v\n", strTrim, cmd, lnr, g.ignoring)

			if cmd == "ifdef" {
				if len(lnr) != 2 { continue }
				def := lnr[1]
				_, ok := g.defines[def]
				g.ignoring = !ok
			} else if cmd == "ifndef" {
				if len(lnr) != 2 { continue }
				def := lnr[1]
				_, ok := g.defines[def]
				g.ignoring = ok
			} else if cmd == "else" {
				g.ignoring = !g.ignoring
			} else if cmd == "endif" {
				if !g.ignoring {
					continue
				}
				g.ignoring = false
			} else if cmd == "define" {
				if g.ignoring { continue }
				if len(lnr) != 2 { continue }
				lnr = strings.SplitN(lnr[1], " ", 2)
				g.DefineValue(lnr[0], lnr[1])
			} else if cmd == "undef" {
				if g.ignoring { continue }
				if len(lnr) != 2 { continue }
				g.Undefine(lnr[1])
			}

		}
		close(g.output)
	}()
	return nil
}

func (g *gopp) Print(w io.Writer) {
	outbuf := new(bytes.Buffer)
	for tok := range g.output {
		fmt.Fprintf(outbuf, " %s", tok)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "<stdin>", outbuf, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	printer.Fprint(os.Stdout, fset, file)
}

func NewGopp(strip bool) *gopp {
	return &gopp{
		defines: make(map[string]interface{}),
		output: make(chan string),
		StripComments: strip,
		ignoring: false,
		Prefix: "//gopp:",
	}
}

///////////////////
var defines = goopt.Strings([]string{"-D"}, "GOLANG", "Define some magic.")
var stripComments = goopt.Flag([]string{"-c", "--comments"}, []string{"-C", "--no-comments"}, "", "")

func main() {
	goopt.Description = func() string {
		return "What have I done?!"
	}
	goopt.Version = "0.1"
	goopt.Summary = "Horrifying C-like Go preprocessor."
	goopt.Parse(nil)

	if len(goopt.Args) < 1 {
		fmt.Println("Supply a file to process!")
		return
	}

	file, err := os.Open(goopt.Args[0])
	if err != nil {
		panic(err)
	}

	g := NewGopp(true)
	g.StripComments = !*stripComments
	g.DefineValue("GOPP_VER", goopt.Version)
	for _, def := range *defines {
		if strings.Contains(def, "=") {
			lnr := strings.SplitN(def, "=", 2)
			g.DefineValue(lnr[0], lnr[1])
		} else {
			g.Define(def)
		}
	}

	if err := g.Parse(file); err != nil {
		fmt.Println(err)
		return
	}
	g.Print(os.Stdout)
}
