package undgen

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
)

func clone(fset *token.FileSet, f *ast.File) (*ast.File, *token.FileSet) {
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, f) // stupid deep clone
	pos := fset.Position(f.FileStart)

	fset = token.NewFileSet()
	f, err := parser.ParseFile(fset, pos.Filename, buf.String(), parser.ParseComments)
	if err != nil {
		panic(err)
	}
	return f, fset
}
