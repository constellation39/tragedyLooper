package main

import (
	"io"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		panic("usage: writer <file>")
	}
	file, err := os.Create(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()
	_, err = io.Copy(file, os.Stdin)
	if err != nil {
		panic(err)
	}
}
