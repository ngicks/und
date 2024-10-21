package elastic

import (
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/validate"
)

func UndValidate[T validate.UndValidator](e Elastic[T]) error {
	return option.MapOrOption(e.Unwrap().Unwrap(), nil, func(opt option.Option[option.Options[T]]) error {
		for _, opt := range opt.Value() {
			if err := option.MapOrOption(opt, nil, func(t T) error { return t.UndValidate() }); err != nil {
				return err
			}
		}
		return nil
	})
}
