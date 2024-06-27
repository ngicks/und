package option

import "slices"

type Options[T any] []Option[T]

func (o Options[T]) Equal(opts Options[T]) bool {
	return slices.EqualFunc(
		o, opts,
		func(o1, o2 Option[T]) bool {
			return o1.Equal(o2)
		},
	)
}

func (o Options[T]) Clone() Options[T] {
	if o == nil {
		return nil
	}
	opts := make(Options[T], len(o))
	copy(opts, o)
	return opts
}
