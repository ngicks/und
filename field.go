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
	Option[T]
}

func (n Nullable[T]) IsNull() bool {
	return n.IsNone()
}

func (n Nullable[T]) IsNonNull() bool {
	return n.IsSome()
}

func NonNull[T any](v T) Nullable[T] {
	return Nullable[T]{
		Option: Option[T]{
			some: true,
			v:    v,
		},
	}
}

func Null[T any]() Nullable[T] {
	return Nullable[T]{}
}

func (n Nullable[T]) Equal(other Nullable[T]) bool {
	return n.Option.Equal(other.Option)
}

type Undefinedable[T any] struct {
	Option[Nullable[T]]
}

func Field[T any](v T) Undefinedable[T] {
	return Undefinedable[T]{
		Option: Option[Nullable[T]]{
			some: true,
			v: Nullable[T]{
				Option: Option[T]{
					some: true,
					v:    v,
				},
			},
		},
	}
}

func NullField[T any]() Undefinedable[T] {
	return Undefinedable[T]{
		Option: Option[Nullable[T]]{
			some: true,
			v:    Nullable[T]{},
		},
	}
}

func UndefinedField[T any]() Undefinedable[T] {
	return Undefinedable[T]{}
}

func (u Undefinedable[T]) IsUndefined() bool {
	return u.IsNone()
}

func (u Undefinedable[T]) IsDefined() bool {
	return u.IsSome()
}

func (u Undefinedable[T]) IsNull() bool {
	if u.IsUndefined() {
		return false
	}
	return u.v.IsNull()
}

func (u Undefinedable[T]) IsNonNull() bool {
	if u.IsUndefined() {
		return false
	}
	return u.v.IsNonNull()
}

func (f Undefinedable[T]) Value() *T {
	if f.IsUndefined() {
		return nil
	}
	return f.v.Value()
}

func (f Undefinedable[T]) Equal(other Undefinedable[T]) bool {
	if f.IsUndefined() || other.IsUndefined() {
		return f.some == other.some
	}

	return f.v.Equal(other.v)
}

func (f Undefinedable[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Value())
}

func (f *Undefinedable[T]) UnmarshalJSON(data []byte) error {
	// json.Unmarshal would not call this if input json has a corresponding field.
	// So at the moment this line is reached, f is a defined field.
	f.some = true
	return f.v.UnmarshalJSON(data)
}

func (u Undefinedable[T]) IsQuotable() bool {
	return u.v.IsQuotable()
}

type Option[T any] struct {
	some bool
	v    T
}

func Some[T any](v T) Option[T] {
	return Option[T]{
		some: true,
		v:    v,
	}
}

func None[T any]() Option[T] {
	return Option[T]{}
}

func (o Option[T]) IsSome() bool {
	return o.some
}

func (o Option[T]) IsNone() bool {
	return !o.IsSome()
}

func (o Option[T]) Value() *T {
	if !o.some {
		return nil
	}
	return &o.v
}

func (o Option[T]) Equal(other Option[T]) bool {
	if o.IsNone() || other.IsNone() {
		return o.some == other.some
	}

	// Try type assert first.
	// reflect.ValueOf escapes value into heap (currently).
	//
	// check for *T so that we can find method implemented for *T not only ones for T.
	eq, ok := any(&o.v).(Equality[T])
	if ok {
		return eq.Equal(other.v)
	}

	rv := reflect.ValueOf(o.v)

	if !rv.Type().Comparable() {
		return false
	}

	otherRv := reflect.ValueOf(other.v)
	return rv.Interface() == otherRv.Interface()
}

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.IsNone() {
		return nullByte, nil
	}
	return json.Marshal(o.v)
}

func (o *Option[T]) UnmarshalJSON(data []byte) error {
	if string(data) == string(nullByte) {
		o.some = false
		return nil
	}

	o.some = true
	return json.Unmarshal(data, &o.v)
}

func (Option[T]) IsQuotable() bool {
	var t T
	return IsQuotable(reflect.TypeOf(t))
}

func (o Option[T]) Map(f func(v T) T) Option[T] {
	if o.IsSome() {
		return Some(f(o.v))
	}
	return None[T]()
}
