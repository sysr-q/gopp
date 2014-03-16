package main

import (
	"os"
	"fmt"
)

const (
	Version = "1.1.1"
	Description = "Horrifying C-like Go preprocessor."
)

var commands = map[string]func() error{
	"prep": prep,
	"clean": clean,
}

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

func main() {
	printWrong := func() {
		fmt.Println("That's not a command! Try one of these: (you can append --help)")
		for c := range commands {
			fmt.Printf("  %s %s\n", os.Args[0], c)
		}
	}

	if len(os.Args) < 2 {
		printWrong()
		return
	}

	// Strip the command out of os.Args so goopt won't choke.
	cmd := os.Args[1]
	if len(os.Args) > 2 {
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
	} else {
		os.Args = os.Args[:1]
	}

	f, ok := commands[cmd]
	if !ok {
		printWrong()
		return
	}
	if err := f(); err != nil {
		fmt.Println("Failed:", err)
	}
}
