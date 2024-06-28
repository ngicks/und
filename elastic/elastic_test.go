package elastic

import (
	"testing"

	"github.com/ngicks/und/internal/testcase"
	"github.com/ngicks/und/option"
)

func TestElastic(t *testing.T) {
	opts := []option.Option[string]{
		option.Some("foo"),
		option.None[string](),
		option.Some("bar"),
		option.Some("baz"),
	}
	testcase.TestElastic_non_addressable(
		t,
		FromOptions(opts),
		Null[string](),
		Undefined[string](),
		opts,
		`["foo",null,"bar","baz"]`,
	)
}
