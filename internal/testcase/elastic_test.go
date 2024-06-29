package testcase_test

import (
	"encoding/json"
	"testing"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/ngicks/und/elastic"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

type elasticV2 struct {
	Padding1 int                     `json:",omitzero"`
	V        elastic.Elastic[string] `json:",omitzero"`
	Padding2 int                     `json:",omitzero"`
}

type sliceElasticV1 struct {
	Padding1 int                          `json:",omitempty"`
	V        sliceelastic.Elastic[string] `json:",omitempty"`
	Padding2 int                          `json:",omitempty"`
}

type sliceElasticV2 struct {
	Padding1 int                          `json:",omitzero"`
	V        sliceelastic.Elastic[string] `json:",omitzero"`
	Padding2 int                          `json:",omitzero"`
}

type serdeTestCaseElastic struct {
	bin       string
	marshaled string
	state     int
	values    []*string
}

func ptr[T any](t T) *T {
	return &t
}

var elasticTestCases = []serdeTestCaseElastic{
	// single value
	{bin: `{"Padding1":10,"Padding2":20}`, state: 0, values: nil},
	{bin: `{"Padding1":10,"V":null,"Padding2":20}`, state: 1, values: nil},
	{bin: `{"Padding1":10,"V":"foo","Padding2":20}`, marshaled: `{"Padding1":10,"V":["foo"],"Padding2":20}`, state: 2, values: []*string{ptr("foo")}},
	{bin: `{"Padding2":20}`, state: 0, values: nil},
	{bin: `{"V":null,"Padding2":20}`, state: 1, values: nil},
	{bin: `{"V":"bar","Padding2":20}`, marshaled: `{"V":["bar"],"Padding2":20}`, state: 2, values: []*string{ptr("bar")}},
	{bin: `{"Padding1":10}`, state: 0, values: nil},
	{bin: `{"Padding1":10,"V":null}`, state: 1, values: nil},
	{bin: `{"Padding1":10,"V":"baz"}`, marshaled: `{"Padding1":10,"V":["baz"]}`, state: 2, values: []*string{ptr("baz")}},
	// array
	{bin: `{"Padding1":10,"V":[],"Padding2":20}`, state: 3, values: []*string{}},
	{bin: `{"Padding1":10,"V":["foo"],"Padding2":20}`, state: 3, values: []*string{ptr("foo")}},
	{bin: `{"Padding1":10,"V":["foo","bar"],"Padding2":20}`, state: 3, values: []*string{ptr("foo"), ptr("bar")}},
	{bin: `{"Padding1":10,"V":[null,"foo",null,"bar"],"Padding2":20}`, state: 3, values: []*string{nil, ptr("foo"), nil, ptr("bar")}},
}

func TestElastic_serde(t *testing.T) {
	for _, tc := range elasticTestCases {
		t.Run(tc.bin, func(t *testing.T) {
			var (
				u2 elasticV2
				s1 sliceElasticV1
				s2 sliceElasticV2
			)

			assert.NilError(t, json.Unmarshal([]byte(tc.bin), &s1))
			assert.NilError(t, jsonv2.Unmarshal([]byte(tc.bin), &u2))
			assert.NilError(t, jsonv2.Unmarshal([]byte(tc.bin), &s2))

			assertStateElastic(t, u2.V, tc.state, tc.values)
			assertStateElastic(t, s1.V, tc.state, tc.values)
			assertStateElastic(t, s2.V, tc.state, tc.values)

			var (
				bin []byte
				err error
			)

			marshaled := tc.marshaled
			if marshaled == "" {
				marshaled = tc.bin
			}

			bin, err = json.Marshal(s1)
			assert.NilError(t, err)
			assert.Equal(t, string(bin), marshaled)

			bin, err = jsonv2.Marshal(u2)
			assert.NilError(t, err)
			assert.Equal(t, string(bin), marshaled)

			bin, err = jsonv2.Marshal(s2)
			assert.NilError(t, err)
			assert.Equal(t, string(bin), marshaled)
		})
	}
}

type ielastic[T any] interface {
	iund[T]
	Pointers() []*T
}

func assertStateElastic[T ielastic[U], U any](t *testing.T, u T, state int, v []*U) {
	t.Helper()
	switch state {
	case 0:
		assert.Assert(t, u.IsUndefined())
	case 1:
		assert.Assert(t, u.IsNull())
	case 2, 3:
		assert.Assert(t, u.IsDefined())
	}
	assert.Assert(t, cmp.DeepEqual(u.Pointers(), v))
}
