package elastic

import (
	"encoding/json"
	"encoding/xml"
	"log/slog"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
)

var (
	_ option.Equality[Elastic[any]] = Elastic[any]{}
	_ option.Cloner[Elastic[any]]   = Elastic[any]{}
	_ json.Marshaler                = Elastic[any]{}
	_ json.Unmarshaler              = (*Elastic[any])(nil)
	_ jsonv2.MarshalerV2            = Elastic[any]{}
	_ xml.Marshaler                 = Elastic[any]{}
	_ xml.Unmarshaler               = (*Elastic[any])(nil)
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
	_ slog.LogValuer = Elastic[any]{}
)

type Elastic[T any] sliceund.Und[option.Options[T]]

func Null[T any]() Elastic[T] {
	return Elastic[T](sliceund.Null[option.Options[T]]())
}

func Undefined[T any]() Elastic[T] {
	return Elastic[T](sliceund.Undefined[option.Options[T]]())
}

func FromOptions[T any, Opts ~[]option.Option[T]](options Opts) Elastic[T] {
	return Elastic[T](sliceund.Defined(option.Options[T](options)))
}

func (e Elastic[T]) inner() sliceund.Und[option.Options[T]] {
	return sliceund.Und[option.Options[T]](e)
}

func (e Elastic[T]) IsDefined() bool {
	return e.inner().IsDefined()
}

func (e Elastic[T]) IsNull() bool {
	return e.inner().IsNull()
}

func (e Elastic[T]) IsUndefined() bool {
	return e.inner().IsUndefined()
}

func (e Elastic[T]) Equal(other Elastic[T]) bool {
	return e.inner().Equal(other.inner())
}

func (e Elastic[T]) Clone() Elastic[T] {
	return Elastic[T](e.inner().Clone())
}

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

func (u Elastic[T]) Unwrap() sliceund.Und[option.Options[T]] {
	return sliceund.FromOption(u.inner().Unwrap())
}

func (e Elastic[T]) Map(f func(sliceund.Und[option.Options[T]]) sliceund.Und[option.Options[T]]) Elastic[T] {
	return Elastic[T](
		f(e.inner().Map(func(o option.Option[option.Option[option.Options[T]]]) option.Option[option.Option[option.Options[T]]] {
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
	)
}
