// Package staticlint performs code checking for direct os.Exit calls
package staticlint

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "noosexitmain",
	Doc:  "reports calls to os.Exit inside main functions",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}
		for _, decl := range file.Decls {
			if f, ok := decl.(*ast.FuncDecl); ok {
				if f.Name.Name != "main" {
					continue
				}
				ast.Inspect(f.Body, func(n ast.Node) bool {
					call, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}
					sel, ok := call.Fun.(*ast.SelectorExpr)
					if !ok {
						return true
					}
					ident, ok := sel.X.(*ast.Ident)
					if !ok {
						return true
					}
					if ident.Name == "os" && sel.Sel.Name == "Exit" {
						pass.Reportf(call.Pos(),
							"direct call to os.Exit in main is prohibited")
					}
					return true
				})
			}
		}
	}
	return nil, nil
}
