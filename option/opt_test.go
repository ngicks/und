package option_test

import (
	"sync/atomic"
	"testing"

	"github.com/ngicks/und/option"
	"github.com/stretchr/testify/assert"
)

func TestOption(t *testing.T) {
	assert := assert.New(t)

	opt := option.Option[int]{}
	assert.True(opt.IsNone())
	assert.False(opt.IsSome())
	assert.Equal(int(0), opt.Value())
	assert.Nil(opt.Plain())
	cc := callCount[int]{fn: func(v int) int { return v * 2 }}
	assert.Equal(option.Option[int]{}, opt.Map(cc.Fn))
	assert.Equal(int32(0), cc.count.Load())

	opt = option.None[int]()
	assert.True(opt.IsNone())
	assert.False(opt.IsSome())
	assert.Equal(int(0), opt.Value())
	assert.Nil(opt.Plain())
	cc = callCount[int]{fn: func(v int) int { return v * 2 }}
	assert.Equal(option.Option[int]{}, opt.Map(cc.Fn))
	assert.Equal(int32(0), cc.count.Load())

	opt = option.Some[int](12)
	assert.False(opt.IsNone())
	assert.True(opt.IsSome())
	assert.Equal(int(12), opt.Value())
	assert.Equal(int(12), *opt.Plain())
	cc = callCount[int]{fn: func(v int) int { return v * 2 }}
	assert.Equal(option.Some[int](24), opt.Map(cc.Fn))
	assert.Equal(int32(1), cc.count.Load())

	// copied
	p := opt.Plain()
	*p = 5
	assert.Equal(int(12), *opt.Plain())
}

type callCount[T any] struct {
	count atomic.Int32
	fn    func(v T) T
}

func (c *callCount[T]) Fn(v T) T {
	c.count.Add(1)
	return c.fn(v)
}

func TestSerde(t *testing.T) {
	assert := assert.New(t)

	opt := option.None[int]()

	bin, err := opt.MarshalJSON()
	assert.NoError(err)
	assert.Equal("null", string(bin))

	opt = option.Some[int](268)
	bin, err = opt.MarshalJSON()
	assert.NoError(err)
	assert.Equal("268", string(bin))

	err = opt.UnmarshalJSON([]byte("null"))
	assert.NoError(err)
	assert.True(opt.IsNone())
	assert.Equal(int(0), opt.Value())

	err = opt.UnmarshalJSON([]byte("8993"))
	assert.NoError(err)
	assert.True(opt.IsSome())
	assert.Equal(int(8993), opt.Value())
}
