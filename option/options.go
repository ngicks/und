package option

import (
	"errors"
	"fmt"
	"slices"

	"github.com/ngicks/und/validate"
)

var (
	_ validate.UndValidator = Options[any]{}
	_ validate.UndChecker   = Options[any]{}
)

type Options[T any] []Option[T]

// EqualFunc tests equality of l and r using an equality function cmp.
func (o Options[T]) EqualFunc(opts Options[T], cmp func(i, j T) bool) bool {
	return slices.EqualFunc(
		o, opts,
		func(o1, o2 Option[T]) bool {
			return o1.EqualFunc(o2, cmp)
		},
	)
}

// EqualOptions tests equality of l and r then returns true if they are equal, false otherwise
func EqualOptions[T comparable, S ~[]Option[T]](l, r S) bool {
	return Options[T](l).EqualFunc(Options[T](r), func(i, j T) bool { return i == j })
}

// EqualOptionsFunc tests equality of l and r using cmp then returns true if they are equal, false otherwise.
func EqualOptionsFunc[T any, S ~[]Option[T]](l, r S, cmp func(i, j T) bool) bool {
	return Options[T](l).EqualFunc(Options[T](r), cmp)
}

func (o Options[T]) CloneFunc(cloneT func(T) T) Options[T] {
	if o == nil { // in case it matters.
		return nil
	}
	opts := make(Options[T], len(o), cap(o)) // exact cap copying, in case it matters.
	for i, v := range o {
		opts[i] = v.CloneFunc(cloneT)
	}
	return opts
}

func CloneOptions[T comparable, S ~[]Option[T]](o S) Options[T] {
	if o == nil {
		return nil
	}
	opts := make(Options[T], len(o), cap(o)) // exact cap copying, in case it matters.
	copy(opts, o)
	return opts
}

func (o Options[T]) UndValidate() error {
	for i, oo := range o {
		err := MapOr(oo, nil, func(t T) error {
			return validate.UndValidate(t)
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

func (o Options[T]) UndCheck() error {
	var zero T
	err := validate.UndCheck(zero)
	if errors.Is(err, validate.ErrNotStruct) {
		return nil
	}
	return err
}
