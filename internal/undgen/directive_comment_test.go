package undgen

import (
	"go/parser"
	"go/token"
	"testing"

	"gotest.tools/v3/assert"
)

const hello = `package main

import "fmt"

//undgen:ignore

//undgen:generated

// undgen:ignore

// undgen:generated


//undgen:ignore
//undgen:generated

/*
undgen:ignore
*/

/*
undgen:generated
*/
`

func TestDirective_Parse(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "hello.go", hello, parser.ParseComments|parser.AllErrors)
	if err != nil {
		panic(err)
	}

	for _, cg := range f.Comments {
		d, found, err := ParseUndComment(cg)
		assert.NilError(t, err)
		t.Logf("d: %#v, found: %t, err: %#v", d, found, err)
	}
}
