package main

import (
	"os"
	"fmt"
)

//gopp:ifdef WINDOWS
	//gopp:define OS "Windows"
	// Fake, just for example.
	import "_win32"
	const Thoughts = "Why?!?"
//gopp:else
	//gopp:define OS "Linux"
	const Thoughts = "I think?"
//gopp:endif

func main() {
	fmt.Printf("Greets from %s! (%s)\n", OS, Thoughts)
	// God help you if you didn't use gppc.
	fmt.Printf("gppc'd using version: %v\n", _GPPC)
	for i := 0; i < 5; i++ {
		// Do something, it doesn't matter what!
		// So let's flood stdout for eternity.
		fmt.Println("Greets from", os.Args[0])
	}
}
// vim: syntax=go
