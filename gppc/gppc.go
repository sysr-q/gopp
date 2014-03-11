package main

import (
	"fmt"
	"github.com/sysr-q/gopp/gopp"
)

func isDirectory(path string) (bool, error) {
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

	g := New(true)
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
