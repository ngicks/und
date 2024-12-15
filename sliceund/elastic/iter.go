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
	options := option.Options[T](slices.Collect(seq))
	if options == nil {
		options = make(option.Options[T], 0)
	}
	return Elastic[T](sliceund.Defined(options))
}

// TODO: add more useful into-iterator kind methods here?
