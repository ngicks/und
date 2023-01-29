# undefinedablejson

json fields that can be undefined or null or T.

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

## Rationale

- Go has no way to express `undefined` field, which is often useful in javascript.
- JSON PATCHing in RESTful API is easier when data payload can express explicit empty value, `null` namely.
  - Implementations can clear the field if input json has `null` for corresponding fields.
- Go has `omitempty` option that skip the field when marshalling.
- Go's std `encoding/json` package marshal `nil` pointer to `null`, and vice versa.
- User codes can utilize \*\*T to define 3 states, `undefined` or `null` or T.
  - But it is harder to write: It end up clutter codes like `num := 3; numP := &num; {Field: &numP}`. You always need to do this.
- Helper types can make it easier.
- There's no way to skip field other than `omitmepty` tag, in std `encoding/json`.
  - Unfortunately, `omitempty` does not skip struct type.
- So user code must pay some effort on it.

## How is it implemented

The Nullable is simply

```go
type Nullable[T any] struct {
	v *T
}
```

if v is nil, it is null.

The Undefinedable is wraps Nullable:

```go
type Undefinedable[T any] struct {
	v *Nullable[T]
}
```

if v is nil, it is undefined, if v.v is nil, it is null.

MarshalFieldsJSON and UnmarshalFieldsJSON are far-less careful version of json.Marshal / json.Unmarshal.
Those rely on `reflect` package. First they read through information of given type, cache result, and then marshal/unmarhsal given type.

## TODO

- [ ] add support for `any`
- [ ] add `und:"disallowNull"`, `und:"disallowUndefined"`, `und:"required"`
- [ ] add code generator to avoid `reflect` usage.
- [ ] add more tests.
- [ ] find other packages doing same thing, and implemented more elegantly.
  - Once found, archive this package and support that package.
