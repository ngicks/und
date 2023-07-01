package main

import (
	"fmt"

	"github.com/ngicks/und/jsonfield"
	"github.com/ngicks/und/nullable"
	"github.com/ngicks/und/serde"
)

// define struct with UndefinedableField[T]
type Emm string

type Embedded struct {
	Foo string
	Bar nullable.Nullable[string]   `json:"bar"`
	Baz jsonfield.JsonField[string] `json:"baz"`
}

type sample struct {
	Emm // embedded non struct.
	Embedded
	Corge  string
	Grault nullable.Nullable[string]
	Garply jsonfield.JsonField[string]
}

func main() {
	v := sample{
		Emm: Emm("emm"),
		Embedded: Embedded{
			Foo: "aaa",
			Bar: nullable.Null[string](),
			Baz: jsonfield.Undefined[string](),
		},
		Corge:  "corge",
		Grault: nullable.NonNull("grault"),
		Garply: jsonfield.Undefined[string](),
	}

	bin, _ := serde.MarshalJSON(v)
	fmt.Println(string(bin))
	// This prints
	// {"Emm":"emm","Foo":"aaa","bar":null,"Corge":"corge","Grault":"grault"}
	// See Baz and Garply fields are skipped.
}
