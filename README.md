# und - option, nullable and undefinedable types for JSON field.

Types to interoperate with applications that make full use of JSON. And a
specialized marshaler / unmarshaler for those types which skips `undefined`
fields in an input struct.

Currently it has 4 types.

- Nullable: `null` or `T`
- Undefinedable: `undefined` or `T`
  - A field that can be absent (undefined) but not a null.
  - Nullable and Undefinedable is 95% identical to each other. The only
    difference is that Undefinedable implements `IsUndefined` method which is
    used to determine the field should be skipped or not.
- JsonField: `null`, `undefined` or `T`
  - A default-ish type for fields of a JSON object.
- Elastic: `undefined | (null | T) | (null | T)[]`
  - An overly elastic type where it can be anything.
  - This can be used with the Elasticsearch JSON objects.
  - Or configuration files.

## Usage

```go
// or run this by calling go run ./example/main.go
package main

import (
	"fmt"

	"github.com/ngicks/und/v2/jsonfield"
	"github.com/ngicks/und/v2/nullable"
	"github.com/ngicks/und/v2/serde"
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
```

## Background

- Some applications / APIs uses full potential of JSON.
  - For example, Elasticsearch's update API uses
    [partial documents to update the documents partially.](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-update.html#_update_part_of_a_document)
    Setting `null` for fields overwrites that field to `null`.
- For Go, the programming language offers no comfortable way to achieve this.
  - Users can encode any structs into `map[string]any` and then can remove
    fields if those should be skipped / not updated.
  - This introduces cluttering to the codebase in some cases.

With help of generics which is
[introduced in Go 1.18](https://tip.golang.org/doc/go1.18#generics), we can do
this more effortlessly.

## Under the hood.

All those types -- `Nullable[T]`, `Undefinedable[T]`, `JsonField[T]`,
`Elastic[T]` -- are based on `Option[T]`.

The `Option[T]` is similar to the Rust's `Option<T>` but less methods (because
the Go has GC runtime and thus does not need to exposes no such concepts like
ownership to users).

It's just a comparable version of `*T`.

```go
type Option[T any] struct {
	some bool
	v    T
}
```

`Undefinedable[T]` and `Nullable[T]` is just wrappers around `Option`, just
adding named methods on it.

`JsonField[T]` and `Elastic[T]` combine them to represent 3 or 4 distinct states
with a single type.

To skip undefined fields when marshaling, all those type can be processed
through `serde` package, which is just an extension of
`github.com/json-iterator/go` and a config to which the extension is
pre-applied.

The extension, which is named `UndefinedSkipperExtension`, redirects the
emptiness evaluation to `IsUndefined` method implemented on each field type. And
it also fakes struct tags so that `IsUndefined` implementors always considered
tagged with `,omitempty`. This achieves mimicking the behavior where undefined
fields are skipped.
