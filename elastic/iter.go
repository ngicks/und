package elastic

import (
	"iter"

	"github.com/ngicks/und/option"
)

// Iter returns an iterator over the internal option.
// If e is undefined, the iterator yields nothing, otherwise the internal option.
func (e Elastic[T]) Iter() iter.Seq[option.Option[option.Options[T]]] {
	return e.Unwrap().Iter()
}

// TODO: add more useful into-iterator kind methods here?
