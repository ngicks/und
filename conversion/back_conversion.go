package conversion

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
)

func NullishUnd[T any](o option.Option[*struct{}]) und.Und[T] {
	if o.IsSome() {
		return und.Null[T]()
	}
	return und.Undefined[T]()
}

func NullishUndSlice[T any](o option.Option[*struct{}]) sliceund.Und[T] {
	if o.IsSome() {
		return sliceund.Null[T]()
	}
	return sliceund.Undefined[T]()
}

func OptionUnd[T any](null bool, o option.Option[T]) und.Und[T] {
	if o.IsSome() {
		return und.Defined(o.Value())
	}
	if null {
		return und.Null[T]()
	}
	return und.Undefined[T]()
}

func OptionUndSlice[T any](null bool, o option.Option[T]) sliceund.Und[T] {
	if o.IsSome() {
		return sliceund.Defined(o.Value())
	}
	if null {
		return sliceund.Null[T]()
	}
	return sliceund.Undefined[T]()
}

func NullishElastic[T any](o option.Option[*struct{}]) elastic.Elastic[T] {
	if o.IsSome() {
		return elastic.Null[T]()
	}
	return elastic.Undefined[T]()
}

func NullishElasticSlice[T any](o option.Option[*struct{}]) sliceelastic.Elastic[T] {
	if o.IsSome() {
		return sliceelastic.Null[T]()
	}
	return sliceelastic.Undefined[T]()
}

func OptionOptionElastic[T any](null bool, o option.Option[[]option.Option[T]]) elastic.Elastic[T] {
	if o.IsSome() {
		return elastic.FromOptions(o.Value()...)
	}
	if null {
		return elastic.Null[T]()
	}
	return elastic.Undefined[T]()
}

func OptionOptionElasticSlice[T any](null bool, o option.Option[[]option.Option[T]]) sliceelastic.Elastic[T] {
	if o.IsSome() {
		return sliceelastic.FromOptions(o.Value()...)
	}
	if null {
		return sliceelastic.Null[T]()
	}
	return sliceelastic.Undefined[T]()
}

func nonNullToUndMapper[T any](o option.Option[[]T]) option.Option[[]option.Option[T]] {
	return option.Map(o, func(s []T) []option.Option[T] {
		r := make([]option.Option[T], len(s), cap(s)) // in case it matters
		for i, v := range s {
			r[i] = option.Some(v)
		}
		return r
	})
}

func Nullify[T any](u und.Und[[]T]) und.Und[[]option.Option[T]] {
	return und.FromOption(option.Map(u.Unwrap(), nonNullToUndMapper))
}

func NullifySlice[Opts ~[]option.Option[T], T any](u sliceund.Und[[]T]) sliceund.Und[[]option.Option[T]] {
	return sliceund.FromOption(option.Map(u.Unwrap(), nonNullToUndMapper))
}

func wrapLen1Mapper[T any](o option.Option[T]) option.Option[[1]T] {
	return option.Map(o, func(t T) [1]T {
		return [1]T{t}
	})
}

func WrapLen1[T any](u und.Und[T]) und.Und[[1]T] {
	return und.FromOption(option.Map(u.Unwrap(), wrapLen1Mapper))
}

func WrapLen1Slice[T any](u sliceund.Und[T]) sliceund.Und[[1]T] {
	return sliceund.FromOption(option.Map(u.Unwrap(), wrapLen1Mapper))
}
