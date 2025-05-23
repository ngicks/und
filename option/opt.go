package option

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"log/slog"
)

var (
	_ json.Marshaler   = Option[any]{}
	_ json.Unmarshaler = (*Option[any])(nil)
	_ xml.Marshaler    = Option[any]{}
	_ xml.Unmarshaler  = (*Option[any])(nil)
	_ slog.LogValuer   = Option[any]{}
)

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

func FromSqlNull[T any](v sql.Null[T]) Option[T] {
	if !v.Valid {
		return None[T]()
	}
	return Some(v.V)
}

// FromPointer converts *T into Option[T].
// If v is nil, it returns a none Option.
// Otherwise, it returns some Option whose value is the dereferenced v.
//
// If you need to keep t as pointer, use [WrapPointer] instead.
func FromPointer[T any](t *T) Option[T] {
	if t != nil {
		return Some(*t)
	}
	return None[T]()
}

// WrapPointer converts *T into Option[*T].
// The option is some if t is non nil, none otherwise.
//
// If you want t to be dereferenced, use [FromPointer] instead.
func WrapPointer[T any](t *T) Option[*T] {
	if t != nil {
		return Some(t)
	}
	return None[*T]()
}

func (o Option[T]) IsZero() bool {
	return o.IsNone()
}

func (o Option[T]) IsSome() bool {
	return o.some
}

// IsSomeAnd returns true if o is some and calling f with value of o returns true.
// Otherwise it returns false.
func (o Option[T]) IsSomeAnd(f func(T) bool) bool {
	if o.IsSome() {
		return f(o.Value())
	}
	return false
}

func (o Option[T]) IsNone() bool {
	return !o.IsSome()
}

// Value returns its internal as T.
// T would be zero value if o is None.
func (o Option[T]) Value() T {
	return o.v
}

func (o Option[T]) Get() (T, bool) {
	return o.Value(), o.IsSome()
}

// Pointer transforms o to *T, the plain conventional Go representation of an optional value.
// The value is copied by assignment before returned from Pointer.
func (o Option[T]) Pointer() *T {
	if o.IsNone() {
		return nil
	}
	t := o.v
	return &t
}

// CloneFunc clones o using the cloneT function.
func (o Option[T]) CloneFunc(cloneT func(T) T) Option[T] {
	return o.Map(func(t T) T {
		return cloneT(t)
	})
}

// EqualFunc tests o and other if both are Some or None.
// If their state does not match, it returns false immediately.
// If both have value, it tests equality of their values by cmp.
//
// If T is just a comparable type, use [Equal].
// If T is an implementor of interface { Equal(t T) bool }, e.g time.Time, use [EqualEqualer].
func (o Option[T]) EqualFunc(other Option[T], cmp func(i, j T) bool) bool {
	if !o.some || !other.some {
		return o.some == other.some
	}

	return cmp(o.v, other.v)
}

// Equal tests equality of l and r then returns true if they are equal, false otherwise
func Equal[T comparable](l, r Option[T]) bool {
	return l.EqualFunc(r, func(i, j T) bool { return i == j })
}

// EqualEqualer tests equality of l and r by calling Equal method implemented on l.
func EqualEqualer[T interface{ Equal(t T) bool }](l, r Option[T]) bool {
	return l.EqualFunc(r, func(i, j T) bool {
		return i.Equal(j)
	})
}

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.IsNone() {
		// same as bytes.Clone.
		return []byte(`null`), nil
	}
	return json.Marshal(o.v)
}

func (o *Option[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		o.some = false
		var zero T
		o.v = zero
		return nil
	}

	// not gonna call like json.Unmarshal(data, &o.v)
	// since it could be half-baked result if unmarshaling fails at some point.
	// o's state must stay valid.
	var v T
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	o.some = true
	o.v = v
	return nil
}

func (o Option[T]) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if o.IsNone() {
		return nil
	}
	return e.EncodeElement(o.Value(), start)
}

func (o *Option[T]) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var t T
	err := d.DecodeElement(&t, &start)
	if err != nil {
		return err
	}

	o.some = true
	o.v = t

	return nil
}

// LogValue implements slog.LogValuer
func (o Option[T]) LogValue() slog.Value {
	if o.IsNone() {
		return slog.AnyValue(nil)
	}
	return slog.AnyValue(o.Value())
}

func (o Option[T]) SqlNull() sql.Null[T] {
	if o.IsNone() {
		return sql.Null[T]{}
	}
	return sql.Null[T]{
		Valid: true,
		V:     o.Value(),
	}
}

// And returns u if o is some, otherwise None[T].
func (o Option[T]) And(u Option[T]) Option[T] {
	if o.IsSome() {
		return u
	} else {
		return None[T]()
	}
}

// AndThen calls f with value of o if o is some, otherwise returns None[T].
func (o Option[T]) AndThen(f func(x T) Option[T]) Option[T] {
	if o.IsSome() {
		return f(o.Value())
	} else {
		return None[T]()
	}
}

// Filter returns o if o is some and calling pred against o's value returns true.
// Otherwise it returns None[T].
func (o Option[T]) Filter(pred func(t T) bool) Option[T] {
	if o.IsSome() && pred(o.Value()) {
		return o
	}
	return None[T]()
}

// Flatten converts Option[Option[T]] into Option[T].
func Flatten[T any](o Option[Option[T]]) Option[T] {
	if o.IsNone() {
		return None[T]()
	}
	v := o.Value()
	if v.IsNone() {
		return None[T]()
	}
	return v
}

// Map returns Some[U] whose inner value is o's value mapped by f if o is Some.
// Otherwise it returns None[U].
func Map[T, U any](o Option[T], f func(T) U) Option[U] {
	if o.IsSome() {
		return Some(f(o.Value()))
	}
	return None[U]()
}

// Map returns Option[T] whose inner value is o's value mapped by f if o is some.
// Otherwise it returns None[T].
func (o Option[T]) Map(f func(v T) T) Option[T] {
	return Map(o, f)
}

// MapOr returns o's value applied by f if o is some.
// Otherwise it returns defaultValue.
func MapOr[T, U any](o Option[T], defaultValue U, f func(T) U) U {
	if o.IsNone() {
		return defaultValue
	}
	return f(o.Value())
}

// MapOr returns value o's value applied by f if o is some.
// Otherwise it returns defaultValue.
func (o Option[T]) MapOr(defaultValue T, f func(T) T) T {
	return MapOr(o, defaultValue, f)
}

// MapOrOpt is like [Option.MapOr] but wraps the returned value into some Option.
func (o Option[T]) MapOrOpt(defaultValue T, f func(T) T) Option[T] {
	return Some(MapOr(o, defaultValue, f))
}

// MapOrElse returns value o's value applied by f if o is some.
// Otherwise it returns a defaultFn result.
func MapOrElse[T, U any](o Option[T], defaultFn func() U, f func(T) U) U {
	if o.IsNone() {
		return defaultFn()
	}
	return f(o.Value())
}

// MapOrElse returns value o's value applied by f if o is some.
// Otherwise it returns a defaultFn result.
func (o Option[T]) MapOrElse(defaultFn func() T, f func(T) T) T {
	return MapOrElse(o, defaultFn, f)
}

// MapOrElseOpt is like [Option.MapOrElse] but wraps the returned value into some Option.
func (o Option[T]) MapOrElseOpt(defaultFn func() T, f func(T) T) Option[T] {
	return Some(MapOrElse(o, defaultFn, f))
}

// Or returns o if o is some, otherwise u.
func (o Option[T]) Or(u Option[T]) Option[T] {
	if o.IsSome() {
		return o
	} else {
		return u
	}
}

// OrElse returns o if o is some, otherwise calls f and returns the result.
func (o Option[T]) OrElse(f func() Option[T]) Option[T] {
	if o.IsSome() {
		return o
	} else {
		return f()
	}
}

// Xor returns o or u if either is some.
// If both are some or both none, it returns None[T].
func (o Option[T]) Xor(u Option[T]) Option[T] {
	if o.IsSome() && u.IsNone() {
		return o
	}
	if o.IsNone() && u.IsSome() {
		return u
	}
	return None[T]()
}
