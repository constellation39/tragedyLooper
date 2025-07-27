package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

func main() {
	dataDir := "data"
	schemaDir := "data/jsonschema"

	// Automatically discover and validate data files based on schemas.
	fmt.Println("Starting validation...")
	discoverAndValidate(dataDir, schemaDir)
	fmt.Println("Validation finished.")
}

func discoverAndValidate(dataDir, schemaDir string) {
	schemaFiles, err := filepath.Glob(filepath.Join(schemaDir, "*.json"))
	if err != nil {
		fmt.Printf("Error finding schema files in %s: %s\n", schemaDir, err.Error())
		os.Exit(1)
	}

	validatedCount := 0
	for _, schemaPath := range schemaFiles {
		var dataPath string
		schemaName := filepath.Base(schemaPath)

		// Rule 1: For "XxxConfigLib.json", validate "xxx.json"
		if strings.HasSuffix(schemaName, "ConfigLib.json") {
			dataFileName := strings.TrimSuffix(schemaName, "ConfigLib.json")
			dataFileName = strings.ToLower(dataFileName) + ".json"
			dataPath = filepath.Join(dataDir, dataFileName)
		} else if strings.HasSuffix(schemaName, "Config.json") {
			// Rule 2: For "XxxConfig.json", validate "XxxConfig.json" in the data root
			dataPath = filepath.Join(dataDir, schemaName)
		} else {
			// Skip schemas that don't represent a main data file (e.g., enums, individual objects)
			continue
		}

		// Check if the inferred data file exists before trying to validate
		if _, err := os.Stat(dataPath); os.IsNotExist(err) {
			// This is not an error, just means there's no corresponding data file for this schema.
			continue
		}

		validateFile(schemaPath, dataPath)
		validatedCount++
	}

	if validatedCount == 0 {
		fmt.Println("No data files were found to validate.")
	}
}

func validateFile(schemaPath, docPath string) {
	schemaAbsPath, _ := filepath.Abs(schemaPath)
	docAbsPath, _ := filepath.Abs(docPath)

	schemaLoader := gojsonschema.NewReferenceLoader("file://" + filepath.ToSlash(schemaAbsPath))
	documentLoader := gojsonschema.NewReferenceLoader("file://" + filepath.ToSlash(docAbsPath))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		fmt.Printf("Error validating %s against %s: %s\n", docPath, schemaPath, err.Error())
		return
	}

	if result.Valid() {
		fmt.Printf("OK: %s is valid\n", docPath)
	} else {
		fmt.Printf("FAIL: %s is not valid. Errors:\n", docPath)
		for _, desc := range result.Errors() {
			fmt.Printf("  - %s\n", desc)
		}
	}
}