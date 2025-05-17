package elastic

import (
	"strconv"

	"github.com/ngicks/und/option"
	"github.com/ngicks/und/validate"
)

func UndValidate[T validate.UndValidator](e Elastic[T], skipIf func(int, option.Option[T]) bool) error {
	return option.MapOr(e.Unwrap().Unwrap(), nil, func(opt option.Option[option.Options[T]]) error {
		for i, opt := range opt.Value() {
			if skipIf != nil && skipIf(i, opt) {
				continue
			}
			if err := option.MapOr(opt, nil, func(t T) error { return t.UndValidate() }); err != nil {
				return validate.AppendValidationErrorIndex(err, strconv.FormatInt(int64(i), 10))
			}
		}
		return nil
	})
}
