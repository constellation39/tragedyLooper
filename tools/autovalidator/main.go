package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"
)

func main() {
	dataDir := "data"
	schemaDir := filepath.Join(dataDir, "schemas")

	fmt.Println("Starting validation...")
	if err := discoverAndValidate(dataDir, schemaDir); err != nil {
		fmt.Println("validation stopped with error:", err)
	}
	fmt.Println("Validation finished.")
}

// discoverAndValidate 遍历 dataDir（递归），
// • 如果是文件则直接校验；
// • 如果是目录则校验其内部的 *.json, *.yaml, *.yml；
// schema 的路径规则见上文说明。
// 注意：schemaDir 自身及其子目录会被完全跳过。
func discoverAndValidate(dataDir, schemaDir string) error {
	var validatedCount int
	schemaDirAbs, _ := filepath.Abs(schemaDir)

	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if insideSchemaDir(path, schemaDirAbs) {
			return handleSchemaDir(d)
		}

		if d.IsDir() {
			return processDirectory(path, schemaDirAbs, &validatedCount)
		}

		return processFile(path, d.Name(), schemaDirAbs, &validatedCount)
	}

	if err := filepath.WalkDir(dataDir, walkFn); err != nil {
		return err
	}

	if validatedCount == 0 {
		fmt.Println("No data files were found to validate.")
	}
	return nil
}

func handleSchemaDir(d fs.DirEntry) error {
	if d.IsDir() {
		return filepath.SkipDir
	}
	return nil
}

func processDirectory(path, schemaDirAbs string, validatedCount *int) error {
	dirName := filepath.Base(path)
	schemaPath := filepath.Join(schemaDirAbs, dirName+".json")
	if !fileExists(schemaPath) {
		return nil // No schema, no validation
	}

	entries, _ := os.ReadDir(path)
	for _, e := range entries {
		if e.IsDir() || !isSupportedDataFile(e.Name()) {
			continue
		}
		dataPath := filepath.Join(path, e.Name())
		validate(schemaPath, dataPath)
		*validatedCount++
	}
	return nil
}

func processFile(path, name, schemaDirAbs string, validatedCount *int) error {
	if isSupportedDataFile(name) {
		var schemaPath string
		if name == "basic_tragedy_x.yaml" {
			schemaPath = filepath.Join(schemaDirAbs, "ScriptConfig.json")
		} else {
			schemaPath = filepath.Join(schemaDirAbs, strings.TrimSuffix(name, filepath.Ext(name))+".json")
		}

		if fileExists(schemaPath) {
			validate(schemaPath, path)
			*validatedCount++
		}
	}
	return nil
}

func validate(schemaPath, dataPath string) {
	schemaAbs, _ := filepath.Abs(schemaPath)

	schemaLoader := gojsonschema.NewReferenceLoader("file://" + filepath.ToSlash(schemaAbs))

	var dataToValidate interface{}
	var err error
	if strings.HasSuffix(dataPath, ".yaml") || strings.HasSuffix(dataPath, ".yml") {
		dataToValidate, err = loadYamlFile(dataPath)
	} else {
		dataToValidate, err = loadJsonFile(dataPath)
	}
	if err != nil {
		fmt.Printf("Error loading data from %s: %v\n", dataPath, err)
		return
	}

	dataLoader := gojsonschema.NewGoLoader(dataToValidate)

	result, err := gojsonschema.Validate(schemaLoader, dataLoader)
	if err != nil {
		fmt.Printf("Error validating %s against %s: %v\n", dataPath, schemaPath, err)
		return
	}

	if result.Valid() {
		fmt.Printf("OK  : %s is valid\n", dataPath)
	} else {
		fmt.Printf("FAIL: %s is not valid. Errors:\n", dataPath)
		for _, desc := range result.Errors() {
			fmt.Printf("  - %s\n", desc)
		}
	}
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

func loadJsonFile(path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out interface{}
	err = json.Unmarshal(data, &out)
	return out, err
}

func isSupportedDataFile(name string) bool {
	return strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}

// insideSchemaDir 判断 path 是否位于 schemaDir（含其子目录）内部。
func insideSchemaDir(path, schemaDirAbs string) bool {
	abs, _ := filepath.Abs(path)
	if abs == schemaDirAbs {
		return true
	}
	rel, err := filepath.Rel(schemaDirAbs, abs)
	return err == nil && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != ".."
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
