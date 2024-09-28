package sliceund

import (
	"iter"

	"github.com/ngicks/und/option"
)

// Iter returns an iterator over the internal option.
// If u is undefined, the iterator yields nothing, otherwise the internal option.
func (u Und[T]) Iter() iter.Seq[option.Option[T]] {
	return func(yield func(option.Option[T]) bool) {
		if !u.IsUndefined() {
			yield(u[0])
		}
	}
}
