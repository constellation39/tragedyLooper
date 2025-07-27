package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

func main() {
	dataDir := "data"
	schemaDir := filepath.Join(dataDir, "jsonschema")

	fmt.Println("Starting validation...")
	if err := discoverAndValidate(dataDir, schemaDir); err != nil {
		fmt.Println("validation stopped with error:", err)
	}
	fmt.Println("Validation finished.")
}

// discoverAndValidate 遍历 dataDir（递归），
// • 如果是文件则直接校验；
// • 如果是目录则校验其内部的 *.json；
// schema 的路径规则见上文说明。
// 注意：schemaDir 自身及其子目录会被完全跳过。
func discoverAndValidate(dataDir, schemaDir string) error {
	var validatedCount int

	// 将 schemaDir 归一化，方便后面比较
	schemaDirAbs, _ := filepath.Abs(schemaDir)

	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 忽略 schemaDir 里的任何东西
		if insideSchemaDir(path, schemaDirAbs) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 目录 —— 用目录名+".json" 找 schema
		if d.IsDir() {
			schemaPath := filepath.Join(schemaDirAbs, d.Name()+".json")
			if !fileExists(schemaPath) {
				return nil // 没有 schema 就不校验
			}
			entries, _ := os.ReadDir(path)
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
					continue
				}
				dataPath := filepath.Join(path, e.Name())
				validateFile(schemaPath, dataPath)
				validatedCount++
			}
			return nil
		}

		// 单个文件
		if strings.HasSuffix(d.Name(), ".json") {
			schemaPath := filepath.Join(schemaDirAbs, d.Name())
			if fileExists(schemaPath) {
				validateFile(schemaPath, path)
				validatedCount++
			}
		}
		return nil
	}

	if err := filepath.WalkDir(dataDir, walkFn); err != nil {
		return err
	}

	if validatedCount == 0 {
		fmt.Println("No data files were found to validate.")
	}
	return nil
}

func validateFile(schemaPath, docPath string) {
	schemaAbs, _ := filepath.Abs(schemaPath)
	docAbs, _ := filepath.Abs(docPath)

	schemaLoader := gojsonschema.NewReferenceLoader("file://" + filepath.ToSlash(schemaAbs))
	docLoader := gojsonschema.NewReferenceLoader("file://" + filepath.ToSlash(docAbs))

	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		fmt.Printf("Error validating %s against %s: %v\n", docPath, schemaPath, err)
		return
	}

	if result.Valid() {
		fmt.Printf("OK  : %s is valid\n", docPath)
	} else {
		fmt.Printf("FAIL: %s is not valid. Errors:\n", docPath)
		for _, desc := range result.Errors() {
			fmt.Printf("  - %s\n", desc)
		}
	}
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
