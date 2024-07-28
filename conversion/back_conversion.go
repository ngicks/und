package conversion

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
)

//TODO: revisit names

func UndNullishBack[T any](o option.Option[*struct{}]) und.Und[T] {
	if o.IsSome() {
		return und.Null[T]()
	}
	return und.Undefined[T]()
}

func UndNullishBackSlice[T any](o option.Option[*struct{}]) sliceund.Und[T] {
	if o.IsSome() {
		return sliceund.Null[T]()
	}
	return sliceund.Undefined[T]()
}

func MapOptionToUnd[T any](null bool, o option.Option[T]) und.Und[T] {
	if o.IsSome() {
		return und.Defined(o.Value())
	}
	if null {
		return und.Null[T]()
	}
	return und.Undefined[T]()
}

func MapOptionToUndSlice[T any](null bool, o option.Option[T]) sliceund.Und[T] {
	if o.IsSome() {
		return sliceund.Defined(o.Value())
	}
	if null {
		return sliceund.Null[T]()
	}
	return sliceund.Undefined[T]()
}

func UndNullishBackElastic[T any](o option.Option[*struct{}]) elastic.Elastic[T] {
	if o.IsSome() {
		return elastic.Null[T]()
	}
	return elastic.Undefined[T]()
}

func UndNullishBackElasticSlice[T any](o option.Option[*struct{}]) sliceelastic.Elastic[T] {
	if o.IsSome() {
		return sliceelastic.Null[T]()
	}
	return sliceelastic.Undefined[T]()
}

func MapOptionOptionToElastic[T any](null bool, o option.Option[[]option.Option[T]]) elastic.Elastic[T] {
	if o.IsSome() {
		return elastic.FromOptions(o.Value())
	}
	if null {
		return elastic.Null[T]()
	}
	return elastic.Undefined[T]()
}

func MapOptionOptionToElasticSlice[T any](null bool, o option.Option[[]option.Option[T]]) sliceelastic.Elastic[T] {
	if o.IsSome() {
		return sliceelastic.FromOptions(o.Value())
	}
	if null {
		return sliceelastic.Null[T]()
	}
	return sliceelastic.Undefined[T]()
}

func nonNullToUndMapper[T any](s []T) []option.Option[T] {
	r := make([]option.Option[T], len(s), cap(s)) // in case it matters
	for i, v := range s {
		r[i] = option.Some(v)
	}
	return r
}

func Nullify[T any](u und.Und[[]T]) und.Und[[]option.Option[T]] {
	return und.Map(u, nonNullToUndMapper)
}

func NullifySlice[Opts ~[]option.Option[T], T any](u sliceund.Und[[]T]) sliceund.Und[[]option.Option[T]] {
	return sliceund.Map(u, nonNullToUndMapper)
}

func wrapLen1Mapper[T any](v T) [1]T {
	return [1]T{v}
}

func WrapLen1[T any](u und.Und[T]) und.Und[[1]T] {
	return und.Map(u, wrapLen1Mapper)
}

func WrapLen1Slice[T any](u sliceund.Und[T]) sliceund.Und[[1]T] {
	return sliceund.Map(u, wrapLen1Mapper)
}
