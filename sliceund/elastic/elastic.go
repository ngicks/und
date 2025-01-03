package elastic

import (
	"encoding/json"
	"encoding/xml"
	"log/slog"

	_ "github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	"github.com/ngicks/und/validate"
)

var (
	_ json.Marshaler   = Elastic[any]{}
	_ json.Unmarshaler = (*Elastic[any])(nil)
	_ xml.Marshaler    = Elastic[any]{}
	_ xml.Unmarshaler  = (*Elastic[any])(nil)
	_ slog.LogValuer   = Elastic[any]{}
)

var (
	_ validate.UndValidator = Elastic[any]{}
	_ validate.UndChecker   = Elastic[any]{}
	_ validate.ElasticLike  = Elastic[any]{}
)

// Elastic[T] is a slice-based variant of [elastic.Elastic].
//
// Elastic[T] exposes same set of methods as [elastic.Elastic] and can be used in almost same way.
// Although it exposes its internal value, it is only intended to let some encoders, e.g. encoding/json, etc, see it as omittable value.
// You should manipulate the value only through methods.
//
// *undefined* Elastic[T] struct fields are omitted by encoding/json
// if either or both of `json:",omitempty"` and `json:",omitzero"` (for Go 1.24 or later) options are attached to those fields.
type Elastic[T any] sliceund.Und[option.Options[T]]

// Null returns a null Elastic[T].
func Null[T any]() Elastic[T] {
	return Elastic[T](sliceund.Null[option.Options[T]]())
}

// Undefined returns an undefined Elastic[T].
func Undefined[T any]() Elastic[T] {
	return Elastic[T](sliceund.Undefined[option.Options[T]]())
}

// FromOptions converts variadic option.Option[T] values into Elastic[T].
// options is retained by the returned value.
func FromOptions[T any](options ...option.Option[T]) Elastic[T] {
	if options == nil {
		// prevent accidentally returning nil options
		options = make(option.Options[T], 0)
	}
	return Elastic[T](sliceund.Defined(option.Options[T](options)))
}

// FromUnd converts sliceund.Und[option.Options[T]] into Elastic[T].
//
// u is retained by the returned value.
func FromUnd[T any, Opts ~[]option.Option[T]](u sliceund.Und[Opts]) Elastic[T] {
	switch {
	case u.IsUndefined():
		return Undefined[T]()
	case u.IsNull():
		return Null[T]()
	default:
		return Elastic[T](sliceund.Map(u, func(o Opts) option.Options[T] { return option.Options[T](o) }))
	}
}

func (e Elastic[T]) inner() sliceund.Und[option.Options[T]] {
	return sliceund.Und[option.Options[T]](e)
}

// IsDefined returns true if e is a defined Elastic[T],
// which includes a slice with no element.
func (e Elastic[T]) IsDefined() bool {
	return e.inner().IsDefined()
}

// IsNull returns true if e is a null Elastic[T].
func (e Elastic[T]) IsNull() bool {
	return e.inner().IsNull()
}

// IsUndefined returns true if e is an undefined Elastic[T].
func (e Elastic[T]) IsUndefined() bool {
	return e.inner().IsUndefined()
}

// EqualFunc reports whether two Elastic values are equal.
// EqualFunc checks state of both. If both state does not match, it returns false.
// If both are *defined* and lengths of their internal value match,
// it then checks equality of their value by cmp.
// It returns true if they are equal.
func (e Elastic[T]) EqualFunc(f Elastic[T], cmp func(i, j T) bool) bool {
	return e.inner().EqualFunc(
		f.inner(),
		func(i, j option.Options[T]) bool {
			return i.EqualFunc(j, cmp)
		},
	)
}

// Equal tests equality of l and r then returns true if they are equal, false otherwise.
// For those types that are comparable but need special tests, e.g. time.Time, you should use [Elastic.EqualFunc] instead.
//
// Equal is a specialized [slices.Equal] where it also considers value state of l and r.
func Equal[T comparable](l, r Elastic[T]) bool {
	return l.EqualFunc(r, func(i, j T) bool { return i == j })
}

func (e Elastic[T]) CloneFunc(cloneT func(T) T) Elastic[T] {
	return Elastic[T](e.inner().CloneFunc(func(o option.Options[T]) option.Options[T] { return o.CloneFunc(cloneT) }))
}

// Clone clones e.
func Clone[T comparable](e Elastic[T]) Elastic[T] {
	return e.CloneFunc(func(t T) T { return t })
}

// Len returns length of values.
func (e Elastic[T]) Len() int {
	if !e.IsDefined() {
		return 0
	}
	return len(e.Unwrap().Value())
}

// HasNull reports e is defined value has null in ins value.
func (e Elastic[T]) HasNull() bool {
	if !e.IsDefined() {
		return false
	}
	for _, o := range e.Unwrap().Value() {
		if o.IsNone() {
			return true
		}
	}
	return false
}

// Value returns a first value of its internal option slice if e is defined.
// Otherwise it returns zero value for T.
func (e Elastic[T]) Value() T {
	if e.IsDefined() {
		vs := e.inner().Value()
		if len(vs) > 0 {
			return vs[0].Value()
		}
	}
	var zero T
	return zero
}

// Values returns internal option slice as plain []T.
//
// If e is not defined, it returns nil.
// Any None value in its internal option slice will be converted
// to zero value of T.
func (e Elastic[T]) Values() []T {
	if !e.IsDefined() {
		return []T(nil)
	}
	opts := e.inner().Value()
	vs := make([]T, len(opts))
	for i, opt := range opts {
		vs[i] = opt.Value()
	}
	return vs
}

// Pointer returns a first value of its internal option slice as *T if e is defined.
//
// Pointer returns nil if
//   - e is not defined
//   - e has no element
//   - e's first element is None.
func (e Elastic[T]) Pointer() *T {
	if e.IsDefined() {
		vs := e.inner().Value()
		if len(vs) > 0 && vs[0].IsSome() {
			v := vs[0].Value()
			return &v
		}
	}
	return nil
}

// Pointer returns its internal option slice as []*T if e is defined.
func (e Elastic[T]) Pointers() []*T {
	if !e.IsDefined() {
		return nil
	}
	opts := e.inner().Value()
	ptrs := make([]*T, len(opts))
	for i, opt := range opts {
		ptrs[i] = opt.Pointer()
	}
	return ptrs
}

// Unwrap unwraps e.
func (e Elastic[T]) Unwrap() sliceund.Und[option.Options[T]] {
	return e.inner()
}

// Map returns a new Elastic[T] whose internal value is e's mapped by f.
//
// The internal slice of e is capped to its length before passed to f.
func (e Elastic[T]) Map(f func(sliceund.Und[option.Options[T]]) sliceund.Und[option.Options[T]]) Elastic[T] {
	return Elastic[T](
		f(e.inner().Map(func(o option.Option[option.Option[option.Options[T]]]) option.Option[option.Option[option.Options[T]]] {
			return o.Map(func(v option.Option[option.Options[T]]) option.Option[option.Options[T]] {
				return v.Map(func(v option.Options[T]) option.Options[T] {
					return v[:len(v):len(v)]
				})
			})
		})),
	)
}

// UnmarshalXML implements xml.Unmarshaler.
func (e *Elastic[T]) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var t option.Options[T]
	err := d.DecodeElement(&t, &start)
	if err != nil {
		return err
	}

	if len(e.inner().Value()) == 0 {
		*e = FromOptions(t...)
	} else {
		*e = e.Map(func(u sliceund.Und[option.Options[T]]) sliceund.Und[option.Options[T]] {
			return u.Map(func(o option.Option[option.Option[option.Options[T]]]) option.Option[option.Option[option.Options[T]]] {
				return o.Map(func(v option.Option[option.Options[T]]) option.Option[option.Options[T]] {
					return v.Map(func(v option.Options[T]) option.Options[T] {
						return append(v, t...)
					})
				})
			})
		})
	}
	return nil
}
