package option

import (
	"errors"

	"github.com/ngicks/und/validate"
)

var (
	_ validate.UndValidator = Option[any]{}
	_ validate.UndChecker   = Option[any]{}
)

func (o Option[T]) UndValidate() error {
	return MapOr(o, nil, func(t T) error {
		err := validate.UndValidate(t)
		if errors.Is(err, validate.ErrNotStruct) {
			return nil
		}
		return err
	})
}

func (o Option[T]) UndCheck() error {
	var zero T
	err := validate.UndCheck(zero)
	if errors.Is(err, validate.ErrNotStruct) {
		return nil
	}
	return err
}

func UndValidate[T validate.UndValidator](o Option[T]) error {
	if o.IsNone() {
		return nil
	}
	return o.Value().UndValidate()
}
