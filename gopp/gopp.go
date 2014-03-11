package gopp

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/printer"
	"go/scanner"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type Token struct {
	Position token.Pos
	Token    token.Token
	String   string
}

type gopp struct {
	defines                 map[string]interface{}
	output                  chan Token
	StripComments, ignoring bool
	Prefix                  string
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
			pos, tok_, str := s.Scan()
			if len(str) == 0 {
				str = tok_.String()
			}
			tok := Token{pos, tok_, str}

			if tok.Token == token.EOF {
				break
			}

			if tok.Token != token.COMMENT && !g.ignoring {
				val, ok := g.defines[tok.String]
				if ok {
					tok.String = val.(string)
				}
				g.output <- tok
				continue
			}

			if !strings.HasPrefix(str, g.Prefix) {
				if !g.StripComments && !g.ignoring {
					tok.String += "\n"
					g.output <- tok
				}
				continue
			}

			// Trim the prefix from the start.
			strTrim := strings.Replace(tok.String, g.Prefix, "", 1)
			lnr := strings.SplitN(strTrim, " ", 2)
			if len(lnr) < 1 {
				fmt.Println("Invalid gopp comment:", str)
				continue
			}

			cmd := strings.ToLower(lnr[0])

			if cmd == "ifdef" {
				if len(lnr) != 2 {
					continue
				}
				def := lnr[1]
				_, ok := g.defines[def]
				g.ignoring = !ok
			} else if cmd == "ifndef" {
				if len(lnr) != 2 {
					continue
				}
				def := lnr[1]
				_, ok := g.defines[def]
				g.ignoring = ok
			} else if cmd == "else" {
				g.ignoring = !g.ignoring
			} else if cmd == "endif" && g.ignoring {
				g.ignoring = false
			} else if cmd == "define" && !g.ignoring {
				if len(lnr) != 2 {
					continue
				}
				lnr = strings.SplitN(lnr[1], " ", 2)
				g.DefineValue(lnr[0], lnr[1])
			} else if cmd == "undef" && !g.ignoring {
				if len(lnr) != 2 {
					continue
				}
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
		fmt.Fprintf(outbuf, " %s", tok.String)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "<stdin>", outbuf, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	printer.Fprint(os.Stdout, fset, file)
}

// This resets the bits of the gopp which you should really redefine
// each time you want to parse a new file.
func (g *gopp) Reset() {
	g.defines = make(map[string]interface{})
	g.output = make(chan Token)
	g.ignoring = false
}

// New creates a new gopp instance with sane defaults.
func New(strip bool) *gopp {
	return &gopp{
		defines:       make(map[string]interface{}),
		output:        make(chan Token),
		StripComments: strip,
		ignoring:      false,
		Prefix:        "//gopp:",
	}
}