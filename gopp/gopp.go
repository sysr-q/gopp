// gopp provides an easy to use (but somewhat illogical) preprocessor for Go.
// It allows similar functionality to cpp, such as ifdef, ifndef, define, undef
// and other odd combinations thereof.
//
// Note: There is currently no support for nested ifdef/ifndefs.
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
	"strings"
)

const Version = "0.4.2"

// Token provides a simple struct face to the `pos, tok, lit` returned by
// Go's go/scanner package. This is passed around by gopp internally in a chan.
type Token struct {
	Position token.Pos
	Token    token.Token
	String   string
}

type Gopp struct {
	Macros                  map[string]interface{}
	output                  chan Token
	StripComments, ignoring bool
	Prefix                  string
}

// DefineValue takes a name and a value, and defines them as a macro to use when
// preprocessing a Go file.
func (g *Gopp) DefineValue(name string, value interface{}) {
	g.Macros[name] = value
}

// Define takes a name, and sets it to the sane default of 1, which can then be
// used when preprocessing a Go file.
func (g *Gopp) Define(name string) {
	g.DefineValue(name, 1)
}

// Undefine will remove a given name from the macro list used when processing
// Go files.
func (g *Gopp) Undefine(name string) {
	delete(g.Macros, name)
}

// Parse takes an io.Reader (usually from an os.Open call), and will process the
// resulting code read from it, preprocessing and substituting in macros as
// required.
func (g *Gopp) Parse(r io.Reader) error {
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
				val, ok := g.Macros[tok.String]
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
				_, ok := g.Macros[def]
				g.ignoring = !ok
			} else if cmd == "ifndef" {
				if len(lnr) != 2 {
					continue
				}
				def := lnr[1]
				_, ok := g.Macros[def]
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

// Print takes the resulting new AST after a call to Parse, and will print it
// to the given io.Writer
func (g *Gopp) Print(w io.Writer) error {
	outbuf := new(bytes.Buffer)
	for tok := range g.output {
		fmt.Fprintf(outbuf, " %s", tok.String)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "<stdin>", outbuf, parser.ParseComments)
	if err != nil {
		return err
	}

	return printer.Fprint(w, fset, file)
}

// Reset will redefine the bits of a gopp instance, which you can use (e.g.)
// each time you want to parse a new file.
func (g *Gopp) Reset() {
	g.Macros = make(map[string]interface{})
	g.output = make(chan Token)
	g.ignoring = false
}

// New creates a new gopp instance with sane defaults. The macro `_GOPP` is
// automatically set to the gopp version.
func New(strip bool) (g *Gopp) {
	g = &Gopp{
		Macros:        make(map[string]interface{}),
		output:        make(chan Token),
		StripComments: strip,
		ignoring:      false,
		Prefix:        "//gopp:",
	}
	g.DefineValue("_GOPP", Version)
	return
}
