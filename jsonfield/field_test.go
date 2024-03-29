package jsonfield_test

import (
	"testing"
	"time"

	"github.com/ngicks/und/internal/testhelper"
	"github.com/ngicks/und/jsonfield"
	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	assert := assert.New(t)

	field := jsonfield.JsonField[int]{}
	assert.False(field.IsNull())
	assert.False(field.IsNonNull())
	assert.False(field.IsDefined())
	assert.True(field.IsUndefined())
	assert.Equal(int(0), field.Value())
	assert.Nil(field.Plain())

	field = jsonfield.Null[int]()
	assert.True(field.IsNull())
	assert.False(field.IsNonNull())
	assert.True(field.IsDefined())
	assert.False(field.IsUndefined())
	assert.Equal(int(0), field.Value())
	if p := field.Plain(); assert.NotNil(p) {
		assert.Nil(*p)
	}

	field = jsonfield.Defined[int](123)
	assert.False(field.IsNull())
	assert.True(field.IsNonNull())
	assert.True(field.IsDefined())
	assert.False(field.IsUndefined())
	if p := field.Plain(); assert.NotNil(p) && assert.NotNil(*p) {
		assert.Equal(int(123), **p)
	}
}

func TestEqual(t *testing.T) {
	pattern := []jsonfield.JsonField[int]{
		jsonfield.Undefined[int](),
		jsonfield.Null[int](),
		jsonfield.Defined[int](1),
		jsonfield.Defined[int](2178065),
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

func TestSerdeError(t *testing.T) {
	testhelper.TestSerdeError[jsonfield.JsonField[float64]](
		t,
		[]string{
			``,
			`false`,
			`[true,false]`,
		},
	)
	testhelper.TestSerdeError[jsonfieldSerdeTestTy[jsonfield.JsonField[float64]]](
		t,
		[]string{
			``,
			`{"F1":false}`,
			`{"F1":[true,false]}`,
		},
	)
}

func TestSerde(t *testing.T) {
	testhelper.TestSerde[jsonfieldSerdeTestTy[float64]](
		t,
		[]testhelper.SerdeTestSet[jsonfieldSerdeTestTy[float64]]{
			{
				Intern:      jsonfieldSerdeTestTy[float64]{F1: jsonfield.Undefined[float64]()},
				EncodedInto: `{}`,
			},
			{
				Intern:      jsonfieldSerdeTestTy[float64]{F1: jsonfield.Null[float64]()},
				EncodedInto: `{"F1":null}`,
			},
			{
				Intern:      jsonfieldSerdeTestTy[float64]{F1: jsonfield.Defined[float64](64905.790)},
				EncodedInto: `{"F1":64905.79}`,
			},
		},
	)

	// T is []U
	testhelper.TestSerde(
		t,
		[]testhelper.SerdeTestSet[jsonfieldSerdeTestTy[[]float64]]{
			{
				Intern:      jsonfieldSerdeTestTy[[]float64]{F1: jsonfield.Undefined[[]float64]()},
				EncodedInto: `{}`,
			},
			{
				Intern:      jsonfieldSerdeTestTy[[]float64]{F1: jsonfield.Defined[[]float64]([]float64{123})},
				EncodedInto: `{"F1":[123]}`,
			},
			{
				Intern:      jsonfieldSerdeTestTy[[]float64]{F1: jsonfield.Defined[[]float64]([]float64{123, 456, 789})},
				EncodedInto: `{"F1":[123,456,789]}`,
			},
		},
	)

	// types with a custom json.Marshal implementation.
	testhelper.TestSerde(
		t,
		[]testhelper.SerdeTestSet[jsonfieldSerdeTestTy[time.Time]]{
			{
				Intern: jsonfieldSerdeTestTy[time.Time]{
					F1: jsonfield.Defined[time.Time](time.Date(2022, 03, 04, 2, 12, 54, 0, time.UTC)),
				},
				Possible:    []string{`{"F1":"2022-03-04T02:12:54.000Z"}`, `{"F1":"2022-03-04T02:12:54Z"}`},
				EncodedInto: `{"F1":"2022-03-04T02:12:54Z"}`,
			},
		},
	)

	// recursive
	testhelper.TestSerde(
		t,
		[]testhelper.SerdeTestSet[jsonfieldSerdeTestTy[jsonfieldSerdeTestTy[string]]]{
			{
				Intern:      jsonfieldSerdeTestTy[jsonfieldSerdeTestTy[string]]{},
				EncodedInto: `{}`,
			},
			{
				Intern:      jsonfieldSerdeTestTy[jsonfieldSerdeTestTy[string]]{F1: jsonfield.Null[jsonfieldSerdeTestTy[string]]()},
				EncodedInto: `{"F1":null}`,
			},
			{
				Intern: jsonfieldSerdeTestTy[jsonfieldSerdeTestTy[string]]{
					F1: jsonfield.Defined[jsonfieldSerdeTestTy[string]](jsonfieldSerdeTestTy[string]{
						F1: jsonfield.Undefined[string](),
					}),
				},
				EncodedInto: `{"F1":{}}`,
			},
			{
				Intern: jsonfieldSerdeTestTy[jsonfieldSerdeTestTy[string]]{
					F1: jsonfield.Defined[jsonfieldSerdeTestTy[string]](jsonfieldSerdeTestTy[string]{
						F1: jsonfield.Null[string](),
					}),
				},
				EncodedInto: `{"F1":{"F1":null}}`,
			},
			{
				Intern: jsonfieldSerdeTestTy[jsonfieldSerdeTestTy[string]]{
					F1: jsonfield.Defined[jsonfieldSerdeTestTy[string]](jsonfieldSerdeTestTy[string]{
						F1: jsonfield.Defined[string]("foobar"),
					}),
				},
				EncodedInto: `{"F1":{"F1":"foobar"}}`,
			},
		},
	)
}

// special type for this test.
type jsonfieldSerdeTestTy[T any] struct {
	F1 jsonfield.JsonField[T]
}

func (t jsonfieldSerdeTestTy[T]) Equal(u jsonfieldSerdeTestTy[T]) bool {
	return t.F1.Equal(u.F1)
}
