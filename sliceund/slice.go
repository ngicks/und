package sliceund

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"log/slog"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/validate"
)

var (
	_ option.Equality[Und[any]] = Und[any]{}
	_ option.Cloner[Und[any]]   = Und[any]{}
	_ validate.ValidatorUnd     = Und[any]{}
	_ validate.CheckerUnd       = Und[any]{}
	_ json.Marshaler            = Und[any]{}
	_ json.Unmarshaler          = (*Und[any])(nil)
	_ jsonv2.MarshalerV2        = Und[any]{}
	_ jsonv2.UnmarshalerV2      = (*Und[any])(nil)
	_ xml.Marshaler             = Und[any]{}
	_ xml.Unmarshaler           = (*Und[any])(nil)
	_ slog.LogValuer            = Und[any]{}
)

// Und[T] is a type that can express a value (`T`), empty (`null`), or absent (`undefined`).
//
// Und[T] can be a skippable struct field with omitempty option of `encoding/json`.
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

// FromPointer converts *T into Und[T].
// If v is nil, it returns an undefined Und.
// Otherwise, it returns Defined[T] whose value is the dereferenced v.
func FromPointer[T any](v *T) Und[T] {
	if v == nil {
		return Undefined[T]()
	}
	return Defined(*v)
}

// FromOptions converts opt into an Und[T].
// opt is retained by the returned value.
func FromOption[T any](opt option.Option[option.Option[T]]) Und[T] {
	if opt.IsNone() {
		return Undefined[T]()
	}
	return Und[T]{opt.Value()}
}

// FromUnd converts non-slice version of Und[T] into Und[T].
func FromUnd[T any](u und.Und[T]) Und[T] {
	return FromOption(u.Unwrap())
}

// FromSqlNull converts a valid sql.Null[T] to a defined Und[T]
// and invalid one into a null Und[].
func FromSqlNull[T any](v sql.Null[T]) Und[T] {
	if !v.Valid {
		return Null[T]()
	}
	return Defined(v.V)
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

// IsNull returns true if u is a null value, otherwise false.
func (u Und[T]) IsNull() bool {
	return len(u) > 0 && u[0].IsNone()
}

// IsUndefined returns true if u is an undefined value, otherwise false.
func (u Und[T]) IsUndefined() bool {
	return len(u) == 0
}

// Value returns an internal value.
func (u Und[T]) Value() T {
	if u.IsDefined() {
		return u[0].Value()
	}
	var zero T
	return zero
}

// MarshalJSON implements json.Marshaler.
func (u Und[T]) MarshalJSON() ([]byte, error) {
	if !u.IsDefined() {
		return []byte(`null`), nil
	}
	return json.Marshal(u[0].Value())
}

// UnmarshalJSON implements json.Unmarshaler.
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

// MarshalJSONV2 implements jsonv2.MarshalerV2.
func (u Und[T]) MarshalJSONV2(enc *jsontext.Encoder, opts jsonv2.Options) error {
	if !u.IsDefined() {
		return enc.WriteToken(jsontext.Null)
	}
	return jsonv2.MarshalEncode(enc, u.Value(), opts)
}

// UnmarshalJSONV2 implements jsonv2.UnmarshalerV2.
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

// Equal implements Equality[Und[T]].
// Equal panics if T is uncomparable and does not implement Equality[T].
func (u Und[T]) Equal(other Und[T]) bool {
	if u.IsUndefined() || other.IsUndefined() {
		return u.IsUndefined() == other.IsUndefined()
	}
	return u[0].Equal(other[0])
}

// Clone clones u.
// This is only a copy-by-assign unless T implements Cloner[T].
func (u Und[T]) Clone() Und[T] {
	return u.Map(func(o option.Option[option.Option[T]]) option.Option[option.Option[T]] { return o.Clone() })
}

func (u Und[T]) ValidateUnd() error {
	return u.Unwrap().Value().ValidateUnd()
}

func (u Und[T]) CheckUnd() error {
	return u.Unwrap().Value().CheckUnd()
}

// Pointer returns u's internal value as a pointer.
func (u Und[T]) Pointer() *T {
	if !u.IsDefined() {
		return nil
	}
	v := u.Value()
	return &v
}

// DoublePointer returns nil if u is undefined, &(*T)(nil) if null, the internal value if defined.
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

// Unwrap converts u to a nested options.
func (u Und[T]) Unwrap() option.Option[option.Option[T]] {
	if u.IsUndefined() {
		return option.None[option.Option[T]]()
	}
	opt := u[0] // copy by assign; it's a value.
	return option.Some(opt)
}

// Und converts u into non-slice version Und[T].
func (u Und[T]) Und() und.Und[T] {
	return und.FromOption(u.Unwrap())
}

// Map returns a new Und[T] whose internal value is u's mapped by f.
func (u Und[T]) Map(f func(option.Option[option.Option[T]]) option.Option[option.Option[T]]) Und[T] {
	return FromOption(f(u.Unwrap()))
}

// MarshalXML implements xml.Marshaler.
func (o Und[T]) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return o.Unwrap().Value().MarshalXML(e, start)
}

// UnmarshalXML implements xml.Unmarshaler.
func (o *Und[T]) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var t T
	err := d.DecodeElement(&t, &start)
	if err != nil {
		return err
	}

	*o = Defined(t)

	return nil
}

// LogValue implements slog.LogValuer.
func (u Und[T]) LogValue() slog.Value {
	return u.Unwrap().Value().LogValue()
}

// SqlNull converts o into sql.Null[T].
func (u Und[T]) SqlNull() sql.Null[T] {
	return u.Unwrap().Value().SqlNull()
}
