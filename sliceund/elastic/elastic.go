package elastic

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
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
		if len(vs) > 0 {
			v := vs[0].Value()
			return &v
		}
	}
	return nil
}

func (e Elastic[T]) Pointers() []*T {
	if !e.IsDefined() {
		return []*T(nil)
	}
	opts := e.inner().Value()
	ptrs := make([]*T, len(opts))
	for i, opt := range opts {
		ptrs[i] = opt.Pointer()
	}
	return ptrs
}

func (u Elastic[T]) Unwrap() und.Und[option.Options[T]] {
	return und.FromOption(u.inner().Unwrap())
}

func (e Elastic[T]) Map(f func(sliceund.Und[option.Options[T]]) sliceund.Und[option.Options[T]]) Elastic[T] {
	return Elastic[T](f(e.inner()))
}
