package option

import (
	"errors"

	"github.com/ngicks/und/validate"
)

var (
	_ validate.ValidatorUnd = Option[any]{}
	_ validate.CheckerUnd   = Option[any]{}
)

func (o Option[T]) ValidateUnd() error {
	return MapOrOption(o, nil, func(t T) error {
		err := validate.ValidateUnd(t)
		if errors.Is(err, validate.ErrNotStruct) {
			return nil
		}
		return err
	})
}

func (o Option[T]) CheckUnd() error {
	var zero T
	err := validate.CheckUnd(zero)
	if errors.Is(err, validate.ErrNotStruct) {
		return nil
	}
	return err
}
