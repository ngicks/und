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
		FromOptions(option.Options[string]{option.None[string]()}),
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
	mixed1 := FromOptions([]option.Option[string]{option.Some("foo"), option.None[string](), option.Some("bar")})
	mixed1_2 := FromOptions([]option.Option[string]{option.Some("foo"), option.None[string](), option.Some("bar")})
	mixed2 := FromOptions([]option.Option[string]{option.None[string](), option.Some("bar")})
	null := Null[string]()
	undefined := Undefined[string]()

	t.Run("Equal", func(t *testing.T) {
		for _, combo := range [][2]Elastic[string]{
			{mixed1, mixed1_2},
			{mixed2, mixed2},
			{null, null},
			{undefined, undefined},
		} {
			assert.Assert(t, combo[0].Equal(combo[1]))
		}

		for _, combo := range [][2]Elastic[string]{
			{mixed2, null},
			{mixed2, undefined},
			{null, undefined},
		} {
			assert.Assert(t, !combo[0].Equal(combo[1]))
		}
	})
	t.Run("EqualFunc", func(t *testing.T) {
		for _, combo := range [][2]Elastic[string]{
			{mixed1, mixed1_2},
			{mixed2, mixed2},
			{null, null},
			{undefined, undefined},
		} {
			assert.Assert(t, combo[0].EqualFunc(combo[1], func(i, j string) bool { return i == j }))
		}

		for _, combo := range [][2]Elastic[string]{
			{mixed1, mixed1_2},
			{mixed2, mixed2},
		} {
			assert.Assert(t, !combo[0].EqualFunc(combo[1], func(i, j string) bool { return i != j }))
		}

		for _, combo := range [][2]Elastic[string]{
			{mixed2, null},
			{mixed2, undefined},
			{null, undefined},
		} {
			assert.Assert(t, !combo[0].EqualFunc(combo[1], func(i, j string) bool { return true }))
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
			mixed1.Map(mapper).Equal(FromOptions(
				[]option.Option[string]{option.Some("foofoo"), option.None[string](), option.Some("barbar")},
			)),
		)
		assert.Assert(
			t,
			null.Map(mapper).Equal(Null[string]()),
		)
		assert.Assert(
			t,
			undefined.Map(mapper).Equal(Undefined[string]()),
		)
	})

	t.Run("Unwrap", func(t *testing.T) {
		assert.Assert(t, mixed2.Unwrap().Equal(und.Defined(option.Options[string]{option.None[string](), option.Some("bar")})))
		assert.Assert(t, null.Unwrap().Equal(und.Null[option.Options[string]]()))
		assert.Assert(t, undefined.Unwrap().Equal(und.Undefined[option.Options[string]]()))
	})
}
