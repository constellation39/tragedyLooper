package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	protoDir := "proto/v1"
	files, err := ioutil.ReadDir(protoDir)
	if err != nil {
		fmt.Printf("Error reading proto directory: %v\n", err)
		os.Exit(1)
	}

	var content strings.Builder

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".proto") {
			content.WriteString(fmt.Sprintf("// --- %s ---\n", file.Name()))
			filePath := filepath.Join(protoDir, file.Name())
			fileContent, err := ioutil.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", filePath, err)
				os.Exit(1)
			}
			content.Write(fileContent)
			content.WriteString("\n\n")
		}
	}

	tmpFile, err := ioutil.TempFile("", "proto-dump-*.tmp")
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		os.Exit(1)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content.String()); err != nil {
		fmt.Printf("Error writing to temp file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Proto files dumped to: %s\n", tmpFile.Name())
}
