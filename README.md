# und - option and undefined-able types, mainly for JSON fields. [![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/ngicks/und)

Types to interoperate with applications that make full use of JSON.

## Express T | null | undefined by only using the types in struct field.

Just use `und.Und` (for Go 1.24 or later) or `sliceund.Und` (for Go 1.23 or earlier) as struct field type then place `,omitzero`, `,omitempty` respectively.

```go
type sample struct {
    Foo und.Und[string]      `json:",omitzero"`
    Bar sliceund.Und[string] `json:",omitempty"`
}
```

The zero value is *undefined*, `encoding/json` skips fields.

```go
s := sample{}

bin, _ := json.MarshalIndent(s, "", "    ")
fmt.Printf("zero = %s\n", bin)
/*
    zero = {}
*/
```

Use `Null` functions to create *null* objects.

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

Use `Defined` functions to create *defined* objects.

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

Use `Undefined` functions to create *undefined*  objects.

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
  - have `Equal` method in case `T` is not comparable or comparable but needs custom equality tests(e.g. `time.Time`)
    - `T` should implement `type Equality[T any] interface { Equal(T) bool }` interface.
  - have `EqualFunc` method for cases where `T` is not comparable and does not implement `Equality`.
  - has convenient methods stolen from rust's `core::option::Option<T>`
  - can be used in some (not all) place of `*T`
  - is copied by assign.

Other types are based on `Option[T]`.

- `Und[T]`: *undefined* (*empty* or *unspecified*), *null* or `T` (any type you like)
- `Elastic[T]`: *undefined* (*empty* or *unspecified*), *null*, `T` or [](`T` | null)
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

run example by

Note: this will be fixed after Go 1.24 is released.

```
go install golang.org/dl/go1.24rc1@latest
go1.24rc1 download
go1.24rc1 run github.com/ngicks/und/example@77f793d0c981807e245c2d3c96dd5b4e3f0f6656
```

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

```
go run github.com/ngicks/go-codegen/codegen undgen patch     -v --dir /path/to/root/dir/of/target/package --pkg ./path/to/package ...
go run github.com/ngicks/go-codegen/codegen undgen validator -v --dir /path/to/root/dir/of/target/package --pkg ./...
go run github.com/ngicks/go-codegen/codegen undgen plain     -v --dir /path/to/root/dir/of/target/package --pkg ./...
```

Notable flags:

- `-v`   : verbose logs.
- `--dir`: specify directory under which the target packages are placed.
- `--pkg`: same package pattern that can be passed to `go list`. must be prefixed with `./`. `patch` sub command only accept pattern that matches only a single package.
- `types...`: the `patch` sub command needs `types...` arguments to specify target type names. Use `...` to target all types found under `--pkg`.

Examples below assumes `example.go` is defined in `./pkg/example`

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

- `def` requires value to be *defined*, `null` be *null*, `und` be *undefined*. These 3 can be combined.
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

`plain` sub command emits generated *Plain* types where all und-kind types are converted to *normal* Go types,
and conversion methods between *Plain* and *Raw*(the original) types.

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
  - For example, the field value of *Plain* type converted though `UndPlain` has at least `n`+1 length if `len>n` is specified.
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