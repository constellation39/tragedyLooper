package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: rmrf <path>")
		os.Exit(1)
	}
	path := os.Args[1]
	if err := os.RemoveAll(path); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
