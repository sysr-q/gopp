package main

import (
	"io"
	"os"
	"fmt"
	"strings"
	"github.com/droundy/goopt"
	"github.com/sysr-q/gopp/gopp"
)

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

	goopt.Description = func() string {
		return Description
	}
	goopt.Version = Version
	goopt.Summary = Description
	goopt.Parse(nil)

	if len(goopt.Args) < 1 {
		fmt.Println("Supply a file to process! Here's --help:")
		fmt.Print(goopt.Help())
		return nil
	}

	g := gopp.New(true)
	g.StripComments = !*stripComments
	g.DefineValue("_GPPC", Version)
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

	isDir, err := isDirectory(goopt.Args[0])
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
