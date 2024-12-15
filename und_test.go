package und_test

import (
	"database/sql"
	"testing"

	"github.com/ngicks/und"
	"github.com/ngicks/und/internal/testcase"
	"github.com/ngicks/und/option"
	"gotest.tools/v3/assert"
)

func TestUnd(t *testing.T) {
	testcase.TestUnd_non_addressable(
		t,
		und.Defined[int](155),
		und.Null[int](),
		und.Undefined[int](),
		155,
		"155",
	)
}

// Tests for New-like function, e.g. FromPointer, WrapPointer
func TestUnd_new_functions(t *testing.T) {
	num := 15
	t.Run("FromPointer", func(t *testing.T) {
		fromNonNil := und.FromPointer(&num)
		assert.Equal(t, 15, fromNonNil.Value())
		assert.Equal(t, false, fromNonNil.IsUndefined())
		assert.Equal(t, false, fromNonNil.IsNull())
		assert.Equal(t, true, fromNonNil.IsDefined())
		fromNil := und.FromPointer((*int)(nil))
		assert.Equal(t, 0, fromNil.Value())
		assert.Equal(t, true, fromNil.IsUndefined())
		assert.Equal(t, false, fromNil.IsNull())
		assert.Equal(t, false, fromNil.IsDefined())
	})
	t.Run("WrapPointer", func(t *testing.T) {
		fromNonNil := und.WrapPointer(&num)
		assert.Equal(t, &num, fromNonNil.Value())
		assert.Equal(t, false, fromNonNil.IsUndefined())
		assert.Equal(t, false, fromNonNil.IsNull())
		assert.Equal(t, true, fromNonNil.IsDefined())
		fromNil := und.WrapPointer((*int)(nil))
		assert.Equal(t, (*int)(nil), fromNil.Value())
		assert.Equal(t, true, fromNil.IsUndefined())
		assert.Equal(t, false, fromNil.IsNull())
		assert.Equal(t, false, fromNil.IsDefined())
	})
	t.Run("FromOption", func(t *testing.T) {
		undefined := und.FromOption(option.None[option.Option[int]]())
		assert.Equal(t, 0, undefined.Value())
		assert.Equal(t, true, undefined.IsUndefined())
		assert.Equal(t, false, undefined.IsNull())
		assert.Equal(t, false, undefined.IsDefined())
		null := und.FromOption(option.Some(option.None[int]()))
		assert.Equal(t, 0, null.Value())
		assert.Equal(t, false, null.IsUndefined())
		assert.Equal(t, true, null.IsNull())
		assert.Equal(t, false, null.IsDefined())
		defined := und.FromOption(option.Some(option.Some(num)))
		assert.Equal(t, num, defined.Value())
		assert.Equal(t, false, defined.IsUndefined())
		assert.Equal(t, false, defined.IsNull())
		assert.Equal(t, true, defined.IsDefined())
	})
	t.Run("FromSqlNull", func(t *testing.T) {
		null := und.FromSqlNull(sql.Null[int]{Valid: false, V: 15})
		assert.Equal(t, 0, null.Value())
		assert.Equal(t, false, null.IsUndefined())
		assert.Equal(t, true, null.IsNull())
		assert.Equal(t, false, null.IsDefined())

		defined := und.FromSqlNull(sql.Null[int]{Valid: true, V: 15})
		assert.Equal(t, 15, defined.Value())
		assert.Equal(t, false, defined.IsUndefined())
		assert.Equal(t, false, defined.IsNull())
		assert.Equal(t, true, defined.IsDefined())
	})
}

// Tests for.
//
// - Equal
// - EqualFunc
// - Map
// - Unwrap()
func TestUnd_Methods(t *testing.T) {
	definedFoo := und.Defined("foo")
	definedFoo2 := und.Defined("foo")
	definedBar := und.Defined("bar")
	null := und.Null[string]()
	undefined := und.Undefined[string]()

	t.Run("Equal", func(t *testing.T) {
		assert.Assert(t, definedFoo == definedFoo2)
		assert.Assert(t, definedFoo != definedBar)
		assert.Assert(t, null == null.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] { return o }))
		assert.Assert(t, null != undefined)

		for _, combo := range [][2]und.Und[string]{
			{definedFoo, definedFoo2},
			{definedBar, definedBar},
			{null, null},
			{undefined, undefined},
		} {
			assert.Assert(t, combo[0].Equal(combo[1]))
		}

		for _, combo := range [][2]und.Und[string]{
			{definedBar, null},
			{definedBar, undefined},
			{null, undefined},
		} {
			assert.Assert(t, !combo[0].Equal(combo[1]))
		}
	})

	t.Run("EqualFunc", func(t *testing.T) {
		for _, combo := range [][2]und.Und[string]{
			{definedFoo, definedFoo2},
			{definedBar, definedBar},
			{null, null},
			{undefined, undefined},
		} {
			assert.Assert(t, combo[0].EqualFunc(combo[1], func(i, j string) bool { return i == j }))
		}

		for _, combo := range [][2]und.Und[string]{
			{definedFoo, definedFoo2},
			{definedBar, definedBar},
		} {
			assert.Assert(t, !combo[0].EqualFunc(combo[1], func(i, j string) bool { return i != j }))
		}

		for _, combo := range [][2]und.Und[string]{
			{null, null},
			{undefined, undefined},
		} {
			assert.Assert(t, combo[0].EqualFunc(combo[1], func(i, j string) bool { return i != j }))
		}
	})

	t.Run("Map", func(t *testing.T) {
		assert.Equal(
			t,
			definedBar.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
				return und.Defined(o.Value().Value() + o.Value().Value()).Unwrap()
			}),
			und.Defined("barbar"),
		)
		assert.Equal(
			t,
			null.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
				return und.Defined("aa").Unwrap()
			}),
			und.Defined("aa"),
		)
		assert.Equal(
			t,
			undefined.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
				return und.Defined("bb").Unwrap()
			}),
			und.Defined("bb"),
		)
	})

	t.Run("Unwrap", func(t *testing.T) {
		assert.Equal(t, definedBar.Unwrap(), option.Some(option.Some("bar")))
		assert.Equal(t, null.Unwrap(), option.Some(option.None[string]()))
		assert.Equal(t, undefined.Unwrap(), option.None[option.Option[string]]())
	})
}
