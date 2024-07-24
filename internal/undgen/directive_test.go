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

//undgen:{"json":"message"}

//undgen:{
// "json": "that",
// "spans": "multiple lines"
// }

// in the middle
//undgen:ignore
// of the comment

// after the json
// undgen:{
//
//   "foo": "bar"
// } aaa
// bbb

/*
undgen:{
	"foo":"bar"
*/
/*
}
*/

// just a simple commnet

func main() {
        fmt.Println("Hello, world")
}`

func TestDirective_Parse(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "hello.go", hello, parser.ParseComments|parser.AllErrors)
	if err != nil {
		panic(err)
	}

	for _, cg := range f.Comments {
		d, found, err := ParseComment(cg)
		assert.NilError(t, err)
		t.Logf("d: %#v, found: %t, err: %#v", d, found, err)
	}
}
