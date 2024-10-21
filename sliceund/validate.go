package sliceund

import "github.com/ngicks/und/validate"

func UndValidate[T validate.UndValidator](u Und[T]) error {
	if !u.IsDefined() {
		return nil
	}
	return u.Value().UndValidate()
}
