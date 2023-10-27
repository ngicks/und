package nullable_test

import (
	"testing"

	"github.com/ngicks/und/v2/nullable"
	"github.com/stretchr/testify/assert"
)

func TestNullable(t *testing.T) {
	assert := assert.New(t)

	nullish := nullable.Nullable[int]{}
	assert.True(nullish.IsNull())
	assert.False(nullish.IsNonNull())
	assert.True(nullish.Equal(nullable.Null[int]()))
	assert.False(nullish.Equal(nullable.NonNull[int](0)))

	nullish = nullable.Null[int]()
	assert.True(nullish.IsNull())
	assert.False(nullish.IsNonNull())
	assert.True(nullish.Equal(nullable.Null[int]()))
	assert.False(nullish.Equal(nullable.NonNull[int](0)))

	nullish = nullable.NonNull[int](12)
	assert.False(nullish.IsNull())
	assert.True(nullish.IsNonNull())
	assert.False(nullish.Equal(nullable.Null[int]()))
	assert.False(nullish.Equal(nullable.NonNull[int](0)))
	assert.True(nullish.Equal(nullable.NonNull[int](12)))
}
