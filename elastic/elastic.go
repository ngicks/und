package elastic

import (
	"encoding/json"
	"encoding/xml"
	"log/slog"

	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
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

// Elastic[T] is a type that can express *undefined* | *null* | T | [](null | T).
// Elastic[T] implements json.Unmarshaler so that it can be unmarshaled from all of those type.
// However it always marshaled into an array of JSON value that corresponds to T.
//
// Elastic[T] is defined mainly to create/consume JSON documents stored in Elasticsearch.
// It may also be useful for parsing hand-written configuration files.
//
// Elastic[T] implements IsZero.
// For Go 1.24 or later version, *undefined* Elastic[T] struct fields are omitted by encoding/json
// if `json:",omitzero"` option is attached to those fields.
// For Go 1.23 or older version, instead you can use the sliceund/elastic variant with `json:",omitempty"` option.
type Elastic[T any] struct {
	v und.Und[option.Options[T]]
}

// Null returns a null Elastic[T].
func Null[T any]() Elastic[T] {
	return Elastic[T]{
		v: und.Null[option.Options[T]](),
	}
}

// Undefined returns an undefined Elastic[T].
func Undefined[T any]() Elastic[T] {
	return Elastic[T]{
		v: und.Undefined[option.Options[T]](),
	}
}

// FromOptions converts variadic option.Option[T] values into Elastic[T].
// options is retained by the returned value.
func FromOptions[T any](options ...option.Option[T]) Elastic[T] {
	if options == nil {
		// prevent accidentally returning nil options
		options = make(option.Options[T], 0)
	}
	return Elastic[T]{
		v: und.Defined(option.Options[T](options)),
	}
}

// FromUnd converts und.Und[option.Options[T]] into Elastic[T].
//
// The internal value of u is retained by the returned value.
func FromUnd[T any, Opts ~[]option.Option[T]](u und.Und[Opts]) Elastic[T] {
	return Elastic[T]{und.Map(u, func(o Opts) option.Options[T] { return option.Options[T](o) })}
}

func (e Elastic[T]) inner() und.Und[option.Options[T]] {
	return e.v
}

// IsDefined returns true if e is a defined Elastic[T],
// which includes a slice with no element.
func (e Elastic[T]) IsDefined() bool {
	return e.v.IsDefined()
}

// IsNull returns true if e is a null Elastic[T].
func (e Elastic[T]) IsNull() bool {
	return e.v.IsNull()
}

// IsUndefined returns true if e is an undefined Elastic[T].
func (e Elastic[T]) IsUndefined() bool {
	return e.v.IsUndefined()
}

// EqualFunc reports whether two Elastic values are equal.
// EqualFunc checks state of both. If both state does not match, it returns false.
// If both are "defined" and lengths of their internal value match,
// it then checks equality of their value by cmp.
// It returns true if they are equal.
//
// If T is just a comparable type, use [Equal].
// If T is an implementor of interface { Equal(t T) bool }, e.g time.Time, use [EqualEqualer].
func (e Elastic[T]) EqualFunc(other Elastic[T], cmp func(i, j T) bool) bool {
	return e.v.EqualFunc(
		other.v,
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

// EqualEqualer tests equality of l and r by calling Equal method implemented on l.
func EqualEqualer[T interface{ Equal(t T) bool }](l, r Elastic[T]) bool {
	return l.EqualFunc(r, func(i, j T) bool {
		return i.Equal(j)
	})
}

func (e Elastic[T]) CloneFunc(cloneT func(T) T) Elastic[T] {
	return e.InnerMap(func(u und.Und[option.Options[T]]) und.Und[option.Options[T]] {
		return u.CloneFunc(func(o option.Options[T]) option.Options[T] {
			return o.CloneFunc(cloneT)
		})
	})
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
	return len(e.v.Value())
}

// HasNull reports e is defined value has null in ins value.
func (e Elastic[T]) HasNull() bool {
	if !e.IsDefined() {
		return false
	}
	for _, o := range e.v.Value() {
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
		vs := e.v.Value()
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
		return nil
	}
	opts := e.v.Value()
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
		vs := e.v.Value()
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
	opts := e.v.Value()
	ptrs := make([]*T, len(opts))
	for i, opt := range opts {
		ptrs[i] = opt.Pointer()
	}
	return ptrs
}

// Unwrap unwraps e.
func (e Elastic[T]) Unwrap() und.Und[option.Options[T]] {
	return e.v
}

// Deprecated: Renamed to [Elastic.InnerMap]. The method had same name but behavior was inconsistent to [Map].
func (e Elastic[T]) Map(f func(und.Und[option.Options[T]]) und.Und[option.Options[T]]) Elastic[T] {
	return e.InnerMap(f)
}

// InnerMap returns a new Elastic[T] whose internal value is e's mapped by f.
// Unlike [Map], f is always called, even when e is not a defined value.
//
// The internal slice of e is capped to its length before passed to f.
func (e Elastic[T]) InnerMap(f func(und.Und[option.Options[T]]) und.Und[option.Options[T]]) Elastic[T] {
	return Elastic[T]{
		v: f(e.v.InnerMap(func(o option.Option[option.Option[option.Options[T]]]) option.Option[option.Option[option.Options[T]]] {
			return o.Map(func(v option.Option[option.Options[T]]) option.Option[option.Options[T]] {
				return v.Map(func(v option.Options[T]) option.Options[T] {
					return v[:len(v):len(v)]
				})
			})
		})),
	}
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
		*e = e.InnerMap(func(u und.Und[option.Options[T]]) und.Und[option.Options[T]] {
			return u.InnerMap(func(o option.Option[option.Option[option.Options[T]]]) option.Option[option.Option[option.Options[T]]] {
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
