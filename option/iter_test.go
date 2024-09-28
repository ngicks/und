package option

import (
	"slices"
	"testing"

	"gotest.tools/v3/assert"
)

func TestIter(t *testing.T) {
	s := Some(5)
	n := None[int]()

	assert.DeepEqual(t, []int{5}, slices.Collect(s.Iter()))
	assert.DeepEqual(t, []int(nil), slices.Collect(n.Iter()))
}
