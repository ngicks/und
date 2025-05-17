package elastic

import (
	"slices"
	"testing"
	"time"

	"github.com/ngicks/und/internal/testcase"
	"github.com/ngicks/und/internal/testtime"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	"gotest.tools/v3/assert"
)

func TestElastic_newFuncs(t *testing.T) {
	for _, e := range []Elastic[int]{
		FromOptionSeq(slices.Values([]option.Option[int](nil))),
		FromOptions[int](),
		FromPointers[int](),
		FromValues[int](),
	} {
		assert.Assert(t, e.Unwrap().Value() != nil)
	}
}

func TestElastic(t *testing.T) {
	opts := []option.Option[string]{
		option.Some("foo"),
		option.None[string](),
		option.Some("bar"),
		option.Some("baz"),
	}
	testcase.TestElastic_non_addressable(
		t,
		FromOptions(option.None[string]()),
		FromOptions(opts...),
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
	mixed1 := FromOptions(option.Some("foo"), option.None[string](), option.Some("bar"))
	mixed1_2 := FromOptions(option.Some("foo"), option.None[string](), option.Some("bar"))
	mixed2 := FromOptions(option.None[string](), option.Some("bar"))
	null := Null[string]()
	undefined := Undefined[string]()

	t.Run("Equal", func(t *testing.T) {
		for _, combo := range [][2]Elastic[string]{
			{mixed1, mixed1_2},
			{mixed2, mixed2},
			{null, null},
			{undefined, undefined},
		} {
			assert.Assert(
				t,
				combo[0].EqualFunc(
					combo[1],
					func(i, j string) bool { return i == j },
				),
			)
		}

		for _, combo := range [][2]Elastic[string]{
			{mixed2, null},
			{mixed2, undefined},
			{null, undefined},
		} {
			assert.Assert(
				t,
				!combo[0].EqualFunc(
					combo[1],
					func(i, j string) bool { return i == j },
				),
			)
		}
	})

	t.Run("EqualEqualer", func(t *testing.T) {
		mixed1 := FromOptions(option.Some(testtime.CurrInUTC), option.None[time.Time](), option.Some(testtime.CurrInAsiaTokyo))
		mixed1_2 := FromOptions(option.Some(testtime.CurrInUTC), option.None[time.Time](), option.Some(testtime.CurrInAsiaTokyo))
		mixed2 := FromOptions(option.None[time.Time](), option.Some(testtime.CurrInAsiaTokyo))
		null := Null[time.Time]()
		undefined := Undefined[time.Time]()

		for _, combo := range [][2]Elastic[time.Time]{
			{mixed1, mixed1_2},
			{mixed2, mixed2},
			{null, null},
			{undefined, undefined},
		} {
			assert.Assert(t, EqualEqualer(combo[0], combo[1]))
		}

		for _, combo := range [][2]Elastic[time.Time]{
			{mixed2, null},
			{mixed2, undefined},
			{null, undefined},
		} {
			assert.Assert(t, !EqualEqualer(combo[0], combo[1]))
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
		mapper := func(u sliceund.Und[option.Options[string]]) sliceund.Und[option.Options[string]] {
			if !u.IsDefined() {
				return u
			}
			mapped := make([]option.Option[string], len(u.Value()))
			for i, v := range u.Value() {
				if v.IsSome() {
					mapped[i] = option.Some(v.Value() + v.Value())
				}
			}
			return sliceund.Defined(option.Options[string](mapped))
		}

		assert.Assert(
			t,
			mixed1.InnerMap(mapper).EqualFunc(
				FromOptions(
					option.Some("foofoo"),
					option.None[string](),
					option.Some("barbar"),
				),
				func(i, j string) bool { return i == j },
			),
		)
		assert.Assert(
			t,
			null.InnerMap(mapper).EqualFunc(Null[string](), func(i, j string) bool { return i == j }),
		)
		assert.Assert(
			t,
			undefined.InnerMap(mapper).EqualFunc(Undefined[string](), func(i, j string) bool { return i == j }),
		)
	})

	t.Run("Unwrap", func(t *testing.T) {
		assert.Assert(t, mixed2.Unwrap().EqualFunc(sliceund.Defined(option.Options[string]{option.None[string](), option.Some("bar")}), option.EqualOptions))
		assert.Assert(t, null.Unwrap().EqualFunc(sliceund.Null[option.Options[string]](), option.EqualOptions))
		assert.Assert(t, undefined.Unwrap().EqualFunc(sliceund.Undefined[option.Options[string]](), option.EqualOptions))
	})
}

func Test_Clone(t *testing.T) {
	foo := "foo"
	bar := "bar"

	org := foo
	def := WrapPointer(&org)
	null := Null[*string]()
	undefined := Undefined[*string]()

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
	shallow := Clone(def)
	assert.Equal(t, foo, *shallow.Value())

	org = bar
	assert.Equal(t, bar, *def.Value())
	assert.Equal(t, bar, *shallow.Value())
	assert.Equal(t, foo, *cloned.Value())

	cloned = null.CloneFunc(cloneStringP)
	assert.Assert(t, cloned.IsNull())

	cloned = Clone(null)
	assert.Assert(t, cloned.IsNull())

	cloned = undefined.CloneFunc(cloneStringP)
	assert.Assert(t, cloned.IsUndefined())

	cloned = Clone(undefined)
	assert.Assert(t, cloned.IsUndefined())
}
