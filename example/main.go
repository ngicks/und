package main

import (
	"fmt"

	"github.com/ngicks/und/nullable"
	"github.com/ngicks/und/serde"
	"github.com/ngicks/und/undefinedable"
)

// define struct with UndefinedableField[T]
type Emm string

type Embedded struct {
	Foo string
	Bar nullable.Nullable[string]           `json:"bar"`
	Baz undefinedable.Undefinedable[string] `json:"baz"`
}

type sample struct {
	Emm // embedded non struct.
	Embedded
	Corge  string
	Grault nullable.Nullable[string]
	Garply undefinedable.Undefinedable[string]
}

func main() {
	v := sample{
		Emm: Emm("emm"),
		Embedded: Embedded{
			Foo: "aaa",
			Bar: nullable.Null[string](),
			Baz: undefinedable.Undefined[string](),
		},
		Corge:  "corge",
		Grault: nullable.NonNull("grault"),
		Garply: undefinedable.Undefined[string](),
	}

	bin, _ := serde.MarshalJSON(v)
	fmt.Println(string(bin))
	// This prints
	// {"Emm":"emm","Foo":"aaa","bar":null,"Corge":"corge","Grault":"grault"}
	// See Baz and Garply fields are skipped.
}
