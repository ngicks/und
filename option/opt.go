package option

import (
	"reflect"

	"github.com/ngicks/und/serde"
)

type Equality[T any] interface {
	Equal(T) bool
}

type Option[T any] struct {
	Some bool
	V    T
}

func SomeOpt[T any](v T) Option[T] {
	return Option[T]{
		Some: true,
		V:    v,
	}
}

func NoneOpt[T any]() Option[T] {
	return Option[T]{}
}

func (o Option[T]) Value() T {
	return o.V
}

func (o Option[T]) Equal(other Option[T]) bool {
	if !o.Some || !other.Some {
		return o.Some == other.Some
	}

	// Try type assert first.
	// reflect.ValueOf escapes value into heap (currently).

	// Check for T. Below *T is also checked but in case T is already a pointer type, when T = *U, *(*U) might not implement Equality.
	eq, ok := any(o.V).(Equality[T])
	if ok {
		return eq.Equal(other.V)
	}
	// check for *T so that we can find method implemented for *T not only ones for T.
	eq, ok = any(&o.V).(Equality[T])
	if ok {
		return eq.Equal(other.V)
	}

	rv := reflect.ValueOf(o.V)

	if !rv.Type().Comparable() {
		return false
	}

	otherRv := reflect.ValueOf(other.V)
	return rv.Interface() == otherRv.Interface()
}

const nullStr = `null`

var nullByte = []byte(nullStr)

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if !o.Some {
		// same as bytes.Clone.
		return append([]byte{}, nullByte...), nil
	}
	return serde.MarshalJSON(o.V)
}

func (o *Option[T]) UnmarshalJSON(data []byte) error {
	if string(data) == nullStr {
		o.Some = false
		return nil
	}

	o.Some = true
	return serde.UnmarshalJSON(data, &o.V)
}

func (o Option[T]) Map(f func(v T) T) Option[T] {
	if o.Some {
		return Option[T]{
			Some: true,
			V:    (f(o.V)),
		}
	}
	return Option[T]{}
}
