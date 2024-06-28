package sliceund

import (
	"encoding/json"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
)

var (
	_ option.Equality[Und[any]] = Und[any]{}
	_ json.Marshaler            = Und[any]{}
	_ json.Unmarshaler          = (*Und[any])(nil)
	_ jsonv2.MarshalerV2        = Und[any]{}
	_ jsonv2.UnmarshalerV2      = (*Und[any])(nil)
)

// Und[T] is an uncomparable version of und.Und[T].
// It can be skipped by v1 encoding/json when when marshaling if len(u) == 0.
//
// Although it exposes its internal data structure,
// you should not mutate internal data.
// Using map[T]U, []T or json.RawMessage as base type is only allowed hacks
// to make it skippable by `json:",omitempty" option,
// without losing freedom of adding methods.
// Any method implemented on Und[T] assumes
// it has only either 0 or 1 element.
// So mutating an Und[T] object, e.g. appending it to have 2 or more elements,
// causes undefined behaviors and not promised to behave same between versions.
//
// Und[T] is intended to behave much like a simple variable.
// There are only 2 way to change its internal state.
// Assigning a value of corresponding state to the variable you intend to change.
// Or calling UnmarshalJSON on an addressable Und[T].
type Und[T any] []option.Option[T]

// Defined returns a `defined` Und[T] which contains t.
func Defined[T any](t T) Und[T] {
	return Und[T]{option.Some(t)}
}

// Null returns a `null` Und[T].
func Null[T any]() Und[T] {
	return Und[T]{option.None[T]()}
}

// Undefined returns an `undefined` Und[T].
func Undefined[T any]() Und[T] {
	return nil
}

func FromPointer[T any](v *T) Und[T] {
	if v == nil {
		return Undefined[T]()
	}
	return Defined(*v)
}

func FromOption[T any](opt option.Option[option.Option[T]]) Und[T] {
	if opt.IsNone() {
		return Undefined[T]()
	}
	return Und[T]{opt.Value()}
}

func FromUnd[T any](u und.Und[T]) Und[T] {
	return FromOption(u.Unwrap())
}

// IsZero is an alias for IsUndefined.
// Using `json:",omitzero"` option with "github.com/go-json-experiment/json"
// skips this field while encoding if IsZero returns true.
func (u Und[T]) IsZero() bool {
	return u.IsUndefined()
}

// IsDefined returns true if u contains a value.
// Through this method, you can check validity of the value returned from Value method.
func (u Und[T]) IsDefined() bool {
	return len(u) > 0 && u[0].IsSome()
}

func (u Und[T]) IsNull() bool {
	return len(u) > 0 && u[0].IsNone()
}

func (u Und[T]) IsUndefined() bool {
	return len(u) == 0
}

func (u Und[T]) Value() T {
	if u.IsDefined() {
		return u[0].Value()
	}
	var zero T
	return zero
}

func (u Und[T]) MarshalJSON() ([]byte, error) {
	if !u.IsDefined() {
		return []byte(`null`), nil
	}
	return json.Marshal(u[0].Value())
}

func (u *Und[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		if len(*u) == 0 {
			*u = []option.Option[T]{option.None[T]()}
		} else {
			(*u)[0] = option.None[T]()
		}
		return nil
	}

	var t T
	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}

	if len(*u) == 0 {
		*u = []option.Option[T]{option.Some(t)}
	} else {
		(*u)[0] = option.Some(t)
	}
	return nil
}

func (u Und[T]) MarshalJSONV2(enc *jsontext.Encoder, opts jsonv2.Options) error {
	if !u.IsDefined() {
		return enc.WriteToken(jsontext.Null)
	}
	return jsonv2.MarshalEncode(enc, u.Value(), opts)
}

func (u *Und[T]) UnmarshalJSONV2(dec *jsontext.Decoder, opts jsonv2.Options) error {
	if dec.PeekKind() == 'n' {
		err := dec.SkipValue()
		if err != nil {
			return err
		}
		if len(*u) == 0 {
			*u = []option.Option[T]{option.None[T]()}
		} else {
			(*u)[0] = option.None[T]()
		}
		return nil
	}
	var t T
	err := jsonv2.UnmarshalDecode(dec, &t, opts)
	if err != nil {
		return err
	}

	if len(*u) == 0 {
		*u = []option.Option[T]{option.Some(t)}
	} else {
		(*u)[0] = option.Some(t)
	}
	return nil
}

func (u Und[T]) Equal(other Und[T]) bool {
	if u.IsUndefined() || other.IsUndefined() {
		return u.IsUndefined() == other.IsUndefined()
	}
	return u[0].Equal(other[0])
}

func (u Und[T]) Pointer() *T {
	if !u.IsDefined() {
		return nil
	}
	v := u.Value()
	return &v
}

func (u Und[T]) DoublePointer() **T {
	switch {
	case u.IsUndefined():
		return nil
	case u.IsNull():
		var t *T
		return &t
	default:
		t := u.Value()
		tt := &t
		return &tt
	}
}

func (u Und[T]) Unwrap() option.Option[option.Option[T]] {
	if u.IsUndefined() {
		return option.None[option.Option[T]]()
	}
	opt := u[0] // copy by assign; it's a value.
	return option.Some(opt)
}

func (u Und[T]) Map(f func(option.Option[option.Option[T]]) option.Option[option.Option[T]]) Und[T] {
	return FromOption(f(u.Unwrap()))
}
