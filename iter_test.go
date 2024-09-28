package und

import (
	"slices"
	"testing"

	"github.com/ngicks/und/option"
	"gotest.tools/v3/assert"
)

func TestIter(t *testing.T) {
	d := Defined(5)
	n := Null[int]()
	u := Undefined[int]()

	assert.DeepEqual(t, []option.Option[int]{option.Some(5)}, slices.Collect(d.Iter()))
	assert.DeepEqual(t, []option.Option[int]{option.None[int]()}, slices.Collect(n.Iter()))
	assert.DeepEqual(t, []option.Option[int](nil), slices.Collect(u.Iter()))
}
