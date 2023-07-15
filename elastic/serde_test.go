package elastic_test

import (
	"testing"
	"time"

	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/internal/testhelper"
	"github.com/ngicks/und/nullable"
)

func TestSerdeError(t *testing.T) {
	testhelper.TestSerdeError[elastic.Elastic[float64]](
		t,
		[]string{
			``,
			`false`,
			`[true,false]`,
		},
	)
	testhelper.TestSerdeError[elastic.Elastic[elasticSerdeTestTy[float64]]](
		t,
		[]string{
			``,
			`{"F1":false}`,
			`{"F1":[true,false]}`,
		},
	)
}

func TestSerde(t *testing.T) {
	testhelper.TestSerde[elasticSerdeTestTy[float64]](
		t,
		[]testhelper.SerdeTestSet[elasticSerdeTestTy[float64]]{
			{
				Intern:      elasticSerdeTestTy[float64]{F1: elastic.Undefined[float64]()},
				EncodedInto: `{}`,
			},
			{
				Intern:      elasticSerdeTestTy[float64]{F1: elastic.Null[float64]()},
				Possible:    []string{`{"F1":null}`, `{"F1":[null]}`},
				EncodedInto: `{"F1":[null]}`,
			},
			{
				Intern:      elasticSerdeTestTy[float64]{F1: elastic.FromSingle[float64](123)},
				Possible:    []string{`{"F1":123}`, `{"F1":[123]}`},
				EncodedInto: `{"F1":[123]}`,
			},
			{
				Intern:      elasticSerdeTestTy[float64]{F1: elastic.FromMultiple[float64]([]float64{123, 456})},
				EncodedInto: `{"F1":[123,456]}`,
			},
			{
				Intern: elasticSerdeTestTy[float64]{
					F1: elastic.Defined[float64](
						[]nullable.Nullable[float64]{
							nullable.NonNull[float64](123),
							nullable.Null[float64](),
						},
					),
				},
				EncodedInto: `{"F1":[123,null]}`,
			},
			{
				Intern: elasticSerdeTestTy[float64]{
					F1: elastic.Defined[float64](
						[]nullable.Nullable[float64]{
							nullable.Null[float64](),
							nullable.Null[float64](),
						},
					),
				},
				EncodedInto: `{"F1":[null,null]}`,
			},
		},
	)

	// T is []U
	testhelper.TestSerde(
		t,
		[]testhelper.SerdeTestSet[elasticSerdeTestTy[[]float64]]{
			{
				Intern:      elasticSerdeTestTy[[]float64]{F1: elastic.Undefined[[]float64]()},
				EncodedInto: `{}`,
			},
			{
				Intern:      elasticSerdeTestTy[[]float64]{F1: elastic.Null[[]float64]()},
				Possible:    []string{`{"F1":null}`, `{"F1":[null]}`},
				EncodedInto: `{"F1":[null]}`,
			},
			{
				Intern:      elasticSerdeTestTy[[]float64]{F1: elastic.FromSingle[[]float64]([]float64{123})},
				Possible:    []string{`{"F1":[123]}`, `{"F1":[[123]]}`},
				EncodedInto: `{"F1":[[123]]}`,
			},
			{
				Intern: elasticSerdeTestTy[[]float64]{
					F1: elastic.FromMultiple[[]float64]([][]float64{{123, 456}, {789}}),
				},
				EncodedInto: `{"F1":[[123,456],[789]]}`,
			},
			{
				Intern: elasticSerdeTestTy[[]float64]{
					F1: elastic.Defined[[]float64](
						[]nullable.Nullable[[]float64]{
							nullable.NonNull[[]float64]([]float64{123}),
							nullable.Null[[]float64](),
						},
					),
				},
				EncodedInto: `{"F1":[[123],null]}`,
			},
			{
				Intern: elasticSerdeTestTy[[]float64]{
					F1: elastic.Defined[[]float64](
						[]nullable.Nullable[[]float64]{
							nullable.Null[[]float64](),
							nullable.Null[[]float64](),
						},
					),
				},
				EncodedInto: `{"F1":[null,null]}`,
			},
		},
	)

	// types with a custom json.Marshal implementation.
	testhelper.TestSerde(
		t,
		[]testhelper.SerdeTestSet[elasticSerdeTestTy[time.Time]]{
			{
				Intern: elasticSerdeTestTy[time.Time]{
					F1: elastic.Defined[time.Time](
						[]nullable.Nullable[time.Time]{
							nullable.Null[time.Time](),
							nullable.NonNull[time.Time](time.Date(2022, 03, 04, 2, 12, 54, 0, time.UTC)),
						},
					),
				},
				Possible:    []string{`{"F1":[null,"2022-03-04T02:12:54.000Z"]}`, `{"F1":[null,"2022-03-04T02:12:54Z"]}`},
				EncodedInto: `{"F1":[null,"2022-03-04T02:12:54Z"]}`,
			},
		},
	)

	// recursive
	testhelper.TestSerde(
		t,
		[]testhelper.SerdeTestSet[elasticSerdeTestTy[elasticSerdeTestTy[string]]]{
			{
				Intern:      elasticSerdeTestTy[elasticSerdeTestTy[string]]{},
				EncodedInto: `{}`,
			},
			{
				Intern:      elasticSerdeTestTy[elasticSerdeTestTy[string]]{F1: elastic.Null[elasticSerdeTestTy[string]]()},
				Possible:    []string{`{"F1":null}`, `{"F1":[null]}`},
				EncodedInto: `{"F1":[null]}`,
			},
			{
				Intern: elasticSerdeTestTy[elasticSerdeTestTy[string]]{
					F1: elastic.Defined[elasticSerdeTestTy[string]](
						[]nullable.Nullable[elasticSerdeTestTy[string]]{
							nullable.NonNull[elasticSerdeTestTy[string]](
								elasticSerdeTestTy[string]{
									F1: elastic.Undefined[string](),
								},
							),
						},
					),
				},
				Possible:    []string{`{"F1":{}}`, `{"F1":[{}]}`},
				EncodedInto: `{"F1":[{}]}`,
			},
			{
				Intern: elasticSerdeTestTy[elasticSerdeTestTy[string]]{
					F1: elastic.Defined[elasticSerdeTestTy[string]](
						[]nullable.Nullable[elasticSerdeTestTy[string]]{
							nullable.NonNull[elasticSerdeTestTy[string]](
								elasticSerdeTestTy[string]{
									F1: elastic.FromSingle[string]("barrr"),
								},
							),
						},
					),
				},
				Possible: []string{
					`{"F1":{"F1":"barrr"}}`,
					`{"F1":[{"F1":"barrr"}]}`,
					`{"F1":{"F1":["barrr"]}}`,
					`{"F1":[{"F1":["barrr"]}]}`,
				},
				EncodedInto: `{"F1":[{"F1":["barrr"]}]}`,
			},
		},
	)
}

// special type for this test.
type elasticSerdeTestTy[T any] struct {
	F1 elastic.Elastic[T]
}

func (t elasticSerdeTestTy[T]) Equal(u elasticSerdeTestTy[T]) bool {
	return t.F1.Equal(u.F1)
}
