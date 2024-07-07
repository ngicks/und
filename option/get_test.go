package option

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestOption_get(t *testing.T) {
	t.Run("GetMap", func(t *testing.T) {
		m := map[string]int{
			"foo": 0,
			"bar": 2,
		}

		for k, v := range m {
			o := GetMap(m, k)
			assert.Assert(t, o.IsSome())
			assert.Equal(t, o.Value(), v)
		}

		o := GetMap(m, "baz")
		assert.Assert(t, o.IsNone())
	})

	t.Run("GetSlice", func(t *testing.T) {
		s := []string{"foo", "bar", "baz"}

		o := GetSlice(s, -1)
		assert.Assert(t, o.IsNone())
		for i := range s {
			o := GetSlice(s, i)
			assert.Assert(t, o.IsSome())
			assert.Equal(t, o.Value(), s[i])
		}
		o = GetSlice(s, len(s))
		assert.Assert(t, o.IsNone())
	})
}
