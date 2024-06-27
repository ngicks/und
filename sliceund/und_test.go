package sliceund

import (
	"testing"

	"github.com/ngicks/und/internal/testcase"
)

func TestUnd(t *testing.T) {
	testcase.TestUnd(
		t,
		Defined[int](155),
		Null[int](),
		Undefined[int](),
		155,
		"155",
	)
}
