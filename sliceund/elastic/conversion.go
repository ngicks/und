package elastic

import (
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/sliceund"
)

// FromElastic create a new Elastic[T] from non-slice version Elastic[T].
//
// The internal value of e is retained by the returned value.
func FromElastic[T any](e elastic.Elastic[T]) Elastic[T] {
	return FromUnd(sliceund.FromUnd(e.Unwrap()))
}

// Elastic converts e into non-slice version Elastic[T].
//
// The internal value of e is retained by the returned value.
func (e Elastic[T]) Elastic() elastic.Elastic[T] {
	return elastic.FromUnd(e.inner().Und())
}
