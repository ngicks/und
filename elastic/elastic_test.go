package elastic_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ngicks/generic"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/jsonfield"
	"github.com/ngicks/und/nullable"
	"github.com/stretchr/testify/assert"
)

func TestElastic_basic(t *testing.T) {
	assert := assert.New(t)

	{
		v := elastic.Undefined[int]()
		assert.False(v.IsSingle())
		assert.False(v.IsMultiple())
		assert.False(v.IsNull())
		assert.True(v.IsNullish())
		assert.Equal(0, v.ValueSingle())
		assert.Empty(cmp.Diff([]int{}, v.ValueMultiple()))
		assert.Nil(v.PlainSingle())
		assert.NotNil(v.PlainMultiple())
		assert.Empty(cmp.Diff([]*int{}, v.PlainMultiple()))
		assert.Equal(jsonfield.Undefined[int](), v.First())
	}
	{
		v := elastic.Null[int]()
		assert.True(v.IsSingle())
		assert.False(v.IsMultiple())
		assert.True(v.IsNull())
		assert.True(v.IsNullish())
		assert.Equal(0, v.ValueSingle())
		assert.Empty(cmp.Diff([]int{0}, v.ValueMultiple()))
		assert.Nil(v.PlainSingle())
		assert.NotNil(v.PlainMultiple())
		assert.Empty(cmp.Diff([]*int{nil}, v.PlainMultiple()))
		assert.Equal(jsonfield.Null[int](), v.First())
	}
	{
		v := elastic.FromSingle[int](123)
		assert.True(v.IsSingle())
		assert.False(v.IsMultiple())
		assert.False(v.IsNull())
		assert.False(v.IsNullish())
		assert.Equal(123, v.ValueSingle())
		assert.Empty(cmp.Diff([]int{123}, v.ValueMultiple()))
		assert.Equal(123, *v.PlainSingle())
		assert.NotNil(v.PlainMultiple())
		assert.Empty(cmp.Diff([]*int{generic.Escape(123)}, v.PlainMultiple()))
		assert.Equal(jsonfield.Defined[int](123), v.First())
	}
	{
		v := elastic.Defined[int]([]nullable.Nullable[int]{
			nullable.Null[int](),
			nullable.NonNull[int](555),
		})
		assert.False(v.IsSingle())
		assert.True(v.IsMultiple())
		assert.False(v.IsNull())
		assert.False(v.IsNullish())
		assert.Equal(0, v.ValueSingle())
		assert.Empty(cmp.Diff([]int{0, 555}, v.ValueMultiple()))
		assert.Nil(v.PlainSingle())
		assert.NotNil(v.PlainMultiple())
		assert.Empty(cmp.Diff([]*int{nil, generic.Escape(555)}, v.PlainMultiple()))
		assert.Equal(jsonfield.Null[int](), v.First())
	}
}

func TestEq(t *testing.T) {
	pattern := []elastic.Elastic[int]{
		elastic.Undefined[int](),
		elastic.Null[int](),
		elastic.FromSingle[int](0),
		elastic.FromSingle[int](5),
		elastic.FromSingle[int](123),
		elastic.FromMultiple[int]([]int{123, 555}),
		elastic.Defined[int]([]nullable.Nullable[int]{
			nullable.Null[int](),
			nullable.NonNull[int](555),
		}),
		elastic.Defined[int]([]nullable.Nullable[int]{
			nullable.Null[int](),
			nullable.NonNull[int](555),
			nullable.Null[int](),
		}),
	}

	for lIdx, l := range pattern {
		for rIdx, r := range pattern {
			eq := lIdx == rIdx
			if l.Equal(r) != eq {
				t.Errorf("Equal must return %t.\nl = %+#v\nr = %+#v", eq, l, r)
			}
		}
	}
}
