// This package ever stays unstable.
package testhelper

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ngicks/und/v2/option"
	"github.com/ngicks/und/v2/serde"
	"github.com/stretchr/testify/assert"
)

// A set of internal, possible input and encode output representations.
type SerdeTestSet[T option.Equality[T]] struct {
	Intern      T        // internal representation
	Possible    []string // possible representations
	EncodedInto string   // intern encoded into
}

func TestSerde[T option.Equality[T]](t *testing.T, sets []SerdeTestSet[T]) bool {
	t.Helper()

	var firstFail bool
	for _, set := range sets {
		if set.Possible == nil {
			set.Possible = append(set.Possible, set.EncodedInto)
		}

		for _, encoded := range set.Possible {
			if !TestDecode[T](t, encoded, set.Intern) {
				firstFail = true
			}
			if !TestEncode[T](t, set.Intern, set.EncodedInto) {
				firstFail = true
			}
		}
	}
	return !firstFail
}

func TestDecode[T option.Equality[T]](t *testing.T, input string, expected T) bool {
	t.Helper()

	var v T
	err := serde.Unmarshal([]byte(input), &v)
	if err != nil {
		t.Errorf("must not cause an error but is %+#v", err)
		return false
	}
	// test itself is depending on the Equal method.
	if !expected.Equal(v) {
		diff := cmp.Diff(expected, v)
		t.Errorf(
			"not equal:\ninput = %s\nexpected = %+v,\nactual   = %+v,\ndiff = %s",
			input, expected, v, diff,
		)
		return false
	}

	return true
}

func TestEncode[T any](t *testing.T, intern T, expected string) bool {
	t.Helper()

	bin, err := serde.Marshal(intern)
	if err != nil {
		t.Errorf("must not cause an error but is %+#v", err)
		return false
	}
	if string(bin) != expected {
		t.Errorf("not equal.\nencoded  = %s\nexpected = %s", bin, expected)
		return false
	}

	return true
}

func TestSerdeError[T any](t *testing.T, inputs []string) bool {
	assert := assert.New(t)

	var firstError bool

	for _, erroneousInput := range inputs {
		var v T
		err := serde.Unmarshal([]byte(erroneousInput), &v)
		if !assert.Error(err) {
			firstError = true
		}
		if !assert.True(reflect.ValueOf(v).IsZero(), "must be zero but is %+#v", v) {
			firstError = true
		}
	}

	return !firstError
}
