package elastic

// Legacy methods

// Deprecated: use Value.
func (e Elastic[T]) ValueSingle() T {
	return e.Value()
}

// Deprecated: use Values.
func (e Elastic[T]) ValueMultiple() []T {
	return e.Values()
}

// Deprecated: use Pointer.
func (e Elastic[T]) PlainSingle() *T {
	return e.Pointer()
}

// Deprecated: use Pointers
func (e Elastic[T]) PlainMultiple() []*T {
	return e.Pointers()
}
