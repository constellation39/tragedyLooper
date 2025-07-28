package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	protoDir := "proto/v1"
	files, err := os.ReadDir(protoDir)
	if err != nil {
		return fmt.Errorf("reading proto directory: %w", err)
	}

	var content strings.Builder

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".proto") {
			content.WriteString(fmt.Sprintf("// --- %s ---", file.Name()))
			filePath := filepath.Join(protoDir, file.Name())
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %s: %w", filePath, err)
			}
			content.Write(fileContent)
			content.WriteString("")
		}
	}

	tmpFile, err := os.CreateTemp("", "proto-dump-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content.String()); err != nil {
		return fmt.Errorf("writing to temp file: %w", err)
	}

	fmt.Printf("Proto files dumped to: %s", tmpFile.Name())
	return nil
}
