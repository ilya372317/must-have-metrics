package mainexit

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "mainexit",
	Doc:  "Analyzer checks for the presence of os.Exit() in main function of main package.",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}
		filename := pass.Fset.Position(file.Pos()).Filename
		if !strings.HasSuffix(filename, ".go") {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			fun, ok := n.(*ast.FuncDecl)
			if ok {
				if fun.Name.Name == "main" {
					ast.Inspect(fun.Body, func(n ast.Node) bool {
						if call, ok := n.(*ast.CallExpr); ok {
							if isOsExitCall(call) {
								pass.Reportf(call.Pos(), "calling Exit function of os package not recomended")
							}
						}

						return true
					})
				}
			}

			return true
		})
	}

	return nil, nil
}

func isOsExitCall(call *ast.CallExpr) bool {
	if selectorIdent, ok := call.Fun.(*ast.SelectorExpr); ok {
		if parentIdent, ok := selectorIdent.X.(*ast.Ident); ok {
			if parentIdent.Name == "os" && selectorIdent.Sel.Name == "Exit" {
				return true
			}
		}
	}

	return false
}
