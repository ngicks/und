package option

// FromOk converts conventional (t T, ok bool) into an option.
// The options is some if ok is true, none otherwise.
//
// For getting values from maps or slices, instead you may want to use [GetMap], [GetSlice] respectively.
func FromOk[T any](t T, ok bool) Option[T] {
	if ok {
		return Some(t)
	}
	return None[T]()
}

// Assert type-asserts v into T.
// If v's internal value is T then returns Some of that value,
// None otherwise.
func Assert[T any](v any) Option[T] {
	if t, ok := v.(T); ok {
		return Some(t)
	}
	return None[T]()
}

// GetMap gets a value associated with key.
// If key has a value, the Option is some wrapping the value.
// Otherwise it returns none Option.
func GetMap[M ~map[K]V, K comparable, V any](m M, key K) Option[V] {
	v, ok := m[key]
	if !ok {
		return None[V]()
	}
	return Some(v)
}

// GetMap gets a value associated with idx.
// If idx is within interval [0, len(s)), then the Option is some wrapping a value associated to the idx.
// Otherwise it returns none Option.
func GetSlice[S ~[]T, T any](s S, idx int) Option[T] {
	if idx < 0 || len(s) <= idx {
		return None[T]()
	}
	return Some(s[idx])
}
