package undefinedablejson

import (
	"encoding/json"
	"reflect"
)

var nullByte = []byte(`null`)

type Equality[T any] interface {
	Equal(T) bool
}

type Nullable[T any] struct {
	isNonNull bool
	v         T
}

func NewNullable[T any](v T) *Nullable[T] {
	return &Nullable[T]{
		v: v,
	}
}

func NonNull[T any](v T) Nullable[T] {
	return Nullable[T]{
		isNonNull: true,
		v:         v,
	}
}

func Null[T any]() Nullable[T] {
	return Nullable[T]{}
}

func (n Nullable[T]) IsNull() bool {
	return !n.isNonNull
}

func (n Nullable[T]) Value() *T {
	if !n.isNonNull {
		return nil
	}
	return &n.v
}

func (n Nullable[T]) Equal(other Nullable[T]) bool {
	if !n.isNonNull || !other.isNonNull {
		return n.isNonNull == other.isNonNull
	}

	// Try type assert first.
	// reflect.ValueOf escapes value into heap (currently).
	//
	// check for *T so that we can find method implemented for *T not only ones for T.
	eq, ok := any(&n.v).(Equality[T])
	if ok {
		return eq.Equal(other.v)
	}

	rv := reflect.ValueOf(n.v)

	if !rv.Type().Comparable() {
		return false
	}

	otherRv := reflect.ValueOf(other.v)
	return rv.Interface() == otherRv.Interface()
}

func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if !n.isNonNull {
		return nullByte, nil
	}
	return json.Marshal(n.v)
}

func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	if string(data) == string(nullByte) {
		n.isNonNull = false
		return nil
	}

	n.isNonNull = true
	return json.Unmarshal(data, &n.v)
}

func (f Nullable[T]) IsQuotable() bool {
	var t T
	return IsQuotable(reflect.TypeOf(t))
}

type Undefinedable[T any] struct {
	isDefined bool
	v         Nullable[T]
}

func Field[T any](v T) Undefinedable[T] {
	return Undefinedable[T]{
		isDefined: true,
		v:         NonNull(v),
	}
}

func NullField[T any]() Undefinedable[T] {
	return Undefinedable[T]{
		isDefined: true,
		v:         Nullable[T]{},
	}
}

func UndefinedField[T any]() Undefinedable[T] {
	return Undefinedable[T]{}
}

func NewField[T any](v T) *Undefinedable[T] {
	f := Field(v)
	return &f
}

func (f Undefinedable[T]) IsNull() bool {
	if !f.isDefined {
		return false
	}
	return f.v.IsNull()
}

func (f Undefinedable[T]) IsUndefined() bool {
	return !f.isDefined
}

func (f Undefinedable[T]) Value() *T {
	if !f.isDefined {
		return nil
	}
	return f.v.Value()
}

func (f Undefinedable[T]) Equal(other Undefinedable[T]) bool {
	if !f.isDefined || !other.isDefined {
		return f.isDefined == other.isDefined
	}

	return f.v.Equal(other.v)
}

func (f Undefinedable[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Value())
}

func (f *Undefinedable[T]) UnmarshalJSON(data []byte) error {
	// json.Unmarshal would not call this if input json has corresponding field.
	// So at the moment this line is reached, f is q defined field.
	f.isDefined = true
	return f.v.UnmarshalJSON(data)
}

func (f Undefinedable[T]) IsQuotable() bool {
	var t T
	return IsQuotable(reflect.TypeOf(t))
}
