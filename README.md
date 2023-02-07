# undefinedablejson

json fields that can be undefined or null or T. And the marshaller for the struct type containing them.

## Usage

```go
// define struct with UndefinedableField[T]
type Emm string

type Embedded struct {
	Foo string
	Bar undefinedablejson.Nullable[string]      `json:"bar"`
	Baz undefinedablejson.Undefinedable[string] `json:"baz"`
}

type sample struct {
	Emm // embedded non struct.
	Embedded
	Corge  string
	Grault undefinedablejson.Nullable[string]
	Garply undefinedablejson.Undefinedable[string]
}

func main() {
	v := sample{
    Emm: Emm("emm"),
    Embedded: Embedded{
      Foo: "aaa",
      Bar: undefinedablejson.Null[string](),
      Baz: undefinedablejson.UndefinedField[string](),
    },
    Corge:  "corge",
    Grault: undefinedablejson.NonNull("grault"),
    Garply: undefinedablejson.UndefinedField[string](),
  }

  // Marshal with MarshalFieldsJSON,
  // Unmarshal with UnmarshalFieldsJSON.
  bin, _ := undefinedablejson.MarshalFieldsJSON(v)

  // string(bin) is
  // {"Emm":"emm","Foo":"aaa","bar":null,"Corge":"corge","Grault":"grault"}
  // See it skips undefined Undefinedable[T]
}
```

## Supported tags

- json:"name"
  - json:"-" skips the field
  - json:"-," tags the field as `-`.
- json:",omitempty"
- json:",string"
  - It also works on `Nullable[T]` and `Undefinedable[T]`.
  - If you don't like linters warn about incorrect json:",string" usage, use und:"string" instead.

## Background

- Some APIs are aware of `undefined | null | T`.
  - For example, Elasticsearch's update API can use [partial documents to update the documents partially.](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-update.html#_update_part_of_a_document) Setting `null` for fields overwrites that field to be `null`.
- AFAIK most of programming languages do not natively have 2 types to express `being empty`. That's the JavaScript's good or odd point.
  - Namely `undefined` and `null`
- Go also does not have 2 types to state `being empty`.
  - Go uses `*T` for `being empty`. `nil` is empty of course.
- If you need to determine, what field and whether you should skip or set `null` to, at runtime, you need an additional data structure for that.

## How is it implemented

Above 1.18 or later, Go has generics.

With help of type parameters, the Nullable is simply

```go
type Option[T any] struct {
	some bool
	v    T
}

type Nullable[T any] struct {
	Option[T]
}
```

If some is false, it is null.

The Undefinedable is wraps Nullable:

```go
type Undefinedable[T any] struct {
	Option[Option[T]]
}
```

If some is false, it is undefined. If v.some is false, it is null.

MarshalFieldsJSON and UnmarshalFieldsJSON are far-less careful version of json.Marshal / json.Unmarshal.
Those rely on `reflect` package. First they read through information of given type, cache result, and then marshal/unmarhsal given type.

## TODO

- [ ] add support for `any`
- [ ] add `und:"disallowNull"`, `und:"disallowUndefined"`, `und:"required"`
- [ ] add code generator to avoid `reflect` usage.
- [ ] add more tests.
- [ ] find other packages doing same thing, and implemented more elegantly.
  - Once found, archive this package and support that package.
