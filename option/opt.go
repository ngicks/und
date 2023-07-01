package option

import (
	"reflect"

	"github.com/ngicks/und/serde"
)

type Equality[T any] interface {
	Equal(T) bool
}

// Option represents an optional value.
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

// Value returns its internal as T.
// T would be zero value if o is None.
func (o Option[T]) Value() T {
	return o.v
}

// Plain transforms o to *T, the plain conventional Go representation of an optional value.
// The value is copied by assignment before returned from Plain.
func (o Option[T]) Plain() *T {
	if o.IsNone() {
		return nil
	}
	t := o.v
	return &t
}

// Equal implements Equality[Option[T]].
//
// Equal tests if both o and other are Some or both are None.
// If both have value, it tests equality of their values.
//
// Option is hashable if T is hashable, so is the comparable.
// Equal only exists for cases where T needs a special Equal method (e.g. time.Time.)
//
// To test value equality correctly,
// T must implement Equality[T]
// or must be any of comparable, slice of comparable type or map of comparable value type.
// Other then Equal always returns false.
//
// Equal first checks if T implements Equality[T], then also for *T.
// If it does not, then Equal compares values through the `reflect` package.
//
// Be cautious that the comparison of comparable types may still panic at runtime.
// See doc comments for reflect.Type#Comparable or other related documents.
func (o Option[T]) Equal(other Option[T]) bool {
	if !o.some || !other.some {
		return o.some == other.some
	}

	return equal[T](o.v, other.v)
}

func equal[T any](t, u T) bool {
	// Try type assertion first.
	// The implemented interface has the precedence.
	//
	// Uses of reflect.ValueOf incur overhead of escaping values into the heap (at least currently).
	// Of course converting the value into any (interface{}) might also cause the overhead.
	// However considering the fact that it is used anywhere,
	// the chance of the compiler optimization is higher.

	// Check for T. Below *T is also checked but in case T is already a pointer type, when T = *U, *(*U) might not implement Equality.
	eq, ok := any(t).(Equality[T])
	if ok {
		return eq.Equal(u)
	}
	// check for *T so that we can find method implemented for *T not only ones for T.
	eq, ok = any(&t).(Equality[T])
	if ok {
		return eq.Equal(u)
	}

	rt, ru := reflect.ValueOf(t), reflect.ValueOf(u)

	k := rt.Type().Kind()
	switch k {
	case reflect.Slice, reflect.Map:
		if !rt.Type().Elem().Comparable() {
			return false
		}
		pt, pu := rt.UnsafePointer(), ru.UnsafePointer()
		if pt == pu {
			// might be both are nil.
			return true
		}
		if rt.Len() != ru.Len() {
			return false
		}
	}

	switch k {
	case reflect.Slice:
		for i := 0; i < rt.Len(); i++ {
			rti, rui := rt.Index(i), ru.Index(i)
			if rti.Interface() != rui.Interface() {
				return false
			}
		}
		return true

	case reflect.Map:
		iter := rt.MapRange()
		for iter.Next() {
			key := iter.Key()
			rti, rui := rt.MapIndex(key), ru.MapIndex(key)
			var zeroValue reflect.Value
			if rui == zeroValue {
				// The key does not exist in ru.
				return false
			}
			if rti.Interface() != rui.Interface() {
				return false
			}
		}
		return true
	}

	if !rt.Type().Comparable() {
		return false
	}

	return rt.Interface() == ru.Interface()
}

const nullStr = `null`

var nullByte = []byte(nullStr)

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if !o.some {
		// same as bytes.Clone.
		return append([]byte{}, nullByte...), nil
	}
	return serde.MarshalJSON(o.v)
}

func (o *Option[T]) UnmarshalJSON(data []byte) error {
	if string(data) == nullStr {
		o.some = false
		var zero T
		o.v = zero
		return nil
	}

	err := serde.UnmarshalJSON(data, &o.v)
	if err != nil {
		return err
	}
	o.some = true
	return nil
}

func (o Option[T]) Map(f func(v T) T) Option[T] {
	if o.some {
		return Option[T]{
			some: true,
			v:    f(o.v),
		}
	}
	return Option[T]{}
}
