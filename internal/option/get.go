package option

func GetMap[M ~map[K]V, K comparable, V any](m M, key K) Option[V] {
	v, ok := m[key]
	if !ok {
		return None[V]()
	}
	return Some(v)
}

func GetSlice[S ~[]T, T any](s S, idx int) Option[T] {
	if idx < 0 || len(s) <= idx {
		return None[T]()
	}
	return Some(s[idx])
}
