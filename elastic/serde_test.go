package elastic_test

import (
	"testing"
	"time"

	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/internal/testhelper"
	"github.com/ngicks/und/nullable"
)

func TestElasticSerde(t *testing.T) {
	testhelper.TestSerde[elasticDecodeTy[float64]](
		t,
		[]testhelper.SerdeTestSet[elasticDecodeTy[float64]]{
			{
				Intern:      elasticDecodeTy[float64]{F1: elastic.Undefined[float64]()},
				EncodedInto: `{}`,
			},
			{
				Intern:      elasticDecodeTy[float64]{F1: elastic.Null[float64]()},
				Possible:    []string{`{"F1":null}`, `{"F1":[null]}`},
				EncodedInto: `{"F1":[null]}`,
			},
			{
				Intern:      elasticDecodeTy[float64]{F1: elastic.Single[float64](123)},
				Possible:    []string{`{"F1":123}`, `{"F1":[123]}`},
				EncodedInto: `{"F1":[123]}`,
			},
			{
				Intern:      elasticDecodeTy[float64]{F1: elastic.Multiple[float64]([]float64{123, 456})},
				EncodedInto: `{"F1":[123,456]}`,
			},
			{
				Intern: elasticDecodeTy[float64]{
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
				Intern: elasticDecodeTy[float64]{
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
		[]testhelper.SerdeTestSet[elasticDecodeTy[[]float64]]{
			{
				Intern:      elasticDecodeTy[[]float64]{F1: elastic.Undefined[[]float64]()},
				EncodedInto: `{}`,
			},
			{
				Intern:      elasticDecodeTy[[]float64]{F1: elastic.Null[[]float64]()},
				Possible:    []string{`{"F1":null}`, `{"F1":[null]}`},
				EncodedInto: `{"F1":[null]}`,
			},
			{
				Intern:      elasticDecodeTy[[]float64]{F1: elastic.Single[[]float64]([]float64{123})},
				Possible:    []string{`{"F1":[123]}`, `{"F1":[[123]]}`},
				EncodedInto: `{"F1":[[123]]}`,
			},
			{
				Intern: elasticDecodeTy[[]float64]{
					F1: elastic.Multiple[[]float64]([][]float64{{123, 456}, {789}}),
				},
				EncodedInto: `{"F1":[[123,456],[789]]}`,
			},
			{
				Intern: elasticDecodeTy[[]float64]{
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
				Intern: elasticDecodeTy[[]float64]{
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
		[]testhelper.SerdeTestSet[elasticDecodeTy[time.Time]]{
			{
				Intern: elasticDecodeTy[time.Time]{
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
		[]testhelper.SerdeTestSet[elasticDecodeTy[elasticDecodeTy[string]]]{
			{
				Intern:      elasticDecodeTy[elasticDecodeTy[string]]{},
				EncodedInto: `{}`,
			},
			{
				Intern:      elasticDecodeTy[elasticDecodeTy[string]]{F1: elastic.Null[elasticDecodeTy[string]]()},
				Possible:    []string{`{"F1":null}`, `{"F1":[null]}`},
				EncodedInto: `{"F1":[null]}`,
			},
			{
				Intern: elasticDecodeTy[elasticDecodeTy[string]]{
					F1: elastic.Defined[elasticDecodeTy[string]](
						[]nullable.Nullable[elasticDecodeTy[string]]{
							nullable.NonNull[elasticDecodeTy[string]](
								elasticDecodeTy[string]{
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
				Intern: elasticDecodeTy[elasticDecodeTy[string]]{
					F1: elastic.Defined[elasticDecodeTy[string]](
						[]nullable.Nullable[elasticDecodeTy[string]]{
							nullable.NonNull[elasticDecodeTy[string]](
								elasticDecodeTy[string]{
									F1: elastic.Single[string]("barrr"),
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
type elasticDecodeTy[T any] struct {
	F1 elastic.Elastic[T]
}

func (t elasticDecodeTy[T]) Equal(u elasticDecodeTy[T]) bool {
	return t.F1.Equal(u.F1)
}
