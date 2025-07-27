package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

func main() {
	dataDir := "data"
	schemaDir := "data/jsonschema"

	// Get all subdirectories in the data directory
	dataSubDirs, err := ioutil.ReadDir(dataDir)
	if err != nil {
		fmt.Printf("Error reading data directory: %s\n", err.Error())
		os.Exit(1)
	}

	for _, dirInfo := range dataSubDirs {
		if dirInfo.IsDir() && dirInfo.Name() != "jsonschema" {
			// Determine the schema name from the directory name
			dirName := dirInfo.Name()
			schemaName := strings.TrimSuffix(dirName, "s")
			schemaName = strings.Title(schemaName) + ".json"
			schemaPath := filepath.Join(schemaDir, schemaName)

			// Check if the schema file exists
			if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
				fmt.Printf("Skipping directory %s: schema %s not found\n", dirName, schemaPath)
				continue
			}

			// Get all json files in the directory
			jsonFiles, err := filepath.Glob(filepath.Join(dataDir, dirName, "*.json"))
			if err != nil {
				fmt.Printf("Error finding json files in %s: %s\n", dirName, err.Error())
				continue
			}

			// Validate each json file
			for _, jsonFile := range jsonFiles {
				validateFile(schemaPath, jsonFile)
			}
		}
	}
}

func validateFile(schemaPath, docPath string) {
	schemaAbsPath, _ := filepath.Abs(schemaPath)
	docAbsPath, _ := filepath.Abs(docPath)

	schemaLoader := gojsonschema.NewReferenceLoader("file://" + strings.ReplaceAll(schemaAbsPath, "\\", "/"))
	documentLoader := gojsonschema.NewReferenceLoader("file://" + strings.ReplaceAll(docAbsPath, "\\", "/"))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		fmt.Printf("Error validating %s: %s\n", docPath, err.Error())
		return
	}

	if result.Valid() {
		fmt.Printf("%s is valid\n", docPath)
	} else {
		fmt.Printf("%s is not valid. see errors:\n", docPath)
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
	}
}
