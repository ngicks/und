package elastic

import (
	"github.com/ngicks/und/jsonfield"
	"github.com/ngicks/und/nullable"
	"github.com/ngicks/und/serde"
	"github.com/ngicks/und/undefinedable"
)

var _ serde.IsUndefineder = (*Elastic[any])(nil)

// An overly elastic type where it can be `undefined | (null | T) | (null | T)[]`.
type Elastic[T any] struct {
	undefinedable.Undefinedable[[]nullable.Nullable[T]]
}

func Undefined[T any]() Elastic[T] {
	return Elastic[T]{}
}

func Null[T any]() Elastic[T] {
	return Elastic[T]{
		Undefinedable: undefinedable.Defined([]nullable.Nullable[T]{nullable.Null[T]()}),
	}
}

func Defined[T any](v []nullable.Nullable[T]) Elastic[T] {
	return Elastic[T]{
		Undefinedable: undefinedable.Defined(v),
	}
}

// Single returns Elastic[T] that contains a single T value.
func Single[T any](v T) Elastic[T] {
	return Elastic[T]{
		Undefinedable: undefinedable.Defined([]nullable.Nullable[T]{nullable.NonNull[T](v)}),
	}
}

// Multiple returns Elastic[T] that contains multiple T values.
func Multiple[T any](v []T) Elastic[T] {
	values := make([]nullable.Nullable[T], len(v))
	for i, vv := range v {
		values[i] = nullable.NonNull(vv)
	}

	return Elastic[T]{
		Undefinedable: undefinedable.Defined(values),
	}
}

func (e Elastic[T]) Equal(other Elastic[T]) bool {
	if e.IsUndefined() || other.IsUndefined() {
		return e.IsUndefined() == other.IsUndefined()
	}
	if len(e.Value()) != len(other.Value()) {
		return false
	}
	v1, v2 := e.Value(), other.Value()
	for idx := range v1 {
		if !v1[idx].Equal(v2[idx]) {
			return false
		}
	}
	return true
}

func (e *Elastic[T]) IsSingle() bool {
	if e.IsUndefined() {
		return false
	}
	return len(e.Value()) == 1
}

func (e *Elastic[T]) IsMultiple() bool {
	if e.IsUndefined() {
		return false
	}
	return len(e.Value()) > 1
}

// IsNull returns true when e is a single null value,
// returns false otherwise.
func (e *Elastic[T]) IsNull() bool {
	if e.IsSingle() && e.Value()[0].IsNull() {
		return true
	}
	return false
}

// IsNullish returns true when it is considered empty,
// namely `undefined | null | null[]` or empty `T[]`.
// It returns false otherwise.
func (e *Elastic[T]) IsNullish() bool {
	for _, v := range e.Value() {
		if v.IsNonNull() {
			return false
		}
	}
	return true
}

// ValueSingle returns a first value of e.
// If e is nullish, namely `undefined | null | null[]` or empty `T[]`,
// it returns zero value of T.
func (e Elastic[T]) ValueSingle() T {
	if len(e.Value()) > 0 {
		return e.Value()[0].Value()
	}
	var zero T
	return zero
}

// PlainSingle returns a first value of e as *T,
// the plain conventional Go representation of an optional value.
// If e is undefined or has no value, then it returns nil.
func (e Elastic[T]) PlainSingle() *T {
	if len(e.Value()) > 0 {
		return e.Value()[0].Plain()
	}
	return nil
}

// ValueMultiple returns []T, replacing null value with zero value of T.
func (e Elastic[T]) ValueMultiple() []T {
	out := make([]T, len(e.Value()))
	for i, v := range e.Value() {
		out[i] = v.Value()
	}
	return out
}

// PlainMultiple returns slice of []*T.
// It returns always a non-nil even if e is undefined.
func (e Elastic[T]) PlainMultiple() []*T {
	out := make([]*T, len(e.Value()))
	for i, v := range e.Value() {
		out[i] = v.Plain()
	}
	return out
}

// First returns the first value as `undefined | null | T` type.
func (e Elastic[T]) First() jsonfield.JsonField[T] {
	if e.IsUndefined() || len(e.Value()) == 0 {
		return jsonfield.Undefined[T]()
	}
	if e.Value()[0].IsNull() {
		return jsonfield.Null[T]()
	} else {
		return jsonfield.Defined(e.Value()[0].Value())
	}
}

// MarshalJSON implements json.Marshaler.
//
// MarshalJSON encodes f into a json format.
// It always marshalls defined to be []T, undefined to be null.
func (f Elastic[T]) MarshalJSON() ([]byte, error) {
	// undefined should be skipped by serde.MarshalJSON.
	return serde.MarshalJSON(f.Value())
}

// UnmarshalJSON implements json.Unmarshaler.
//
// UnmarshalJSON accepts null, T and (null | T)[].
func (b *Elastic[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*b = Null[T]()
		return nil
	}

	var storedErr error
	if data[0] == '[' {
		err := b.Undefinedable.UnmarshalJSON(data)
		if err == nil {
			return nil
		}
		// in case of T = []U.
		storedErr = err
	}
	var single T
	err := serde.UnmarshalJSON(data, &single)
	if err != nil {
		if storedErr != nil {
			return storedErr
		} else {
			return err
		}
	}
	*b = Single[T](single)
	return nil
}
