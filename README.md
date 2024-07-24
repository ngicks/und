# und - option and undefined-able types, mainly for JSON fields.

Types to interoperate with applications that make full use of JSON.

## Before v1

und expects no breaking change except for `MarshalJSONV2` and `UnmarshalJSONV2` methods.

Normally those methods are not used directly by users, so this expected breakage should not incur visible effects for most of them.

- und waits for release of `encoding/json/v2`
  - und depends on `github.com/go-json-experiment/json` which is an experimental `encoding/json/v2` implementation.
  - Types defined in this module implement `json.MarshalerV2` and `json.UnmarshalerV2`.
  - Eventually the dependency would be swapped to `encoding/json/v2` and those methods would be changed to use `v2`,
  - or erased completely if `v2` decides not to continue to use that names.

## Example

run example by `go run github.com/ngicks/und/example@v1.0.0-alpha4`.

You'll see zero value fields whose type is defined under this module are omitted by jsonv2(`github.com/go-json-experiment/json`) with `omitzero` json option.

Also types defined under `sliceund` and `sliceund/elastic` are omitted by `encoding/json` v1 if zero, with `omitempty` struct tag option.

```go
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
	// see? undefined (=zero value) fields are omitted.
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
	// You see. Types defined under ./sliceund/ can be omitted by encoding/json.
	// Types defined in ./ and ./elastic cannot be omitted by it.
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
```

## being undefined is harder to express in Go.

- JSON, JavaScript Object Notation, includes concept of being absence(undefined), nil(null) or value(`T`).
- When unmarshaling incoming JSON bytes slice, you can use `*T` to decode value matching `T` and additionally converts undefined **OR** null to nil.
  - `nil` indicates the JSON field was `null` and thus nil was assigned to the field.
  - Or the field was nil and `json.Unmarshal` skipped modifying the field since no matching JSON field was present in the input.
- That's hard to express in Go, since:
  - You can always do that by encoding struct into `map[string]any` then remove fields that are supposed to be absent in an output JSON value.
  - Or define custom JSON marshaler that skips if fields are in some specific state.
  - If you do not want to do those,
    - You'll need a type that has 3 states.
    - You can do it by various ways, which include `**T`, `[]T`, `map[bool]T` and `struct {state int; value T}`.
      - `**T` is definitely not a choice since it is not allowed to be a method receiver as per specification.
      - `struct {state int; value T}` can not be skipped by v1 `encoding/json` since it does not check emptiness of struct.

As a conclusion, this package implements `struct {state int; value T}` and `[]T` kind types.

## types and variants

- `Option[T]`: Rust-like optional value.
  - can be Some or None.
  - is comparable if `T` is comparable.
  - have `Equal` method in case `T` is not comparable or comparable but needs custom equality tests(e.g. `time.Time`)
  - has convenient methods stolen from rust's `option<T>`
  - can be used in place of `*T`
  - is copied by assign.

Other types are based on `Option[T]`.

- `Und[T]`: undefined, null or `T`
- `Elastic[T]`: undefined, null, `T` or [](`T` | null)
  - mainly for consuming elasticsearch JSON documents.
  - or configuration files.

There are 2 variants

- `github.com/ngicks/und`: `Option[Option[T]]` based types.
  - omitted only if encoding through
    - `github.com/go-json-experiment/json`(possibly a future `encoding/json/v2`) with the `,omitzero` options
    - [jsoniter](https://github.com/json-iterator/go) with `,omitempty` and custom encoder.
- `github.com/ngicks/und/sliceund`: `[]Option[T]` based types.
  - omitted with `,omitempty`.
