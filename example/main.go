package main

import (
	"encoding/json"
	"fmt"

	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"

	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
)

type sample1 struct {
	Foo  string
	Bar  und.Und[nested1]              `json:",omitzero"`
	Baz  elastic.Elastic[nested1]      `json:",omitzero"`
	Qux  sliceund.Und[nested1]         `json:",omitzero"`
	Quux sliceelastic.Elastic[nested1] `json:",omitzero"`
}

type nested1 struct {
	Bar  und.Und[string]            `json:",omitzero"`
	Baz  elastic.Elastic[int]       `json:",omitzero"`
	Qux  sliceund.Und[float64]      `json:",omitzero"`
	Quux sliceelastic.Elastic[bool] `json:",omitzero"`
}

type sample2 struct {
	Foo  string
	Bar  und.Und[nested2]              `json:",omitempty"`
	Baz  elastic.Elastic[nested2]      `json:",omitempty"`
	Qux  sliceund.Und[nested2]         `json:",omitempty"`
	Quux sliceelastic.Elastic[nested2] `json:",omitempty"`
}

type nested2 struct {
	Bar  und.Und[string]            `json:",omitempty"`
	Baz  elastic.Elastic[int]       `json:",omitempty"`
	Qux  sliceund.Und[float64]      `json:",omitempty"`
	Quux sliceelastic.Elastic[bool] `json:",omitempty"`
}

func main() {
	s1 := sample1{
		Foo:  "foo",
		Bar:  und.Defined(nested1{Bar: und.Defined("foo")}),
		Baz:  elastic.FromValue(nested1{Baz: elastic.FromOptions(option.Some(5), option.None[int](), option.Some(67))}),
		Qux:  sliceund.Defined(nested1{Qux: sliceund.Defined(float64(1.223))}),
		Quux: sliceelastic.FromValue(nested1{Quux: sliceelastic.FromOptions(option.None[bool](), option.Some(true), option.Some(false))}),
	}

	var (
		bin []byte
		err error
	)
	bin, err = json.MarshalIndent(s1, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("marshaled by with omitzero =\n%s\n", bin)
	// see? undefined (=zero value) fields are omitted with json:",omitzero" option.
	// ,omitzero is introduced in Go 1.24. For earlier version Go, see example of sample2 below.
	/*
		marshaled by with omitzero =
		{
		    "Foo": "foo",
		    "Bar": {
		        "Bar": "foo"
		    },
		    "Baz": [
		        {
		            "Baz": [
		                5,
		                null,
		                67
		            ]
		        }
		    ],
		    "Qux": {
		        "Qux": 1.223
		    },
		    "Quux": [
		        {
		            "Quux": [
		                null,
		                true,
		                false
		            ]
		        }
		    ]
		}
	*/

	s2 := sample2{
		Foo:  "foo",
		Bar:  und.Defined(nested2{Bar: und.Defined("foo")}),
		Baz:  elastic.FromValue(nested2{Baz: elastic.FromOptions(option.Some(5), option.None[int](), option.Some(67))}),
		Qux:  sliceund.Defined(nested2{Qux: sliceund.Defined(float64(1.223))}),
		Quux: sliceelastic.FromValue(nested2{Quux: sliceelastic.FromOptions(option.None[bool](), option.Some(true), option.Some(false))}),
	}

	bin, err = json.MarshalIndent(s2, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("marshaled with omitempty =\n%s\n", bin)
	// You see. Types defined under ./sliceund/ can be omitted by encoding/json@go1.23 or earlier.
	/*
		marshaled with omitempty =
		{
		    "Foo": "foo",
		    "Bar": {
		        "Bar": "foo",
		        "Baz": null
		    },
		    "Baz": [
		        {
		            "Bar": null,
		            "Baz": [
		                5,
		                null,
		                67
		            ]
		        }
		    ],
		    "Qux": {
		        "Bar": null,
		        "Baz": null,
		        "Qux": 1.223
		    },
		    "Quux": [
		        {
		            "Bar": null,
		            "Baz": null,
		            "Quux": [
		                null,
		                true,
		                false
		            ]
		        }
		    ]
		}
	*/
}
