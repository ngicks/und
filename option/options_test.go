package option

import (
	"testing"
	"time"

	"github.com/ngicks/und/internal/testtime"
	"gotest.tools/v3/assert"
)

func TestEqualOptionsEqualer(t *testing.T) {
	a := Options[time.Time]{Some(testtime.CurrInUTC), Some(testtime.CurrInAsiaTokyo)}
	b := Options[time.Time]{Some(testtime.CurrInUTC), Some(testtime.CurrInUTC)}

	assert.Assert(t, !EqualOptions(a, b))
	assert.Assert(t, EqualOptionsEqualer(a, b))
}
