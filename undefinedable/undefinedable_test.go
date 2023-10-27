package undefinedable_test

import (
	"testing"

	"github.com/ngicks/und/v2/undefinedable"
	"github.com/stretchr/testify/assert"
)

func TestUndefinedable(t *testing.T) {
	assert := assert.New(t)

	undefined := undefinedable.Undefinedable[int]{}
	assert.True(undefined.IsUndefined())
	assert.False(undefined.IsDefined())
	assert.True(undefined.Equal(undefinedable.Undefined[int]()))
	assert.False(undefined.Equal(undefinedable.Defined[int](0)))

	undefined = undefinedable.Undefined[int]()
	assert.True(undefined.IsUndefined())
	assert.False(undefined.IsDefined())
	assert.True(undefined.Equal(undefinedable.Undefined[int]()))
	assert.False(undefined.Equal(undefinedable.Defined[int](0)))

	undefined = undefinedable.Defined[int](12)
	assert.False(undefined.IsUndefined())
	assert.True(undefined.IsDefined())
	assert.False(undefined.Equal(undefinedable.Undefined[int]()))
	assert.False(undefined.Equal(undefinedable.Defined[int](0)))
	assert.True(undefined.Equal(undefinedable.Defined[int](12)))
}
