package und_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/ngicks/und"
	"github.com/ngicks/und/internal/testcase"
	"github.com/ngicks/und/internal/testtime"
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
// - EqualEqualer
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
		assert.Assert(t, null == null.InnerMap(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] { return o }))
		assert.Assert(t, null != undefined)

		for _, combo := range [][2]und.Und[string]{
			{definedFoo, definedFoo2},
			{definedBar, definedBar},
			{null, null},
			{undefined, undefined},
		} {
			assert.Assert(t, und.Equal(combo[0], combo[1]))
		}

		for _, combo := range [][2]und.Und[string]{
			{definedBar, null},
			{definedBar, undefined},
			{null, undefined},
		} {
			assert.Assert(t, !und.Equal(combo[0], combo[1]))
		}
	})

	t.Run("EqualEqualer", func(t *testing.T) {
		definedFoo := und.Defined(testtime.CurrInUTC)
		definedFoo2 := und.Defined(testtime.CurrInAsiaTokyo)
		definedBar := und.Defined(testtime.OneSecLater)
		null := und.Null[time.Time]()
		undefined := und.Undefined[time.Time]()

		assert.Assert(t, definedFoo != definedFoo2)
		assert.Assert(t, definedFoo != definedBar)
		assert.Assert(t, null == null.InnerMap(func(o option.Option[option.Option[time.Time]]) option.Option[option.Option[time.Time]] { return o }))
		assert.Assert(t, null != undefined)

		for _, combo := range [][2]und.Und[time.Time]{
			{definedFoo, definedFoo2},
			{definedBar, definedBar},
			{null, null},
			{undefined, undefined},
		} {
			assert.Assert(t, und.EqualEqualer(combo[0], combo[1]))
		}

		for _, combo := range [][2]und.Und[time.Time]{
			{definedBar, null},
			{definedBar, undefined},
			{null, undefined},
		} {
			assert.Assert(t, !und.EqualEqualer(combo[0], combo[1]))
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
			assert.Assert(t, und.Equal(combo[0], combo[1]))
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
			definedBar.InnerMap(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
				return und.Defined(o.Value().Value() + o.Value().Value()).Unwrap()
			}),
			und.Defined("barbar"),
		)
		assert.Equal(
			t,
			null.InnerMap(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
				return und.Defined("aa").Unwrap()
			}),
			und.Defined("aa"),
		)
		assert.Equal(
			t,
			undefined.InnerMap(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
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

func Test_Clone(t *testing.T) {
	foo := "foo"
	bar := "bar"

	org := foo
	def := und.WrapPointer(&org)
	null := und.Null[*string]()
	undefined := und.Undefined[*string]()

	assert.Equal(t, foo, *def.Value())

	cloneStringP := func(s *string) *string {
		if s == nil {
			return nil
		}
		v := *s
		return &v
	}
	cloned := def.CloneFunc(cloneStringP)
	assert.Equal(t, foo, *cloned.Value())
	shallow := und.Clone(def)
	assert.Equal(t, foo, *shallow.Value())

	org = bar
	assert.Equal(t, bar, *def.Value())
	assert.Equal(t, bar, *shallow.Value())
	assert.Equal(t, foo, *cloned.Value())

	cloned = null.CloneFunc(cloneStringP)
	assert.Assert(t, cloned.IsNull())

	cloned = und.Clone(null)
	assert.Assert(t, cloned.IsNull())

	cloned = undefined.CloneFunc(cloneStringP)
	assert.Assert(t, cloned.IsUndefined())

	cloned = und.Clone(undefined)
	assert.Assert(t, cloned.IsUndefined())
}
