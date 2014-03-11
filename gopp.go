package main

import (
	"bytes"
	"fmt"
	"github.com/droundy/goopt"
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

			//fmt.Printf("%q %q %s %v\n", strTrim, cmd, lnr, g.ignoring)

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

func NewGopp(strip bool) *gopp {
	return &gopp{
		defines:       make(map[string]interface{}),
		output:        make(chan Token),
		StripComments: strip,
		ignoring:      false,
		Prefix:        "//gopp:",
	}
}

///////////////////

func IsDirectory(path string) (bool, error) {
	dir, err := os.Stat(path)
	if err == nil {
		return dir != nil && dir.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, err
	}
	return false, err
}

func build() error {
	defined := goopt.Strings(
		[]string{"-D"},
		"NAME[=defn]",
		"Predefine NAME as a macro. Unless given, default macro value is 1.")
	undefined := goopt.Strings([]string{"-U"},
		"NAME",
		"Cancel any previous/builtin definition of macro NAME.")
	stripComments := goopt.Flag([]string{"-c", "--comments"},
		[]string{"-C", "--no-comments"},
		"Don't eat comments.",
		"Eat any comments that are found.")
	outputTo := goopt.String([]string{"-o"},
		"-",
		"Output (default: <stdout> for files, `_build` for directories)")
	/*panicErrors := goopt.Flag([]string{"-p", "--panic"},
	[]string{},
	"panic() on errors, rather than just logging them.",
	"")*/

	// Because you can't supply an arg list to goopt..
	if len(os.Args) > 2 {
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
	} else {
		os.Args = os.Args[:1]
	}

	goopt.Description = func() string {
		return "What have I done?!"
	}
	goopt.Version = "0.2"
	goopt.Summary = "Horrifying C-like Go preprocessor."
	goopt.Parse(nil)

	if len(goopt.Args) < 1 {
		fmt.Println("Supply a file to process! Here's --help:")
		fmt.Print(goopt.Help())
		return nil
	}

	g := NewGopp(true)
	g.StripComments = !*stripComments
	g.DefineValue("_GOPP", goopt.Version)
	for _, def := range *defined {
		if strings.Contains(def, "=") {
			lnr := strings.SplitN(def, "=", 2)
			g.DefineValue(lnr[0], lnr[1])
		} else {
			g.DefineValue(def, 1)
		}
	}

	for _, udef := range *undefined {
		g.Undefine(udef)
	}

	isDir, err := IsDirectory(goopt.Args[0])
	if err != nil {
		return err
	}

	if isDir {
		// TODO: Handle walking path and parsing.
	} else {
		file, err := os.Open(goopt.Args[0])
		if err != nil {
			return err
		}

		if err := g.Parse(file); err != nil {
			return err
		}

		var out io.Writer
		if *outputTo == "-" {
			out = os.Stdout
		} else {
			out, err = os.Open(*outputTo)
			if err != nil {
				return err
			}
		}
		g.Print(out)
	}
	return nil
}

func clean() error {
	return nil
}

var commands = map[string]func() error{
	"build": build,
	"clean": clean,
}

func main() {
	printWrong := func() {
		fmt.Println("Wrong! Try one of these: (you can append --help)")
		for c := range commands {
			fmt.Printf("  %s %s\n", os.Args[0], c)
		}
	}

	if len(os.Args) < 2 {
		printWrong()
		return
	}

	f, ok := commands[os.Args[1]]
	if !ok {
		printWrong()
		return
	}
	if err := f(); err != nil {
		panic(err)
	}
}
