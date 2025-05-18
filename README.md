# und - option and undefined-able types, mainly for JSON fields. [![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/ngicks/und)

Types to interoperate with applications that make full use of JSON.

## Notice: minimum Go version will be Go 1.25

As of [#71497](https://github.com/golang/go/issues/71497), `encoding/json/v2` would be added to `Go 1.25` under `GOEXPERIMENT=jsonv2` constraint if nothing goes wrong.
And also it might added by default at `Go 1.26`.

**Go version will be Go 1.25 at least. If any significant API change DOES occur, it will be Go 1.26**

Slice variants will also be supported for some code that might still need to stick to `v1`.

## Express T | null | undefined by only using the types in struct field.

Just use `und.Und` (for Go 1.24 or later) or `sliceund.Und` (for Go 1.23 or earlier) as struct field type then place `,omitzero`, `,omitempty` respectively.

```go
type sample struct {
    Foo und.Und[string]      `json:",omitzero"`
    Bar sliceund.Und[string] `json:",omitempty"`
}
```

The zero value is _undefined_, both v1 nad v2 of `encoding/json` skips fields.

```go
s := sample{}

bin, _ := json.MarshalIndent(s, "", "    ")
fmt.Printf("zero = %s\n", bin)
/*
    zero = {}
*/
```

Use `Null` functions to create _null_ objects.

```go
s.Foo = und.Null[string]()
s.Bar = sliceund.Null[string]()

bin, _ = json.MarshalIndent(s, "", "    ")
fmt.Printf("null = %s\n", bin)
/*
    null = {
        "Foo": null,
        "Bar": null
    }
*/
```

Use `Defined` functions to create _defined_ objects.

```go
s.Foo = und.Defined("foo")
s.Bar = sliceund.Defined("bar")

bin, _ = json.MarshalIndent(s, "", "    ")
fmt.Printf("defined = %s\n", bin)
/*
    defined = {
        "Foo": "foo",
        "Bar": "bar"
    }
*/
```

Use `Undefined` functions to create _undefined_ objects.

```go
s.Foo = und.Undefined[string]()
s.Bar = sliceund.Undefined[string]()

bin, _ = json.MarshalIndent(s, "", "    ")
fmt.Printf("undefined = %s\n", bin)
/*
    undefined = {}
*/
```

## types and variants

- `Option[T]`: Rust-like optional value.
  - can be Some or None.
  - zero is None.
  - is comparable if `T` is comparable.
  - have `EqualFunc[T](t Option[T], cmp func(i, j T) bool)` method to test equality of 2 options.
    - `cmp` should return value in the same manner as `cmp.Compare` does.
    - use `option.Equal` if options can simply be compared by equality `a == b`.
    - use `option.EqualEqualer` if options can be compared by calling `Equal` method on type `T`.
  - has convenient methods stolen from rust's `core::option::Option<T>`
  - can be used in some (not all) place of `*T`
  - is copied by assign.

Other types are based on `Option[T]`.

- `Und[T]`: _undefined_ (_empty_ or _unspecified_), _null_ or `T` (any type you like)
- `Elastic[T]`: _undefined_ (_empty_ or _unspecified_), _null_, `T` or [](`T` | null)
  - mainly for consuming elasticsearch JSON documents.
  - or maybe useful for user hand written configuration files.

There are 2 variants

- `github.com/ngicks/und`: `Option[Option[T]]` based types.
  - most light-weighted.
  - comparable if `T` is comparable.
  - omitted with `,omitzero` for Go 1.24 or later version.
- `github.com/ngicks/und/sliceund`: `[]Option[T]` based types.
  - omitted with `,omitempty`.
  - For Go 1.23 or earlier version.

## Example

run example in 2 versions, go 1.24.0 (or later) and go 1.23.0.

```
GOTOOLCHAIN=go1.24.0 go run github.com/ngicks/und/example@v1.0.0-alpha9
```

```
GOTOOLCHAIN=go1.23.0 go run github.com/ngicks/und/example@v1.0.0-alpha9
```

see that `sliceund` and `sliceund/elastic` can be omitted even in go1.23.0.

As you can see, types defined in ./ (package `und`) and ./elastic (package `elastic`) can be omitted with `json:",omitzero"` option for Go 1.24 or later version.

For Go 1.23 or earlier version of it, you can also use types under ./sliceund (package `sliceund`) or ./sliceund/elastic (package `elastic`) with `json:",omitempty"` option.

```go
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
```

## generate Patcher, Validator, Plain types with github.com/ngicks/go-codegen/codegen

`github.com/ngicks/go-codegen/codegen` has the `undgen` sub command which generates methods to, types from the types that contains any und types(`option.Option[T]`, `und.Und[T]`, `elastic.Elastic[T]`, `sliceund.Und[T]` and `sliceund/elastic.Elastic[T]`).

```
go run github.com/ngicks/go-codegen/codegen undgen patch     -v --dir /path/to/root/dir/of/target/package --pkg ./path/to/package ...
go run github.com/ngicks/go-codegen/codegen undgen validator -v --dir /path/to/root/dir/of/target/package --pkg ./...
go run github.com/ngicks/go-codegen/codegen undgen plain     -v --dir /path/to/root/dir/of/target/package --pkg ./...
```

- The patch sub-sub commands generates patcher for any struct types.
  - It takes any struct types, then generates the type whose field is same as target's but the type is wrapped in `sliceund.Und` and `json:",omitempty"` added.
  - The generated patch type can be unmarshaled from partial JSON then can be used to patch(partially overwrite fields) the target struct.
- The validator sub-sub commands generates validator method for any types containing any of und types.
  - The method only validates und state of the und fields.
  - It validates according to `und:""` struct tag.
- The plain sub-sub commands generates _plain_ types and interconversion methods on types.
  - It takes any types containing any of und fields, then generates _plain_ type whose fields is same as target's but the type is unwrapped according to `und:""` struct tag.

Notable flags:

- `-v` : verbose logs.
- `--dir`: specify directory under which the target packages are placed.
- `--pkg`: same package pattern that can be passed to `go list`. must be prefixed with `./`. `patch` sub command only accept pattern that matches only a single package.
- `types...`: the `patch` sub command needs `types...` arguments to specify target type names. Use `...` to target all types found under `--pkg`.

Examples below assumes `example.go` is placed under `./pkg/example` and it contains types described.

### patch command

```
go run github.com/ngicks/go-codegen/codegen undgen patch --dir ./pkg/example --pkg ./ ...
```

```go
// example.go
type PatchExample struct {
	Foo string
	Bar *int     `json:",omitempty"`
	Baz []string `json:"baz,omitempty"`
}
```

This emits the type and associated methods.
Output filenames are name of the file in which the target type defined but with suffix `.und_patch`.

```go
// example.und_patch.go

//codegen:generated
type PatchExamplePatch struct {
	Foo sliceund.Und[string]   `json:",omitempty"`
	Bar sliceund.Und[*int]     `json:",omitempty"`
	Baz sliceund.Und[[]string] `json:"baz,omitempty"`
}

//codegen:generated
func (p *PatchExamplePatch) FromValue(v PatchExample) {
	//nolint
	*p = PatchExamplePatch{
		Foo: sliceund.Defined(v.Foo),
		Bar: sliceund.Defined(v.Bar),
		Baz: sliceund.Defined(v.Baz),
	}
}

//codegen:generated
func (p PatchExamplePatch) ToValue() PatchExample {
	//nolint
	return PatchExample{
		Foo: p.Foo.Value(),
		Bar: p.Bar.Value(),
		Baz: p.Baz.Value(),
	}
}

//codegen:generated
func (p PatchExamplePatch) Merge(r PatchExamplePatch) PatchExamplePatch {
	//nolint
	return PatchExamplePatch{
		Foo: sliceund.FromOption(r.Foo.Unwrap().Or(p.Foo.Unwrap())),
		Bar: sliceund.FromOption(r.Bar.Unwrap().Or(p.Bar.Unwrap())),
		Baz: sliceund.FromOption(r.Baz.Unwrap().Or(p.Baz.Unwrap())),
	}
}

//codegen:generated
func (p PatchExamplePatch) ApplyPatch(v PatchExample) PatchExample {
	var orgP PatchExamplePatch
	orgP.FromValue(v)
	merged := orgP.Merge(p)
	return merged.ToValue()
}
```

### validator command

`validator` sub command emits generated `UndValidate` method which validates its und-state.

To generate validator, you must specify its required states by struct tag `und:""` before executing the command.

- `def` requires value to be _defined_, `null` be _null_, `und` be _undefined_. These 3 can be combined.
- `required` and `nullish` are shorthand for `def`, `null,und` respectively. Exclusive to each other and other `def`, `null`, `und`.
- `len` and `values` are only applicable to `Elastic` types.
- `len` specifies required length of field. This also has same effect specifying `def`.
  - comparison operator is placed right after `len`
  - `len>n`, `len>=n`, `len==n`, `len<n` and `len<=n` are allowed.
  - `n` is integer.
  - Operators have the same meaning as in Go.
  - Assume `len` will be replaced with your field length. `len>n` is valid when field length is greater than `n`.
- `values` currently has only `values:nonnull` variant.
  - `nonnull` variant requires all values of `Elastic` field to be non-null. As mentioned in above, normally Elastic field is `[](T | null)`.

Run command by

```
go run github.com/ngicks/go-codegen/codegen undgen validator --dir ./pkg/example --pkg ./ ...
```

```go
// example.go
type Example struct {
	Foo    string
	Bar    option.Option[string]        // no tag
	Baz    option.Option[string]        `und:"def"`
	Qux    und.Und[string]              `und:"def,und"`
	Quux   elastic.Elastic[string]      `und:"null,len==3"`
	Corge  sliceund.Und[string]         `und:"nullish"`
	Grault sliceelastic.Elastic[string] `und:"und,len>=2,values:nonnull"`
}
```

```go
// example.und_validator.go

//codegen:generated
func (v Example) UndValidate() (err error) {
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		if !validator.ValidOpt(v.Baz) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Baz))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"Baz",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
				Und: true,
			},
		}.Into()

		if !validator.ValidUnd(v.Qux) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Qux))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"Qux",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def:  true,
				Null: true,
			},
			Len: &undtag.LenValidator{
				Len: 3,
				Op:  undtag.LenOpEqEq,
			},
		}.Into()

		if !validator.ValidElastic(v.Quux) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Quux))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"Quux",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Null: true,
				Und:  true,
			},
		}.Into()

		if !validator.ValidUnd(v.Corge) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Corge))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"Corge",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
				Und: true,
			},
			Len: &undtag.LenValidator{
				Len: 2,
				Op:  undtag.LenOpGrEq,
			},
			Values: &undtag.ValuesValidator{
				Nonnull: true,
			},
		}.Into()

		if !validator.ValidElastic(v.Grault) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Grault))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"Grault",
			)
		}
	}
	return
}
```

### plain command

`plain` sub command emits generated _Plain_ types where all und-kind types are converted to _normal_ Go types,
and conversion methods between _Plain_ and _Raw_(the original) types.

To generate plain, you must specify its required states by struct tag `und:""` before executing the command as like `validator` command.

The meaning of each `und` struct tag is explained in the `validator` example.

Here's conversion rule for `plain`.

- `def` strips `Und[T]` or `Elastic[T]` into `T` or `[]option.Option[T]` respectively.
- `null`, `und` replace type with special empty type `conversion.Empty`.
- `def,null,und` is no-op. No conversion.
- `def,null` or `def,und` strips types to `Option[T]`
  - If there's `und`, it should add `,omitzero` or `,omitempty` option.
  - Otherwise it removes the option.
- `len==n` option strips `Elastic` type into `und.Und[[n]option.Option[T]]`
- `len==1` is special case where it strip `[]T` to `T`, (`und.Und[[]option.Option[T]]` -> `und.Und[option.Option[T]]`).
- `len>n`, `len>=n`, `len<n` and `len<=n` assures field length at conversion time.
  - For example, the field value of _Plain_ type converted though `UndPlain` has at least `n`+1 length if `len>n` is specified.
  - In case input was shorter, conversion method extends slice with zero value.
- `values:nonnull` unwraps `und.Und[[]option.Option[T]]` into `und.Und[[]T]`

Run command by

```
go run github.com/ngicks/go-codegen/codegen undgen plain --dir ./pkg/example --pkg ./ ...
```

```go
// example.go
type Example struct {
	Foo    string
	Bar    option.Option[string]        // no tag
	Baz    option.Option[string]        `und:"def"`
	Qux    und.Und[string]              `und:"def,und"`
	Quux   elastic.Elastic[string]      `und:"null,len==3"`
	Corge  sliceund.Und[string]         `und:"nullish"`
	Grault sliceelastic.Elastic[string] `und:"und,len>=2,values:nonnull"`
}
```

```go
//codegen:generated
type ExamplePlain struct {
	Foo    string
	Bar    option.Option[string]                   // no tag
	Baz    string                                  `und:"def"`
	Qux    option.Option[string]                   `und:"def,und"`
	Quux   option.Option[[3]option.Option[string]] `und:"null,len==3"`
	Corge  option.Option[conversion.Empty]         `und:"nullish"`
	Grault option.Option[[]string]                 `und:"und,len>=2,values:nonnull"`
}

//codegen:generated
func (v Example) UndPlain() ExamplePlain {
	return ExamplePlain{
		Foo: v.Foo,
		Bar: v.Bar,
		Baz: v.Baz.Value(),
		Qux: v.Qux.Unwrap().Value(),
		Quux: und.Map(
			conversion.UnwrapElastic(v.Quux),
			func(o []option.Option[string]) (out [3]option.Option[string]) {
				copy(out[:], o)
				return out
			},
		).Unwrap().Value(),
		Corge:  conversion.UndNullish(v.Corge),
		Grault: conversion.NonNullSlice(conversion.LenNAtLeastSlice(2, conversion.UnwrapElasticSlice(v.Grault))).Unwrap().Value(),
	}
}

//codegen:generated
func (v ExamplePlain) UndRaw() Example {
	return Example{
		Foo: v.Foo,
		Bar: v.Bar,
		Baz: option.Some(v.Baz),
		Qux: conversion.OptionUnd(false, v.Qux),
		Quux: elastic.FromUnd(und.Map(
			conversion.OptionUnd(true, v.Quux),
			func(s [3]option.Option[string]) []option.Option[string] {
				return s[:]
			},
		)),
		Corge:  conversion.NullishUndSlice[string](v.Corge),
		Grault: sliceelastic.FromUnd(conversion.NullifySlice(conversion.OptionUndSlice(false, v.Grault))),
	}
}
```

### other examples

see sub packages under https://github.com/ngicks/go-codegen/tree/main/codegen/generator/undgen/internal/testtargets
