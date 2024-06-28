package elastic

import (
	"testing"

	"github.com/ngicks/und"
	"github.com/ngicks/und/internal/testcase"
	"github.com/ngicks/und/option"
	"gotest.tools/v3/assert"
)

func TestElastic(t *testing.T) {
	opts := []option.Option[string]{
		option.Some("foo"),
		option.None[string](),
		option.Some("bar"),
		option.Some("baz"),
	}
	testcase.TestElastic_non_addressable(
		t,
		FromOptions(opts),
		Null[string](),
		Undefined[string](),
		opts,
		`["foo",null,"bar","baz"]`,
	)
}

// Test for
//   - Equal(other Elastic[T]) bool
//   - Map(f func(und.Und[option.Options[T]]) und.Und[option.Options[T]]) Elastic[T]
//   - Unwrap() und.Und[option.Options[T]]
func TestElastic_Methods(t *testing.T) {
	u1 := FromOptions([]option.Option[string]{option.Some("foo"), option.None[string](), option.Some("bar")})
	u1_2 := FromOptions([]option.Option[string]{option.Some("foo"), option.None[string](), option.Some("bar")})
	u2 := FromOptions([]option.Option[string]{option.None[string](), option.Some("bar")})
	u3 := Null[string]()
	u4 := Undefined[string]()

	t.Run("Equal", func(t *testing.T) {
		for _, combo := range [][2]Elastic[string]{
			{u1, u1_2},
			{u2, u2},
			{u3, u3},
			{u4, u4},
		} {
			assert.Assert(t, combo[0].Equal(combo[1]))
		}

		for _, combo := range [][2]Elastic[string]{
			{u2, u3},
			{u2, u4},
			{u3, u4},
		} {
			assert.Assert(t, !combo[0].Equal(combo[1]))
		}
	})

	t.Run("Map", func(t *testing.T) {
		mapper := func(u und.Und[option.Options[string]]) und.Und[option.Options[string]] {
			if !u.IsDefined() {
				return u
			}
			mapped := make([]option.Option[string], len(u.Value()))
			for i, v := range u.Value() {
				if v.IsSome() {
					mapped[i] = option.Some(v.Value() + v.Value())
				}
			}
			return und.Defined(option.Options[string](mapped))
		}

		assert.Assert(
			t,
			u1.Map(mapper).Equal(FromOptions(
				[]option.Option[string]{option.Some("foofoo"), option.None[string](), option.Some("barbar")},
			)),
		)
		assert.Assert(
			t,
			u3.Map(mapper).Equal(Null[string]()),
		)
		assert.Assert(
			t,
			u4.Map(mapper).Equal(Undefined[string]()),
		)
	})

	t.Run("Unwrap", func(t *testing.T) {
		assert.Assert(t, u2.Unwrap().Equal(und.Defined(option.Options[string]{option.None[string](), option.Some("bar")})))
		assert.Assert(t, u3.Unwrap().Equal(und.Null[option.Options[string]]()))
		assert.Assert(t, u4.Unwrap().Equal(und.Undefined[option.Options[string]]()))
	})
}
