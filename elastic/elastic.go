package elastic

import (
	"encoding/json"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
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
	// That'll needs
	// _ jsonv2.UnmarshalerV2          = (*Elastic[any])(nil)
)

type Elastic[T any] struct {
	v und.Und[option.Options[T]]
}

func Null[T any]() Elastic[T] {
	return Elastic[T]{
		v: und.Null[option.Options[T]](),
	}
}

func Undefined[T any]() Elastic[T] {
	return Elastic[T]{
		v: und.Undefined[option.Options[T]](),
	}
}

func FromOptions[T any, Opts ~[]option.Option[T]](options Opts) Elastic[T] {
	return Elastic[T]{
		v: und.Defined(option.Options[T](options)),
	}
}

func (e Elastic[T]) inner() und.Und[option.Options[T]] {
	return e.v
}

func (e Elastic[T]) IsDefined() bool {
	return e.v.IsDefined()
}

func (e Elastic[T]) IsNull() bool {
	return e.v.IsNull()
}

func (e Elastic[T]) IsUndefined() bool {
	return e.v.IsUndefined()
}

func (e Elastic[T]) Equal(other Elastic[T]) bool {
	return e.v.Equal(other.v)
}

func (e Elastic[T]) Clone() Elastic[T] {
	return Elastic[T]{v: e.v.Clone()}
}

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

func (e Elastic[T]) Values() []T {
	if !e.IsDefined() {
		return []T(nil)
	}
	opts := e.v.Value()
	vs := make([]T, len(opts))
	for i, opt := range opts {
		vs[i] = opt.Value()
	}
	return vs
}

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

func (u Elastic[T]) Unwrap() und.Und[option.Options[T]] {
	return u.v
}

func (e Elastic[T]) Map(f func(und.Und[option.Options[T]]) und.Und[option.Options[T]]) Elastic[T] {
	return Elastic[T]{
		v: f(e.v.Map(func(o option.Option[option.Option[option.Options[T]]]) option.Option[option.Option[option.Options[T]]] {
			if !o.IsNone() {
				return o
			}
			v := o.Value()
			if v.IsNone() {
				return o
			}
			vv := v.Value()
			return option.Some(option.Some(vv[:len(vv):len(vv)]))
		})),
	}
}
