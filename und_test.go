package und

import (
	"testing"

	"github.com/ngicks/und/internal/testcase"
	"github.com/ngicks/und/option"
	"gotest.tools/v3/assert"
)

func TestUnd(t *testing.T) {
	testcase.TestUnd_non_addressable(
		t,
		Defined[int](155),
		Null[int](),
		Undefined[int](),
		155,
		"155",
	)
}

// Tests for.
//
// - Equal
// - EqualFunc
// - Map
// - Unwrap()
func TestUnd_Methods(t *testing.T) {
	definedFoo := Defined("foo")
	definedFoo2 := Defined("foo")
	definedBar := Defined("bar")
	null := Null[string]()
	undefined := Undefined[string]()

	t.Run("Equal", func(t *testing.T) {
		assert.Assert(t, definedFoo == definedFoo2)
		assert.Assert(t, definedFoo != definedBar)
		assert.Assert(t, null == null.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] { return o }))
		assert.Assert(t, null != undefined)

		for _, combo := range [][2]Und[string]{
			{definedFoo, definedFoo2},
			{definedBar, definedBar},
			{null, null},
			{undefined, undefined},
		} {
			assert.Assert(t, combo[0].Equal(combo[1]))
		}

		for _, combo := range [][2]Und[string]{
			{definedBar, null},
			{definedBar, undefined},
			{null, undefined},
		} {
			assert.Assert(t, !combo[0].Equal(combo[1]))
		}
	})

	t.Run("EqualFunc", func(t *testing.T) {
		for _, combo := range [][2]Und[string]{
			{definedFoo, definedFoo2},
			{definedBar, definedBar},
			{null, null},
			{undefined, undefined},
		} {
			assert.Assert(t, combo[0].EqualFunc(combo[1], func(i, j string) bool { return i == j }))
		}

		for _, combo := range [][2]Und[string]{
			{definedFoo, definedFoo2},
			{definedBar, definedBar},
		} {
			assert.Assert(t, !combo[0].EqualFunc(combo[1], func(i, j string) bool { return i != j }))
		}

		for _, combo := range [][2]Und[string]{
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
				return Defined(o.Value().Value() + o.Value().Value()).Unwrap()
			}),
			Defined("barbar"),
		)
		assert.Equal(
			t,
			null.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
				return Defined("aa").Unwrap()
			}),
			Defined("aa"),
		)
		assert.Equal(
			t,
			undefined.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
				return Defined("bb").Unwrap()
			}),
			Defined("bb"),
		)
	})

	t.Run("Unwrap", func(t *testing.T) {
		assert.Equal(t, definedBar.Unwrap(), option.Some(option.Some("bar")))
		assert.Equal(t, null.Unwrap(), option.Some(option.None[string]()))
		assert.Equal(t, undefined.Unwrap(), option.None[option.Option[string]]())
	})
}
