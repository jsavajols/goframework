package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

func main() {
	root := "./" // Répertoire du projet
	fileSet := token.NewFileSet()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".go" {
			node, err := parser.ParseFile(fileSet, path, nil, parser.AllErrors)
			if err != nil {
				return err
			}
			// Parcourir les déclarations pour trouver les fonctions
			for _, decl := range node.Decls {
				if funcDecl, ok := decl.(*ast.FuncDecl); ok {
					fmt.Printf("Function: %s (File: %s)\n", funcDecl.Name.Name, path)
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error:", err)
	}
}
