package main

import (
	"io"
	"os"
	"fmt"
	"path"
	"strings"
	"path/filepath"
	"github.com/droundy/goopt"
	"github.com/sysr-q/gopp/gopp"
)

func prepDir(in, out string, g *gopp.Gopp, extensions, ignores *[]string) error {
	return filepath.Walk(in, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("%s (1): %s -- skipping\n", walkPath, err)
			return nil
		}

		// Ignore stuff we're building to avoid recursion.
		for _, ignore := range *ignores {
			if strings.HasPrefix(walkPath, ignore) {
				return nil
			}
		}

		if info.IsDir() {
			newDir := path.Join(out, walkPath)
			if err = os.MkdirAll(newDir, os.ModeDir | 0755); err != nil {
				fmt.Printf("%s (2): %s -- skipping\n", walkPath, err)
				return nil
			}
		}

		ext, extValid := filepath.Ext(walkPath), false
		for _, e := range *extensions {
			if e == ext {
				extValid = true
				break
			}
		}
		if !extValid {
			return nil
		}

		fileIn, err := os.Open(walkPath)
		if err != nil {
			fmt.Printf("%s (3): %s -- skipping\n", walkPath, err)
			return nil
		}
		defer fileIn.Close()

		defer g.Reset()
		if err = g.Parse(fileIn); err != nil {
			fmt.Printf("%s (4): %s -- skipping\n", walkPath, err)
			return nil
		}

		fileOut, err := os.Create(path.Join(out, walkPath))
		if err != nil {
			fmt.Printf("%s (5): %s -- skipping\n", walkPath, err)
			return nil
		}
		defer fileOut.Close()

		if err = g.Print(fileOut); err != nil {
			fmt.Printf("%s (6): %s -- skipping\n", walkPath, err)
			return nil
		}
		return nil
	})
}

func prepFile(in, out string, g *gopp.Gopp) (err error) {
	var fileIn io.Reader
	if in == "-" {
		fileIn = os.Stdin
	} else {
		fileIn, err = os.Open(in)
		if err != nil {
			return err
		}
	}

	if err := g.Parse(fileIn); err != nil {
		return err
	}

	var fileOut io.Writer
	if out == "-" {
		fileOut = os.Stdout
	} else {
		fileOut, err = os.Create(out)
		if err != nil {
			return err
		}
	}

	return g.Print(fileOut)
}

func prep() error {
	defined := goopt.Strings([]string{"-D"},
		"NAME[=defn]",
		"Predefine NAME as a macro. Unless given, default macro value is 1.")
	undefined := goopt.Strings([]string{"-U"},
		"NAME",
		"Cancel any previous/builtin definition of macro NAME.")
	extensions := goopt.Strings([]string{"-e"},
		".go",
		"Add file extensions that will be processed. (default: .go)")
	ignores := goopt.Strings([]string{"-i"},
		"_gppc",
		"Add directories/files to avoid when prepping.")
	stripComments := goopt.Flag([]string{"-c", "--comments"},
		[]string{"-C", "--no-comments"},
		"Don't eat comments.",
		"Eat any comments that are found.")
	outputTo := goopt.String([]string{"-o"},
		"_gppc",
		"Output directory (default: _gppc, or <stdout> for files)")

	goopt.Description = func() string {
		return Description
	}
	goopt.Version = Version
	goopt.Summary = Description
	goopt.Parse(nil)

	*extensions = append(*extensions, ".go")
	*ignores = append(*ignores, *outputTo)
	*defined = append([]string{fmt.Sprintf("_GPPC=%q", Version)}, *defined...)

	if len(goopt.Args) < 1 {
		fmt.Println("Supply a directory or file to process! Here's --help:")
		fmt.Print(goopt.Help())
		return nil
	}

	g := gopp.New(true)
	g.StripComments = !*stripComments
	g.DefineValue("_GPPC_DEFINES", fmt.Sprintf("`-D %s`", strings.Join(*defined, " -D ")))
	for _, def := range *defined {
		if strings.Contains(def, "=") {
			lnr := strings.SplitN(def, "=", 2)
			g.DefineValue(lnr[0], lnr[1])
		} else {
			g.Define(def)
		}
	}

	for _, udef := range *undefined {
		g.Undefine(udef)
	}

	prepPath := goopt.Args[0]
	isDir, err := isDirectory(prepPath)
	if err != nil {
		return err
	}

	if isDir {
		if strings.Contains(prepPath, "..") {
			fmt.Println("Warning! This may be interesting if you've got '..' in the prep path.")
		}
		return prepDir(prepPath, *outputTo, g, extensions, ignores)
	}

	// `_gppc` -> `-` (<stdout>)
	if *outputTo == "_gppc" {
		*outputTo = "-"
	}

	return prepFile(prepPath, *outputTo, g)
}
