package conversion

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/internal/structtag"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
)

type UndLike[T any] interface {
	structtag.UndLike
	Value() T
}

type ElasticLike[T any] interface {
	structtag.ElasticLike
	Value() T
	Values() []T
	Option() option.Option[T]
	Options() []option.Option[T]
}

func UndNullish[T structtag.UndLike](t T) option.Option[*struct{}] {
	if t.IsNull() {
		return option.Some[*struct{}](nil)
	}
	return option.None[*struct{}]()
}

func UnwrapElastic[T any](e elastic.Elastic[T]) und.Und[[]option.Option[T]] {
	switch {
	case e.IsUndefined():
		return und.Undefined[[]option.Option[T]]()
	case e.IsNull():
		return und.Null[[]option.Option[T]]()
	default:
		return und.Defined(e.Options())
	}
}

func UnwrapElasticSlice[T any](e sliceelastic.Elastic[T]) sliceund.Und[[]option.Option[T]] {
	switch {
	case e.IsUndefined():
		return sliceund.Undefined[[]option.Option[T]]()
	case e.IsNull():
		return sliceund.Null[[]option.Option[T]]()
	default:
		return sliceund.Defined(e.Options())
	}
}

func LenNAtMost[Opts ~[]T, T any](n int, u und.Und[Opts]) und.Und[Opts] {
	return und.Map(u, func(s Opts) Opts {
		s2 := make(Opts, n)
		s2 = s2[:copy(s2, s)]
		return s2
	})
}

func LenNAtMostSlice[Opts ~[]T, T any](n int, u sliceund.Und[Opts]) sliceund.Und[Opts] {
	return sliceund.Map(u, func(s Opts) Opts {
		s2 := make(Opts, n)
		s2 = s2[:copy(s2, s)]
		return s2
	})
}

func LenNAtLeast[Opts ~[]T, T any](n int, u und.Und[Opts]) und.Und[Opts] {
	return und.Map(u, func(s Opts) Opts {
		capacity := n
		if len(s) > capacity {
			capacity = len(s)
		}
		s2 := make(Opts, len(s), capacity)
		copy(s2, s)
		if len(s2) < n {
			s2 = s2[:n]
		}
		return s2
	})
}

func LenNAtLeastSlice[Opts ~[]T, T any](n int, u sliceund.Und[Opts]) sliceund.Und[Opts] {
	return sliceund.Map(u, func(s Opts) Opts {
		capacity := n
		if len(s) > capacity {
			capacity = len(s)
		}
		s2 := make(Opts, len(s), capacity)
		copy(s2, s)
		if len(s2) < n {
			s2 = s2[:n]
		}
		return s2
	})
}

func NonNull[Opts ~[]option.Option[T], T any](u und.Und[Opts]) und.Und[[]T] {
	return und.Map(u, func(s Opts) []T {
		r := make([]T, len(s), cap(s)) // in case it matters
		for i, v := range s {
			r[i] = v.Value()
		}
		return r
	})
}

func NonNullSlice[Opts ~[]option.Option[T], T any](u sliceund.Und[Opts]) sliceund.Und[[]T] {
	return sliceund.Map(u, func(s Opts) []T {
		r := make([]T, len(s), cap(s))
		for i, v := range s {
			r[i] = v.Value()
		}
		return r
	})
}

func UnwrapLen1[T any](u und.Und[[1]T]) und.Und[T] {
	return und.Map(u, func(v [1]T) T {
		return v[0]
	})
}

func UnwrapLen1Slice[T any](u sliceund.Und[[1]T]) sliceund.Und[T] {
	return sliceund.Map(u, func(v [1]T) T {
		return v[0]
	})
}

func init() {
	e := elastic.FromValue("foo")
	LenNAtLeast(5, UnwrapElastic(e))

	u := und.Defined([]option.Option[string]{option.Some("foo")})
	_ = und.Map(u, func(s []option.Option[string]) (r [2]option.Option[string]) {
		copy(r[:], s)
		return
	})
}
