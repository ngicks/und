package elastic

import (
	"encoding/json"
	"encoding/xml"
	"log/slog"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/validate"
)

var (
	_ option.Equality[Elastic[any]] = Elastic[any]{}
	_ option.Cloner[Elastic[any]]   = Elastic[any]{}
	_ json.Marshaler                = Elastic[any]{}
	_ json.Unmarshaler              = (*Elastic[any])(nil)
	_ jsonv2.MarshalerV2            = Elastic[any]{}
	// We don't implement UnmarshalJSONV2 since there's variants that cannot be unmarshaled without
	// calling unmarshal twice or so.
	// there's 4 possible code paths
	//
	//   - input is T
	//   - input is []T
	//   - input starts with [ but T is []U
	//   - input starts with [ but T implements UnmarshalJSON v1 or v2; it's ambiguous.
	//
	// That'll needs unnecessary complexity to code base, e.g. teeing tokens and token stream decoder.
	//
	// _ jsonv2.UnmarshalerV2          = (*Elastic[any])(nil)
	_ slog.LogValuer = Elastic[any]{}
)

var (
	_ validate.ValidatorUnd = Elastic[any]{}
	_ validate.CheckerUnd   = Elastic[any]{}
	_ validate.ElasticLike  = Elastic[any]{}
)

// Elastic[T] is a type that can express undefined | null | T | [](null | T).
// Elastic[T] is comparable if T is comparable. And it can be copied by assign.
//
// Elastic[T] implements IsZero and can be skippable struct fields when marshaled through appropriate marshalers,
// e.g. "github.com/go-json-experiment/json/jsontext" with omitzero option set to the field,
// or "github.com/json-iterator/go" with omitempty option to the field and an appropriate extension.
//
// If you need to stick with encoding/json v1, you can use github.com/ngicks/und/sliceund/elastic,
// a slice based version of Elastic[T] whish is already skppable by v1.
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

// FromOptions converts slice of option.Option[T] into Elastic[T].
// options is retained by the returned value.
func FromOptions[Opts ~[]option.Option[T], T any](options Opts) Elastic[T] {
	return Elastic[T]{
		v: und.Defined(option.Options[T](options)),
	}
}

// FromUnd converts und.Und[option.Options[T]] into Elastic[T].
//
// The internal value of u is retained by the returned value.
func FromUnd[T any](u und.Und[option.Options[T]]) Elastic[T] {
	return Elastic[T]{u}
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

// Equal implements option.Equality[Elastic[T]].
//
// Equal panics if T is uncomparable.
func (e Elastic[T]) Equal(other Elastic[T]) bool {
	return e.v.Equal(other.v)
}

// Clone implements option.Cloner[Elastic[T]].
//
// Clone clones its internal option.Option slice by copy.
// Or if T implements Cloner[T], each element is cloned.
func (e Elastic[T]) Clone() Elastic[T] {
	return Elastic[T]{v: e.v.Clone()}
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

// Map returns a new Elastic[T] whose internal value is e's mapped by f.
//
// The internal slice of e is capped to its length before passed to f.
func (e Elastic[T]) Map(f func(und.Und[option.Options[T]]) und.Und[option.Options[T]]) Elastic[T] {
	return Elastic[T]{
		v: f(e.v.Map(func(o option.Option[option.Option[option.Options[T]]]) option.Option[option.Option[option.Options[T]]] {
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
		*e = FromOptions(t)
	} else {
		*e = e.Map(func(u und.Und[option.Options[T]]) und.Und[option.Options[T]] {
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
