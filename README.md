# und - option and undefined-able types, mainly for JSON fields.

Types to interoperate with applications that make full use of JSON.

## Before v1

- und waits for release of `encoding/json/v2`
  - und depends on `github.com/go-json-experiment/json` which is an experimental `encoding/json/v2` implementation.
  - Types defined in this module implement `json.MarshalerV2` and `json.UnmarshalerV2`. The API dependency is relatively thin and narrow. I suspects they will not break the interface the part where we are relaying.
- It'll eventually have a breaking change when `encoding/json/v2` is released.
  - However that should not need change of your code, just bump version.

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
  - skippable only if encoding through
    - `github.com/go-json-experiment/json`(possibly a future `encoding/json/v2`) with the `,omitzero` options
    - [jsoniter](https://github.com/json-iterator/go) with `,omitempty` and custom encoder.
- `github.com/ngicks/und/sliceund`: `[]Option[T]` based types.
  - skippable with `,omitempty`.
