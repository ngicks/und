package elastic

import (
	"slices"
	"testing"

	"github.com/ngicks/und/option"
	"gotest.tools/v3/assert"
)

func TestIter(t *testing.T) {
	d := FromValue(5)
	n := Null[int]()
	u := Undefined[int]()

	assert.DeepEqual(t, []option.Option[option.Options[int]]{option.Some(option.Options[int]{option.Some(5)})}, slices.Collect(d.Iter()))
	assert.DeepEqual(t, []option.Option[option.Options[int]]{option.None[option.Options[int]]()}, slices.Collect(n.Iter()))
	assert.DeepEqual(t, []option.Option[option.Options[int]](nil), slices.Collect(u.Iter()))
}
