package elastic

import (
	"testing"

	"gotest.tools/v3/assert"
)

// portable tests that can be copied from github.com/ngicks/und/elastic into github.com/ngicks/und/sliceund/elastic

// Tests for New-like function, e.g. FromPointer, WrapPointer
func TestUnd_new_functions(t *testing.T) {
	num := 15
	t.Run("FromValue", func(t *testing.T) {
		e := FromValue(num)
		assert.Equal(t, 15, e.Value())
		assert.Equal(t, false, e.IsUndefined())
		assert.Equal(t, false, e.IsNull())
		assert.Equal(t, true, e.IsDefined())
	})
	t.Run("FromPointer", func(t *testing.T) {
		fromNonNil := FromPointer(&num)
		assert.Equal(t, 15, fromNonNil.Value())
		assert.Equal(t, 1, fromNonNil.Len())
		assert.Equal(t, false, fromNonNil.IsUndefined())
		assert.Equal(t, false, fromNonNil.IsNull())
		assert.Equal(t, true, fromNonNil.IsDefined())
		fromNil := FromPointer((*int)(nil))
		assert.Equal(t, 0, fromNil.Value())
		assert.Equal(t, 0, fromNil.Len())
		assert.Equal(t, true, fromNil.IsUndefined())
		assert.Equal(t, false, fromNil.IsNull())
		assert.Equal(t, false, fromNil.IsDefined())
	})
	t.Run("WrapPointer", func(t *testing.T) {
		fromNonNil := WrapPointer(&num)
		assert.Equal(t, &num, fromNonNil.Value())
		assert.Equal(t, 1, fromNonNil.Len())
		assert.Equal(t, false, fromNonNil.IsUndefined())
		assert.Equal(t, false, fromNonNil.IsNull())
		assert.Equal(t, true, fromNonNil.IsDefined())
		fromNil := WrapPointer((*int)(nil))
		assert.Equal(t, (*int)(nil), fromNil.Value())
		assert.Equal(t, 0, fromNil.Len())
		assert.Equal(t, true, fromNil.IsUndefined())
		assert.Equal(t, false, fromNil.IsNull())
		assert.Equal(t, false, fromNil.IsDefined())
	})
	t.Run("FromPointers", func(t *testing.T) {
		e := FromPointers(&num, nil)
		assert.Equal(t, 15, e.Value())
		assert.Equal(t, 2, e.Len())
		assert.DeepEqual(t, []int{num, 0}, e.Values())
		assert.Equal(t, false, e.IsUndefined())
		assert.Equal(t, false, e.IsNull())
		assert.Equal(t, true, e.IsDefined())
	})
	t.Run("WrapPointers", func(t *testing.T) {
		e := WrapPointers(&num, nil)
		assert.Equal(t, &num, e.Value())
		assert.Equal(t, 2, e.Len())
		assert.DeepEqual(t, []*int{&num, nil}, e.Values())
		assert.Equal(t, false, e.IsUndefined())
		assert.Equal(t, false, e.IsNull())
		assert.Equal(t, true, e.IsDefined())
	})
}
