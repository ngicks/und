package validate

type optionLite[T any] struct {
	some bool
	t    T
}

func some[T any](t T) optionLite[T] {
	return optionLite[T]{
		some: true,
		t:    t,
	}
}

func (o optionLite[T]) Value() T {
	return o.t
}

func (o optionLite[T]) IsSome() bool {
	return o.some
}
func (o optionLite[T]) IsNone() bool {
	return !o.some
}
func (o optionLite[T]) IsSomeAnd(f func(s T) bool) bool {
	if o.IsNone() {
		return false
	}
	return f(o.t)
}

func (o optionLite[T]) Map(f func(v T) T) optionLite[T] {
	if o.IsSome() {
		return some(f(o.t))
	}
	return o
}
func (o optionLite[T]) Or(u optionLite[T]) optionLite[T] {
	if o.IsSome() {
		return o
	}
	return u
}
