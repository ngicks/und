package und

import (
	"database/sql"
	"database/sql/driver"
)

var (
	_ sql.Scanner   = (*SqlNull[any])(nil)
	_ driver.Valuer = SqlNull[any]{}
)

// SqlNull[T] adapts Und[T] to sql.Scanner and driver.Valuer.
type SqlNull[T any] struct {
	Und[T]
}

// Scan implements sql.Scanner.
//
// If T or *T implements sql.Scanner, the implementation is used.
// Otherwise, SqlNull[T] falls back to sql.Null[T] as sql.Scanner.
func (n *SqlNull[T]) Scan(src any) error {
	if src == nil {
		n.Und = Null[T]()
		return nil
	}

	var (
		t       T
		scanner sql.Scanner
		err     error
	)
	scanner, _ = any(t).(sql.Scanner)
	if scanner == nil {
		scanner, _ = any(&t).(sql.Scanner)
	}
	if scanner != nil {
		err = scanner.Scan(src)
		if err != nil {
			return err
		}
		n.Und = Defined(t)
		return nil
	}

	var null sql.Null[T]
	err = null.Scan(src)
	if err != nil {
		return err
	}
	n.Und = FromSqlNull(null)
	return nil
}

// Value implements driver.Valuer.
//
// If T or *T implements driver.Valuer, the implementation is used.
// In this respect, T should not be a pointer type or Und[T] should not store nil value.
// Otherwise, SqlNull[T] falls back to sql.Null[T] as driver.Valuer.
func (n SqlNull[T]) Value() (driver.Value, error) {
	if !n.Und.IsDefined() {
		return nil, nil
	}

	v := n.Und.Value()
	valuer, _ := any(v).(driver.Valuer)
	if valuer == nil {
		valuer, _ = any(&v).(driver.Valuer)
	}
	if valuer != nil {
		return valuer.Value()
	}

	return n.Und.SqlNull().Value()
}
