package undefinedable

import (
	"github.com/ngicks/und/nullable"
	"github.com/ngicks/und/option"
)

type Undefinedable[T any] struct {
	option.Option[nullable.Nullable[T]]
}

func Undefined[T any]() Undefinedable[T] {
	return Undefinedable[T]{}
}

func Null[T any]() Undefinedable[T] {
	return Undefinedable[T]{
		Option: option.Option[nullable.Nullable[T]]{
			Some: true,
			V:    nullable.Null[T](),
		},
	}
}

func Defined[T any](v T) Undefinedable[T] {
	return Undefinedable[T]{
		Option: option.Option[nullable.Nullable[T]]{
			Some: true,
			V:    nullable.NonNull(v),
		},
	}
}

func (u Undefinedable[T]) IsUndefined() bool {
	return !u.IsDefined()
}

func (u Undefinedable[T]) IsDefined() bool {
	return u.Option.Some
}

func (u Undefinedable[T]) IsNull() bool {
	if u.IsUndefined() {
		return false
	}
	return u.Option.V.IsNull()
}

func (u Undefinedable[T]) IsNonNull() bool {
	if u.IsUndefined() {
		return false
	}
	return u.Option.V.IsNonNull()
}

func (f Undefinedable[T]) Value() T {
	return f.Option.V.Option.V
}

func (f Undefinedable[T]) Equal(other Undefinedable[T]) bool {
	if f.IsUndefined() || other.IsUndefined() {
		return f.IsUndefined() == other.IsUndefined()
	}

	return f.Option.Value().Equal(other.Option.Value())
}

func (f Undefinedable[T]) MarshalJSON() ([]byte, error) {
	return f.Option.V.MarshalJSON()
}

func (f *Undefinedable[T]) UnmarshalJSON(data []byte) error {
	// json.Unmarshal would not call this if input json does not have the corresponding field.
	// So at the moment this line is reached, f is a defined field.
	f.Option.Some = true
	return f.Option.V.UnmarshalJSON(data)
}
