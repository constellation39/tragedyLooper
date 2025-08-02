
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

// phaseNode represents a node in the flowchart graph.
type phaseNode struct {
	Name        string
	Transitions []transition
}

// transition represents a directed edge from one phaseNode to another.
type transition struct {
	TargetPhase string
	Condition   string
}

// main is the entry point of the flowchart generator.
func main() {
	// Path to the source file containing the flowchart definition.
	sourceFilePath := "internal/game/engine/phasehandler/flowchart.go"
	// Path for the output DOT file.
	outputFilePath := "flowchart.dot"

	// Parse the source file to build an abstract syntax tree (AST).
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, sourceFilePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Failed to parse source file '%s': %v", sourceFilePath, err)
	}

	// Find the Flowchart variable declaration in the AST.
	flowchartExpr := findFlowchartDeclaration(node)
	if flowchartExpr == nil {
		log.Fatal("Could not find 'Flowchart' variable declaration in the AST.")
	}

	// Parse the AST expression to extract the flowchart data.
	nodes := parseFlowchart(flowchartExpr)

	// Generate the DOT file content from the parsed data.
	dotContent := generateDotFile(nodes)

	// Write the generated content to the output file.
	if err := os.WriteFile(outputFilePath, []byte(dotContent), 0644); err != nil {
		log.Fatalf("Failed to write to output file '%s': %v", outputFilePath, err)
	}

	fmt.Printf("✅ Flowchart saved to %s\n", outputFilePath)
}

// findFlowchartDeclaration searches the AST for the 'Flowchart' variable.
func findFlowchartDeclaration(file *ast.File) ast.Expr {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}
		for _, spec := range genDecl.Specs {
			valSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			if valSpec.Names[0].Name == "Flowchart" {
				return valSpec.Values[0]
			}
		}
	}
	return nil
}

// parseFlowchart parses the AST expression of the flowchart and returns a map of phase nodes.
func parseFlowchart(expr ast.Expr) map[string]*phaseNode {
	nodes := make(map[string]*phaseNode)
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nodes
	}

	for _, elt := range lit.Elts {
		kvExpr, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key, keyOk := kvExpr.Key.(*ast.SelectorExpr)
		fromPhase := "UNSPECIFIED"
		if keyOk {
			fromPhase = cleanPhaseName(key.Sel.Name)
		}

		if _, exists := nodes[fromPhase]; !exists {
			nodes[fromPhase] = &phaseNode{Name: fromPhase}
		}

		transitionsList, ok := kvExpr.Value.(*ast.CompositeLit)
		if !ok {
			continue
		}

		for _, transElt := range transitionsList.Elts {
			compLit, ok := transElt.(*ast.CompositeLit)
			if !ok {
				continue
			}

			var newTransition transition
			for _, field := range compLit.Elts {
				fieldKv, ok := field.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				fieldName := fieldKv.Key.(*ast.Ident).Name
				switch fieldName {
				case "Next":
					if selExpr, ok := fieldKv.Value.(*ast.SelectorExpr); ok {
						newTransition.TargetPhase = cleanPhaseName(selExpr.Sel.Name)
					}
				case "Condition":
					// A basic attempt to stringify the condition.
					// This part might need to be more robust for complex conditions.
					newTransition.Condition = "Conditional" // Simplified label
				}
			}
			nodes[fromPhase].Transitions = append(nodes[fromPhase].Transitions, newTransition)
		}
	}
	return nodes
}

// cleanPhaseName removes the "GAME_PHASE_" prefix for brevity in the graph.
func cleanPhaseName(name string) string {
	return strings.TrimPrefix(name, "GAME_PHASE_")
}

// generateDotFile creates the content of a .dot file for Graphviz.
func generateDotFile(nodes map[string]*phaseNode) string {
	var buf bytes.Buffer
	buf.WriteString("digraph TragedyLooperFlowchart {\n")
	buf.WriteString("    rankdir=TB;\n")
	buf.WriteString("    node [shape=box, style=\"rounded,filled\", fillcolor=lightblue];\n\n")

	for name, node := range nodes {
		for _, trans := range node.Transitions {
			label := ""
			if trans.Condition != "" {
				label = ` [label="?", style=dashed]` // Using "?" as a placeholder for condition
			}
			buf.WriteString(fmt.Sprintf(`    "%s" -> "%s"%s;`+"\n", name, trans.TargetPhase, label))
		}
	}

	buf.WriteString("}\n")
	return buf.String()
}
