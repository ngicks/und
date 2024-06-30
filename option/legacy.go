package option

// Plain transforms o to *T, the plain conventional Go representation of an optional value.
// The value is copied by assignment before returned from Plain.
//
// Deprecated: use Pointer instead.
func (o Option[T]) Plain() *T {
	return o.Pointer()
}
