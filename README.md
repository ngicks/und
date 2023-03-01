# und - option, nullable and undefinedable types for JSON field.

Types that can be `undefined` or `null` or `T`. And a marshaller implementation for struct types containing them which skips undefined `Undefinable[T]`.

## Usage

```go
// or run this by calling go run ./example/main.go
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
```

## Background

- Some APIs are aware of `undefined | null | T`.
  - For example, Elasticsearch's update API can use [partial documents to update the documents partially.](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-update.html#_update_part_of_a_document) Setting `null` for fields overwrites that field to `null`.
- AFAIK most of programming languages do not natively have 2 types to express `being empty`.
  - Namely `undefined` and `null`
  - That's the JavaScript's good or odd point.
- Go also does not have 2 types to state `being empty`.
  - User code usually uses `*T` for `being empty`. `nil` is empty of course.
- If you need to determine, what field and whether you should skip or set `null` to, at runtime, you need an additional data structure for that.
  - As far as I observed, user codes can use `map[string]any`.

## How it is implemented

With help of generics which is added in Go 1.18, we can define `Null[T]`, `Option[T]` or whatever.

`Undefinedable[T]` simply is `Option[Nullable[T]]`.

Now we only need to skip undefined fields when marshalling struct types.

As you already know, json has `,omitempty` option. However unfortunately, [it won't skip struct type fields](https://cs.opensource.google/go/go/+/refs/tags/go1.20.0:src/encoding/json/encode.go;drc=d5de62df152baf4de6e9fe81933319b86fd95ae4;l=339).

While std does not allow us to determine emptiness of fields by their value, [github.com/json-iterator/go](https://github.com/json-iterator/go), an excellent json serializer/deserializer library that is almost 100% compatible with std interface, exposes IsEmpty function.
It also allow us to fake struct tags set to fields.

`serde.MarshalJSON` uses `UndefinedableExtension` to swap IsEmpty to IsUndefined, and fake struct tag so that those fields look like always tagged with `,omitmepty`.
