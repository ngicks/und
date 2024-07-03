package testcase_test

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	"gotest.tools/v3/assert"
)

type sqlNull[T any] interface {
	SqlNull() sql.Null[T]
	Value() T
}

func TestSqlNull(t *testing.T) {
	for _, constructor := range [](func(sql.Null[string]) sqlNull[string]){
		func(n sql.Null[string]) sqlNull[string] { return option.FromSqlNull(n) },
		func(n sql.Null[string]) sqlNull[string] { return und.FromSqlNull(n) },
		func(n sql.Null[string]) sqlNull[string] { return sliceund.FromSqlNull(n) },
	} {
		valid := constructor(sql.Null[string]{V: "foo", Valid: true})
		invalid := constructor(sql.Null[string]{})
		malformed := constructor(sql.Null[string]{V: "bar"})

		assert.Equal(t, valid.Value(), "foo")
		assert.Equal(t, invalid.Value(), "")
		assert.Equal(t, malformed.Value(), "")

		assert.Equal(t, valid.SqlNull(), sql.Null[string]{V: "foo", Valid: true})
		assert.Equal(t, invalid.SqlNull(), sql.Null[string]{})
		assert.Equal(t, malformed.SqlNull(), sql.Null[string]{})
	}
}

func TestSqlNullWrapper(t *testing.T) {
	var u und.SqlNull[string]
	var su sliceund.SqlNull[string]
	var o option.SqlNull[string]

	assert.NilError(t, u.Scan("foo"))
	assert.NilError(t, su.Scan("bar"))
	assert.NilError(t, o.Scan("baz"))

	var (
		v   driver.Value
		err error
	)
	v, err = u.Value()
	assert.NilError(t, err)
	assert.Equal(t, v, "foo")
	v, err = su.Value()
	assert.NilError(t, err)
	assert.Equal(t, v, "bar")
	v, err = o.Value()
	assert.NilError(t, err)
	assert.Equal(t, v, "baz")
}
