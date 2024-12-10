# und - option and undefined-able types, mainly for JSON fields.

Types to interoperate with applications that make full use of JSON.

## Note: dependency of github.com/go-json-experiment/json will be dropped when Go 1.24 is released

`json:"omitzero"` will be added to Go at Go 1.24.
The dependency of github.com/go-json-experiment/json is no longer necessary since you can omit both the double-option und type and 
the slice-option und type just using `encoding/json` with `json:",omitzero"` option.

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

### Normal way to process JSON in Go.

When processing JSON values in GO, normally, at least I assume, you define a type that matches schema of JSON value, and use them with `encoding/json`.
(Of course there are numbers of third party modules that process JSON nicely, but I place them out of scope.)

I think you'll normally specify `*T` as field type to allow it to be empty. This treats undefined and null fields equally (unless you use non-zero value for an unmarshale target.)
This works fine in many cases. However sometimes its simplicity conflicts the concept of JSON.

### The difference of null | undefined sometimes does matter

As you can see in [Open API spec](https://github.com/OAI/OpenAPI-Specification),
JSON naturally has concept of absent fields(field is not specified in `required` section),
and also nullable field(`nullable` attribute is set to `true`)

**The difference of null and undefined does matter** in some common practice.
For example, [Elasticsearch](https://www.elastic.co/guide/en/elasticsearch/reference/current/elasticsearch-intro.html) allows users to send partial document JSON to [Update part of a document](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-update.html#_update_part_of_a_document). The Elasticsearch treats all of `undefined`(absent), `null`, `[]` equally; nonexistent field. So the partial update API skips updating of undefined fields and clears the fields if corresponding field of input JSON is `null`.

How do you achieve this partial update in Go?
I suspect simplest and most straightforward way is using `map[string]any` as a staging data.

### Unmarshaling T | null | undefined is easy

If a field type implements `json.Unmarshaler`, `encoding/json` calls this method while unamrashaling incoming JSON value only if there's matching field in the data, even when the value is `null` literal.
Therefore, `T | null | undefined` can be easily mapped from 3 state where: `UnmarshalJSON` was called with non-null data | was called with `null` literal | was not called, respectively.

### Marshaling T | null | undefined using stating map[string]any

The problem arises when marshaling the struct if the field has distinct `T | null | undefined` state; `encoding/json` does not omit struct. This is why you end up always specifying `*time.Time` as field type instead of `time.Time`, if you want to omit zero value of the time.

Most simplest way to omit zero struct, I think, is using `map[string]any` as staging data.

You can use [github.com/jeevatkm/go-model](https://github.com/jeevatkm/go-model) to map any arbitrary structs into `map[string]any`. Then you can remove any arbitrary field from that. (You can't use the popular [mapstructure](https://github.com/mitchellh/mapstructure) to achieve this because of [#334](https://github.com/mitchellh/mapstructure/issues/334)). Finally you can marshal `map[string]any` via `json.Marshal`.

This should incur unnecessary overhead to marshaling, also feels clumsier and tiring.

### Solution: use []Option[T] to achieve this

As you can see in here: https://github.com/golang/go/blob/go1.22.5/src/encoding/json/encode.go#L306-L318 ,
`omitempty` works on `[]T` and `map[K]V`.

With generics introduced in Go1.18, you can define a `[]T` based type with convenient methods. The only drawback is that you can't hide internal data structure for those type. Any change to that should be considered a breaking change. But it's ok because I suspect there's not a lot of chance of changing around it.

I've defined type like `type Option[T any]{valid bool; t T}` to express some or none. Then I combine this with `[]T` to express `T | null | undefined` so that I can limit length of slice to 1. The capacity for the slice always stays 1, allocated only once and no chance of growth afterwards, no excess memory consumption(only a single `bool` flag field).

As a conclusion, this module defines `[]Option[T]` based types to handle `T | null | undefined` easily.

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
