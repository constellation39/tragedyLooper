package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"
)

const (
	dataDir      = "data"
	schemaDir    = "data/schemas"
	scriptsDir   = "data/scripts"
	scriptSchema = "ScriptConfig.json"
)

func main() {
	fmt.Println("Starting validation...")
	if err := discoverAndValidate(); err != nil {
		fmt.Printf("Validation stopped with error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Validation finished.")
}

func discoverAndValidate() error {
	schemaDirAbs, _ := filepath.Abs(schemaDir)

	fmt.Println("--- Validating Script Packages ---")
	scriptDirs, err := os.ReadDir(scriptsDir)
	if err != nil {
		return fmt.Errorf("could not read scripts directory '%s': %w", scriptsDir, err)
	}

	schemaPath := filepath.Join(schemaDirAbs, scriptSchema)
	if !fileExists(schemaPath) {
		return fmt.Errorf("script schema not found at %s", schemaPath)
	}

	validatedCount := 0
	for _, entry := range scriptDirs {
		if entry.IsDir() {
			dirPath := filepath.Join(scriptsDir, entry.Name())
			fmt.Printf("Processing script package: %s\n", dirPath)

			assembledData, err := validateScriptPackage(dirPath, schemaPath)
			if err != nil {
				fmt.Printf("FAIL: Validation failed for package %s: %v\n", entry.Name(), err)
			} else {
				outputJSONPath := filepath.Join(scriptsDir, entry.Name()+".json")
				err = writeAssembledDataAsJSON(assembledData, outputJSONPath)
				if err != nil {
					fmt.Printf("WARN: Could not write validated json for package %s: %v\n", entry.Name(), err)
				} else {
					fmt.Printf("OK  : Package %s is valid and JSON written to %s\n", entry.Name(), outputJSONPath)
					validatedCount++
				}
			}
		}
	}

	if validatedCount == 0 {
		fmt.Println("No script packages were found to validate.")
	}
	return nil
}

func validateScriptPackage(dirPath, schemaPath string) (map[string]interface{}, error) {
	assembledData, err := assembleScriptData(dirPath)
	if err != nil {
		return nil, fmt.Errorf("could not assemble data: %w", err)
	}

	schemaLoader := gojsonschema.NewReferenceLoader("file://" + filepath.ToSlash(schemaPath))
	dataLoader := gojsonschema.NewGoLoader(assembledData)

	result, err := gojsonschema.Validate(schemaLoader, dataLoader)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	if !result.Valid() {
		var errs []string
		for _, desc := range result.Errors() {
			errs = append(errs, fmt.Sprintf("  - %s", desc))
		}
		return nil, fmt.Errorf("script is not valid. Errors:\n%s", strings.Join(errs, "\n"))
	}

	return assembledData, nil
}

func assembleScriptData(dirPath string) (map[string]interface{}, error) {
	assembled := make(map[string]interface{})
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		filePath := filepath.Join(dirPath, file.Name())
		content, err := loadYamlFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", filePath, err)
		}

		key := strings.TrimSuffix(file.Name(), ".yaml")
		assembled[key] = content
	}

	baseConfigPath := dirPath + ".yaml"
	if fileExists(baseConfigPath) {
		baseContent, err := loadYamlFile(baseConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load base config %s: %w", baseConfigPath, err)
		}
		if baseMap, ok := baseContent.(map[string]interface{}); ok {
			for k, v := range baseMap {
				if _, exists := assembled[k]; !exists {
					assembled[k] = v
				}
			}
		}
	}

	return assembled, nil
}

func writeAssembledDataAsJSON(data map[string]interface{}, path string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}
	return os.WriteFile(path, jsonData, 0644)
}

func loadYamlFile(path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out interface{}
	err = yaml.Unmarshal(data, &out)
	if err != nil {
		return nil, err
	}
	return convertMapKeysToStrings(out), nil
}

func convertMapKeysToStrings(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v := range x {
			m[fmt.Sprintf("%v", k)] = convertMapKeysToStrings(v)
		}
		return m
	case []interface{}:
		for i, v := range x {
			x[i] = convertMapKeysToStrings(v)
		}
	}
	return i
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}