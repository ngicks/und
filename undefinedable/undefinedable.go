package undefinedable

import (
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/serde"
)

var _ serde.IsUndefineder = (*Undefinedable[any])(nil)

type Undefinedable[T any] struct {
	option.Option[T]
}

func Undefined[T any]() Undefinedable[T] {
	return Undefinedable[T]{}
}

func Defined[T any](v T) Undefinedable[T] {
	return Undefinedable[T]{
		Option: option.Some(v),
	}
}

// FromPointer converts *T into Undefinedable[T].
// If v is nil, it returns an undefined Undefinedable.
// Otherwise, v is copied by assignment.
func FromPointer[T any](v *T) Undefinedable[T] {
	if v == nil {
		return Undefined[T]()
	}
	return Defined[T](*v)
}

func (u Undefinedable[T]) IsUndefined() bool {
	return !u.IsDefined()
}

func (u Undefinedable[T]) IsDefined() bool {
	return u.IsSome()
}

func (f Undefinedable[T]) Equal(other Undefinedable[T]) bool {
	return f.Option.Equal(other.Option)
}
