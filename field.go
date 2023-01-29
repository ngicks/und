package undefinedablejson

import (
	"bytes"
	"encoding/json"
	"reflect"
)

type Nullable[T any] struct {
	v *T
}

func NewNullable[T any](v T) *Nullable[T] {
	return &Nullable[T]{
		v: &v,
	}
}

func NonNull[T any](v T) Nullable[T] {
	return Nullable[T]{
		v: &v,
	}
}

func Null[T any]() Nullable[T] {
	return Nullable[T]{}
}

func (n Nullable[T]) IsNull() bool {
	return n.v == nil
}

func (n Nullable[T]) Value() *T {
	return n.v
}

// IsZero reports if the inner value is zero value for T.
// Null n return false.
func (n Nullable[T]) IsZero() bool {
	if n.v == nil {
		return false
	}
	return reflect.ValueOf(*n.v).IsZero()
}

type Equality[T any] interface {
	Equal(T) bool
}

func (n Nullable[T]) Equal(other Nullable[T]) bool {
	if n.v == other.v {
		return true
	}
	if n.v == nil || other.v == nil {
		return n.v == nil && other.v == nil
	}

	// Try type assert first.
	// reflect.ValueOf escapes value into heap (currently).
	var asAny any = *n.v
	eq, ok := asAny.(Equality[T])
	if ok {
		return eq.Equal(*other.v)
	}
	// In case *T implements Equality[T].
	asAny = n.v
	eq, ok = asAny.(Equality[T])
	if ok {
		return eq.Equal(*other.v)
	}

	rv := reflect.Indirect(reflect.ValueOf(n.v))

	if !rv.Type().Comparable() {
		return false
	}

	otherRv := reflect.Indirect(reflect.ValueOf(other.v))
	return rv.Interface() == otherRv.Interface()
}

func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.v)
}

func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	// zero out
	n.v = nil

	data = bytes.TrimLeftFunc(data, func(r rune) bool { return r == ' ' || r == '\n' })
	if string(data) == `null` {
		return nil
	}

	var zero T
	n.v = &zero

	return json.Unmarshal(data, n.v)
}

func (f Nullable[T]) IsQuotable() bool {
	var t T
	return IsQuotable(reflect.TypeOf(t))
}

type Undefinedable[T any] struct {
	v *Nullable[T]
}

func Field[T any](v T) Undefinedable[T] {
	n := NonNull(v)
	return Undefinedable[T]{
		v: &n,
	}
}

func NullField[T any]() Undefinedable[T] {
	return Undefinedable[T]{
		v: &Nullable[T]{},
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
	if f.v == nil {
		return false
	}
	return f.v.IsNull()
}

func (f Undefinedable[T]) IsUndefined() bool {
	return f.v == nil
}

func (f Undefinedable[T]) Value() *T {
	if f.v == nil {
		return nil
	}
	return f.v.Value()
}

func (f Undefinedable[T]) IsZero() bool {
	if f.v == nil {
		return false
	}
	return f.v.IsZero()
}
func (f Undefinedable[T]) Equal(other Undefinedable[T]) bool {
	if f.v == other.v {
		return true
	}
	if f.v == nil || other.v == nil {
		return f.v == nil && other.v == nil
	}
	return f.v.Equal(*other.v)
}

func (f Undefinedable[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Value())
}

func (f *Undefinedable[T]) UnmarshalJSON(data []byte) error {
	f.v = &Nullable[T]{}
	return f.v.UnmarshalJSON(data)
}

func (f Undefinedable[T]) IsQuotable() bool {
	var t T
	return IsQuotable(reflect.TypeOf(t))
}
