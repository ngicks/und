package und

import (
	"testing"

	"github.com/ngicks/und/internal/testcase"
	"github.com/ngicks/und/option"
	"gotest.tools/v3/assert"
)

func TestUnd(t *testing.T) {
	testcase.TestUnd_addressable(
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
// - Map
// - Unwrap()
func TestUnd_Methods(t *testing.T) {
	u1 := Defined("foo")
	u1_2 := Defined("foo")
	u2 := Defined("bar")
	u3 := Null[string]()
	u4 := Undefined[string]()

	t.Run("Equal", func(t *testing.T) {
		assert.Assert(t, u1 == u1_2)
		assert.Assert(t, u1 != u2)
		assert.Assert(t, u3 == u3.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] { return o }))
		assert.Assert(t, u3 != u4)

		for _, combo := range [][2]Und[string]{
			{u1, u1_2},
			{u2, u2},
			{u3, u3},
			{u4, u4},
		} {
			assert.Assert(t, combo[0].Equal(combo[1]))
		}

		for _, combo := range [][2]Und[string]{
			{u2, u3},
			{u2, u4},
			{u3, u4},
		} {
			assert.Assert(t, !combo[0].Equal(combo[1]))
		}
	})

	t.Run("Map", func(t *testing.T) {
		assert.Equal(
			t,
			u2.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
				return Defined(o.Value().Value() + o.Value().Value()).Unwrap()
			}),
			Defined("barbar"),
		)
		assert.Equal(
			t,
			u3.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
				return Defined("aa").Unwrap()
			}),
			Defined("aa"),
		)
		assert.Equal(
			t,
			u4.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
				return Defined("bb").Unwrap()
			}),
			Defined("bb"),
		)
	})

	t.Run("Unwrap", func(t *testing.T) {
		assert.Equal(t, u2.Unwrap(), option.Some(option.Some("bar")))
		assert.Equal(t, u3.Unwrap(), option.Some(option.None[string]()))
		assert.Equal(t, u4.Unwrap(), option.None[option.Option[string]]())
	})
}
