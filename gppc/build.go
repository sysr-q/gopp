package main

import (
	"os"
	"fmt"
	"path"
	"strings"
	"path/filepath"
	"github.com/droundy/goopt"
	"github.com/sysr-q/gopp/gopp"
)

func build() error {
	defined := goopt.Strings([]string{"-D"},
		"NAME[=defn]",
		"Predefine NAME as a macro. Unless given, default macro value is 1.")
	undefined := goopt.Strings([]string{"-U"},
		"NAME",
		"Cancel any previous/builtin definition of macro NAME.")
	extensions := goopt.Strings([]string{"-e"},
		".go",
		"Add file extensions that will be processed. (default: .go)")
	stripComments := goopt.Flag([]string{"-c", "--comments"},
		[]string{"-C", "--no-comments"},
		"Don't eat comments.",
		"Eat any comments that are found.")
	outputTo := goopt.String([]string{"-o"},
		"_build",
		"Output directory (default: ./_build")

	goopt.Description = func() string {
		return Description
	}
	goopt.Version = Version
	goopt.Summary = Description
	goopt.Parse(nil)

	*extensions = append(*extensions, ".go")

	if len(goopt.Args) < 1 {
		fmt.Println("Supply a directory to process! Here's --help:")
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

	buildPath := goopt.Args[0]
	isDir, err := isDirectory(buildPath)
	if err != nil {
		return err
	}
	if !isDir {
		return fmt.Errorf("%s: not a directory", buildPath)
	}

	err = filepath.Walk(buildPath, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("%s (1): %s -- skipping\n", walkPath, err)
			return nil
		}

		// Ignore stuff we're building to avoid recursion.
		if strings.HasPrefix(walkPath, *outputTo) {
			return nil
		}

		if info.IsDir() {
			newDir := path.Join(*outputTo, walkPath)
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

		in, err := os.Open(walkPath)
		if err != nil {
			fmt.Printf("%s (3): %s -- skipping\n", walkPath, err)
			return nil
		}

		defer g.Reset()
		if err = g.Parse(in); err != nil {
			fmt.Printf("%s (4): %s -- skipping\n", walkPath, err)
			return nil
		}

		out, err := os.Create(path.Join(*outputTo, walkPath))
		defer out.Close()
		if err != nil {
			fmt.Printf("%s (5): %s -- skipping\n", walkPath, err)
			return nil
		}

		if err = g.Print(out); err != nil {
			fmt.Printf("%s (6): %s -- skipping\n", walkPath, err)
			return nil
		}
		return nil
	})
	if err != nil {
		return err
	}

	/*file, err := os.Open(goopt.Args[0])
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
	g.Print(out)*/
	return nil
}
