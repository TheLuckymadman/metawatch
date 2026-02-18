package main

import (
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "test.go", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	if err = ast.Print(fs, f); err != nil {
        panic(err)
    }
	
}