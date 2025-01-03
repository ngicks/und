package und

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"log/slog"

	"github.com/ngicks/und/option"
	"github.com/ngicks/und/validate"
)

var (
	_ json.Marshaler   = Und[any]{}
	_ json.Unmarshaler = (*Und[any])(nil)
	_ xml.Marshaler    = Und[any]{}
	_ xml.Unmarshaler  = (*Und[any])(nil)
	_ slog.LogValuer   = Und[any]{}
)

var (
	_ validate.UndValidator = Und[any]{}
	_ validate.UndChecker   = Und[any]{}
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
//
// If you need to keep t as pointer, use [WrapPointer] instead.
func FromPointer[T any](v *T) Und[T] {
	if v == nil {
		return Undefined[T]()
	}
	return Defined(*v)
}

// WrapPointer converts *T into Und[*T].
// The und value is defined if t is non nil, undefined otherwise.
//
// If you want t to be dereferenced, use [FromPointer] instead.
func WrapPointer[T any](t *T) Und[*T] {
	if t == nil {
		return Undefined[*T]()
	}
	return Defined(t)
}

// FromOptions converts opt into an Und[T].
// opt is retained by the returned value.
func FromOption[T any](opt option.Option[option.Option[T]]) Und[T] {
	return Und[T]{opt: opt}
}

// FromSqlNull converts a valid sql.Null[T] to a defined Und[T]
// and invalid one into a null Und[T].
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

// EqualFunc reports whether two Und values are equal.
// EqualFunc checks state of both. If both state does not match, it returns false.
// If both are *defined* state, then it checks equality of their value by cmp,
// then returns true if they are equal.
func (u Und[T]) EqualFunc(t Und[T], cmp func(i, j T) bool) bool {
	return u.opt.EqualFunc(
		t.opt,
		func(i, j option.Option[T]) bool {
			return i.EqualFunc(j, cmp)
		},
	)
}

// Equal tests equality of l and r then returns true if they are equal, false otherwise.
// For those types that are comparable but need special tests, e.g. time.Time, you should use [Und.EqualFunc] instead.
//
// This only sits here only to keep consistency to sliceund, elastic, sliceund/elastic.
// You can simply test their equality by only doing l == r.
func Equal[T comparable](l, r Und[T]) bool {
	return l.EqualFunc(r, func(i, j T) bool { return i == j })
}

// CloneFunc clones u using the cloneT functions.
func (u Und[T]) CloneFunc(cloneT func(T) T) Und[T] {
	return u.Map(func(o option.Option[option.Option[T]]) option.Option[option.Option[T]] {
		return o.CloneFunc(func(o option.Option[T]) option.Option[T] {
			return o.CloneFunc(cloneT)
		})
	})
}

// Clone clones u.
//
// It just returns u; this only sits here only for consistency to sliceund, elastic, sliceund/elastic.
func Clone[T comparable](u Und[T]) Und[T] {
	return u
}

func (u Und[T]) UndValidate() error {
	return u.opt.Value().UndValidate()
}

func (u Und[T]) UndCheck() error {
	return u.opt.UndCheck()
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

// State returns u's value state.
func (u Und[T]) State() State {
	switch {
	case u.IsUndefined():
		return StateUndefined
	case u.IsNull():
		return StateNull
	default:
		return StateDefined
	}
}

// Map returns a new Und value whose internal value is mapped by f.
func Map[T, U any](u Und[T], f func(t T) U) Und[U] {
	switch {
	case u.IsUndefined():
		return Undefined[U]()
	case u.IsNull():
		return Null[U]()
	default:
		return Defined(f(u.Value()))
	}
}
