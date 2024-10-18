package sliceund

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
	definedNull := Null[string]()
	undefined := Undefined[string]()

	t.Run("Equal", func(t *testing.T) {
		for _, combo := range [][2]Und[string]{
			{definedFoo, definedFoo2},
			{definedBar, definedBar},
			{definedNull, definedNull},
			{undefined, undefined},
		} {
			assert.Assert(t, combo[0].Equal(combo[1]))
		}

		for _, combo := range [][2]Und[string]{
			{definedBar, definedNull},
			{definedBar, undefined},
			{definedNull, undefined},
		} {
			assert.Assert(t, !combo[0].Equal(combo[1]))
		}
	})

	t.Run("EqualFunc", func(t *testing.T) {
		for _, combo := range [][2]Und[string]{
			{definedFoo, definedFoo2},
			{definedBar, definedBar},
			{definedNull, definedNull},
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
			{definedBar, definedNull},
			{definedBar, undefined},
			{definedNull, undefined},
		} {
			assert.Assert(t, !combo[0].EqualFunc(combo[1], func(i, j string) bool { return true }))
		}
	})

	t.Run("Map", func(t *testing.T) {
		assert.Assert(
			t,
			definedBar.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
				return Defined(o.Value().Value() + o.Value().Value()).Unwrap()
			}).Equal(Defined("barbar")),
		)
		assert.Assert(
			t,
			definedNull.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
				return Defined("aa").Unwrap()
			}).Equal(Defined("aa")),
		)
		assert.Assert(
			t,
			undefined.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
				return Defined("bb").Unwrap()
			}).Equal(Defined("bb")),
		)
	})

	t.Run("Unwrap", func(t *testing.T) {
		assert.Equal(t, definedBar.Unwrap(), option.Some(option.Some("bar")))
		assert.Equal(t, definedNull.Unwrap(), option.Some(option.None[string]()))
		assert.Equal(t, undefined.Unwrap(), option.None[option.Option[string]]())
	})
}
