package main

import (
	"encoding/json"
	"fmt"

	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
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
		Baz:  elastic.FromValue(nested1{Baz: elastic.FromOptions([]option.Option[int]{option.Some(5), option.None[int](), option.Some(67)})}),
		Qux:  sliceund.Defined(nested1{Qux: sliceund.Defined(float64(1.223))}),
		Quux: sliceelastic.FromValue(nested1{Quux: sliceelastic.FromOptions([]option.Option[bool]{option.None[bool](), option.Some(true), option.Some(false)})}),
	}

	var (
		bin []byte
		err error
	)
	bin, err = jsonv2.Marshal(s1, jsontext.WithIndent("    "))
	if err != nil {
		panic(err)
	}
	fmt.Printf("marshaled by v2=\n%s\n", bin)
	// see? undefined (=zero value) fields are skipped.
	/*
	   marshaled by v2=
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
		Baz:  elastic.FromValue(nested2{Baz: elastic.FromOptions([]option.Option[int]{option.Some(5), option.None[int](), option.Some(67)})}),
		Qux:  sliceund.Defined(nested2{Qux: sliceund.Defined(float64(1.223))}),
		Quux: sliceelastic.FromValue(nested2{Quux: sliceelastic.FromOptions([]option.Option[bool]{option.None[bool](), option.Some(true), option.Some(false)})}),
	}

	bin, err = json.MarshalIndent(s2, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("marshaled by v1=\n%s\n", bin)
	// You see. Types defined under ./sliceund/ can be skipped by encoding/json.
	// Types defined in ./ and ./elastic cannot be skipped by it.
	/*
	   marshaled by v1=
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
