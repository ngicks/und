package elastic

import (
	"iter"
	"slices"

	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
)

// Iter returns an iterator over the internal option.
// If e is undefined, the iterator yields nothing, otherwise the internal option.
func (e Elastic[T]) Iter() iter.Seq[option.Option[option.Options[T]]] {
	return e.Unwrap().Iter()
}

func FromOptionSeq[T any](seq iter.Seq[option.Option[T]]) Elastic[T] {
	return Elastic[T](sliceund.Defined(option.Options[T](slices.Collect(seq))))
}

// TODO: add more useful into-iterator kind methods here?
