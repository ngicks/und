package nullable

import "github.com/ngicks/und/v2/option"

// A type represents `null | T`.
type Nullable[T any] struct {
	option.Option[T]
}

func Null[T any]() Nullable[T] {
	return Nullable[T]{}
}

func NonNull[T any](v T) Nullable[T] {
	return Nullable[T]{
		Option: option.Some(v),
	}
}

func (n Nullable[T]) Equal(other Nullable[T]) bool {
	return n.Option.Equal(other.Option)
}

func (n Nullable[T]) IsNull() bool {
	return !n.IsNonNull()
}

func (n Nullable[T]) IsNonNull() bool {
	return n.IsSome()
}
