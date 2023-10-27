package jsonfield

import (
	"github.com/ngicks/und/v2/nullable"
	"github.com/ngicks/und/v2/undefinedable"
)

// A default-ish type for fields of JSON object.
// It can be `undefined | null | T`
type JsonField[T any] struct {
	undefinedable.Undefinedable[nullable.Nullable[T]]
}

func Undefined[T any]() JsonField[T] {
	return JsonField[T]{}
}

func Null[T any]() JsonField[T] {
	return JsonField[T]{
		Undefinedable: undefinedable.Defined(nullable.Null[T]()),
	}
}

func Defined[T any](v T) JsonField[T] {
	return JsonField[T]{
		Undefinedable: undefinedable.Defined(nullable.NonNull[T](v)),
	}
}

// FromPointer converts *T into JsonField[T].
// If v is nil, it returns a null JsonField.
// Otherwise, v is copied by assignment.
func FromPointer[T any](v *T) JsonField[T] {
	if v == nil {
		return Null[T]()
	}
	return Defined[T](*v)
}

func (u JsonField[T]) IsNull() bool {
	if u.IsUndefined() {
		return false
	}
	return u.Undefinedable.Value().IsNull()
}

func (u JsonField[T]) IsNonNull() bool {
	if u.IsUndefined() {
		return false
	}
	return u.Undefinedable.Value().IsNonNull()
}

// Value returns value as T.
// If f is undefined or null, it returns zero value of T.
func (f JsonField[T]) Value() T {
	return f.Undefinedable.Value().Value()
}

// Plain returns value as **T, the conventional representation of optional value.
// nil means undefined. *nil is null.
func (f JsonField[T]) Plain() **T {
	if f.IsUndefined() {
		return nil
	}
	v := f.Undefinedable.Value().Plain()
	return &v
}

func (f JsonField[T]) Equal(other JsonField[T]) bool {
	if f.IsUndefined() || other.IsUndefined() {
		return f.IsUndefined() == other.IsUndefined()
	}

	return f.Undefinedable.Value().Equal(other.Undefinedable.Value())
}

func (f JsonField[T]) MarshalJSON() ([]byte, error) {
	return f.Undefinedable.Value().MarshalJSON()
}

func (f *JsonField[T]) UnmarshalJSON(data []byte) error {
	// json.Unmarshal would not call this if input json does not have the corresponding field.
	// f is a defined field at the moment this line is reached.
	err := f.Undefinedable.UnmarshalJSON(data)
	if err != nil {
		return err
	}
	if string(data) == "null" {
		// In case input data == "null", at this line f is undefined state.
		// revert that change.
		f.Undefinedable = undefinedable.Defined[nullable.Nullable[T]](f.Undefinedable.Value())
	}
	return nil
}
