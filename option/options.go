package option

import (
	"errors"
	"fmt"
	"slices"

	"github.com/ngicks/und/validate"
)

var (
	_ Equality[Options[any]] = Options[any]{}
	_ Cloner[Options[any]]   = Options[any]{}
	_ validate.ValidatorUnd  = Options[any]{}
	_ validate.CheckerUnd    = Options[any]{}
)

type Options[T any] []Option[T]

func (o Options[T]) Equal(opts Options[T]) bool {
	return slices.EqualFunc(
		o, opts,
		func(o1, o2 Option[T]) bool {
			return o1.Equal(o2)
		},
	)
}

func (o Options[T]) EqualFunc(opts Options[T], cmp func(i, j T) bool) bool {
	return slices.EqualFunc(
		o, opts,
		func(o1, o2 Option[T]) bool {
			return o1.EqualFunc(o2, cmp)
		},
	)
}

func (o Options[T]) Clone() Options[T] {
	if o == nil {
		return nil
	}
	opts := make(Options[T], len(o))
	var zero T
	if _, hasClone := any(zero).(Cloner[T]); hasClone {
		for i, v := range o {
			opts[i] = v.Map(func(v T) T { return any(v).(Cloner[T]).Clone() })
		}
	} else {
		copy(opts, o)
	}
	return opts
}

func (o Options[T]) ValidateUnd() error {
	for i, oo := range o {
		err := MapOrOption(oo, nil, func(t T) error {
			return validate.ValidateUnd(t)
		})
		if err != nil {
			if errors.Is(err, validate.ErrNotStruct) {
				// no point further inspect.
				// assumes T is not `any`.
				return nil
			}
			return fmt.Errorf("[%d].%w", i, err)
		}
	}
	return nil
}

func (o Options[T]) CheckUnd() error {
	var zero T
	err := validate.CheckUnd(zero)
	if errors.Is(err, validate.ErrNotStruct) {
		return nil
	}
	return err
}
