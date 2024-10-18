package und

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"log/slog"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/validate"
)

var (
	_ json.Marshaler       = Und[any]{}
	_ json.Unmarshaler     = (*Und[any])(nil)
	_ jsonv2.MarshalerV2   = Und[any]{}
	_ jsonv2.UnmarshalerV2 = (*Und[any])(nil)
	_ xml.Marshaler        = Und[any]{}
	_ xml.Unmarshaler      = (*Und[any])(nil)
	_ slog.LogValuer       = Und[any]{}
)

var (
	_ option.Equality[Und[any]] = Und[any]{}
	_ option.Cloner[Und[any]]   = Und[any]{}
	_ validate.ValidatorUnd     = Und[any]{}
	_ validate.CheckerUnd       = Und[any]{}
)

// Und[T] is a type that can express a value (`T`), empty (`null`), or absent (`undefined`).
// Und[T] is comparable if T is comparable. And it can be copied by assign.
//
// Und[T] implements IsZero and is omitted when is a struct field of other structs and appropriate marshalers and appropriate struct tag on the field,
// e.g. "github.com/go-json-experiment/json/jsontext" with omitzero option set to the field,
// or "github.com/json-iterator/go" with omitempty option to the field and an appropriate extension.
//
// If you need to stick with encoding/json v1, you can use github.com/ngicks/und/sliceund,
// a slice based version of Und[T] whish is already skppable by v1.
type Und[T any] struct {
	opt option.Option[option.Option[T]]
}

// Defined returns a defined Und[T] whose internal value is t.
func Defined[T any](t T) Und[T] {
	return Und[T]{
		opt: option.Some(option.Some(t)),
	}
}

// Null returns a null Und[T].
func Null[T any]() Und[T] {
	return Und[T]{
		opt: option.Some(option.None[T]()),
	}
}

// Undefined returns an undefined Und[T].
func Undefined[T any]() Und[T] {
	return Und[T]{}
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
	return Und[T]{opt: opt}
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
func (u Und[T]) IsZero() bool {
	return u.IsUndefined()
}

// IsDefined returns true if u is a defined value, otherwise false.
func (u Und[T]) IsDefined() bool {
	return u.opt.IsSome() && u.opt.Value().IsSome()
}

// IsNull returns true if u is a null value, otherwise false.
func (u Und[T]) IsNull() bool {
	return u.opt.IsSome() && u.opt.Value().IsNone()
}

// IsUndefined returns true if u is an undefined value, otherwise false.
func (u Und[T]) IsUndefined() bool {
	return u.opt.IsNone()
}

// Equal implements Equality[Und[T]].
// Equal panics if T is uncomparable and does not implement Equality[T].
func (u Und[T]) Equal(other Und[T]) bool {
	return u.opt.Equal(other.opt)
}

// EqualFunc reports whether two Und values are equal.
// EqualFunc checks state of both. If both state does not match, it returns false.
// If both are "defined" state, then checks equality of their value by cmp,
// then returns true if they are equal.
func (u Und[T]) EqualFunc(other Und[T], cmp func(i, j T) bool) bool {
	return u.opt.EqualFunc(
		other.opt,
		func(i, j option.Option[T]) bool {
			return i.EqualFunc(j, cmp)
		},
	)
}

// Clone clones u.
// This is only a copy-by-assign unless T implements Cloner[T].
func (u Und[T]) Clone() Und[T] {
	return u.Map(func(o option.Option[option.Option[T]]) option.Option[option.Option[T]] { return o.Clone() })
}

func (u Und[T]) ValidateUnd() error {
	return u.opt.Value().ValidateUnd()
}

func (u Und[T]) CheckUnd() error {
	return u.opt.CheckUnd()
}

// Value returns an internal value.
func (u Und[T]) Value() T {
	if u.IsDefined() {
		return u.opt.Value().Value()
	}
	var zero T
	return zero
}

// Pointer returns u's internal value as a pointer.
func (u Und[T]) Pointer() *T {
	if !u.IsDefined() {
		return nil
	}
	return u.opt.Value().Pointer()
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
		t := u.opt.Value().Value()
		tt := &t
		return &tt
	}
}

// Unwrap returns u's internal value.
func (u Und[T]) Unwrap() option.Option[option.Option[T]] {
	return u.opt
}

// Map returns a new Und[T] whose internal value is u's mapped by f.
func (u Und[T]) Map(f func(option.Option[option.Option[T]]) option.Option[option.Option[T]]) Und[T] {
	return Und[T]{opt: f(u.opt)}
}

// MarshalJSON implements json.Marshaler.
func (u Und[T]) MarshalJSON() ([]byte, error) {
	if !u.IsDefined() {
		return []byte(`null`), nil
	}
	return json.Marshal(u.opt.Value().Value())
}

// UnmarshalJSON implements json.Unmarshaler.
func (u *Und[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*u = Null[T]()
		return nil
	}

	var t T
	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}

	*u = Defined(t)
	return nil
}

// MarshalJSONV2 implements jsonv2.MarshalerV2.
func (u Und[T]) MarshalJSONV2(enc *jsontext.Encoder, opts jsonv2.Options) error {
	if !u.IsDefined() {
		return enc.WriteToken(jsontext.Null)
	}
	return jsonv2.MarshalEncode(enc, u.opt.Value().Value(), opts)
}

// UnmarshalJSONV2 implements jsonv2.UnmarshalerV2.
func (u *Und[T]) UnmarshalJSONV2(dec *jsontext.Decoder, opts jsonv2.Options) error {
	if dec.PeekKind() == 'n' {
		err := dec.SkipValue()
		if err != nil {
			return err
		}
		*u = Null[T]()
		return nil
	}
	var t T
	err := jsonv2.UnmarshalDecode(dec, &t, opts)
	if err != nil {
		return err
	}
	*u = Defined(t)
	return nil
}

// MarshalXML implements xml.Marshaler.
func (o Und[T]) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return o.opt.Value().MarshalXML(e, start)
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
	return u.opt.Value().LogValue()
}

// SqlNull converts o into sql.Null[T].
func (u Und[T]) SqlNull() sql.Null[T] {
	return u.opt.Value().SqlNull()
}
