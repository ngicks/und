package testcase_test

import (
	"encoding/json"
	"testing"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/ngicks/und"
	"github.com/ngicks/und/sliceund"
	"gotest.tools/v3/assert"
)

type undV2 struct {
	Padding1 int             `json:",omitzero"`
	V        und.Und[string] `json:",omitzero"`
	Padding2 int             `json:",omitzero"`
}

type slicesUndV1 struct {
	Padding1 int                  `json:",omitempty"`
	V        sliceund.Und[string] `json:",omitempty"`
	Padding2 int                  `json:",omitempty"`
}

type slicesUndV2 struct {
	Padding1 int                  `json:",omitzero"`
	V        sliceund.Und[string] `json:",omitzero"`
	Padding2 int                  `json:",omitzero"`
}

type serdeTestCase struct {
	bin   string
	state int
	value string
}

var undefinedTestCases = []serdeTestCase{
	{`{"Padding1":10,"Padding2":20}`, 0, ""},
	{`{"Padding1":10,"V":null,"Padding2":20}`, 1, ""},
	{`{"Padding1":10,"V":"foo","Padding2":20}`, 2, "foo"},
	{`{"Padding2":20}`, 0, ""},
	{`{"V":null,"Padding2":20}`, 1, ""},
	{`{"V":"bar","Padding2":20}`, 2, "bar"},
	{`{"Padding1":10}`, 0, ""},
	{`{"Padding1":10,"V":null}`, 1, ""},
	{`{"Padding1":10,"V":"baz"}`, 2, "baz"},
}

func TestUnd_serde(t *testing.T) {
	for _, tc := range undefinedTestCases {
		var (
			u2 undV2
			s1 slicesUndV1
			s2 slicesUndV2
		)

		assert.NilError(t, json.Unmarshal([]byte(tc.bin), &s1))
		assert.NilError(t, jsonv2.Unmarshal([]byte(tc.bin), &u2))
		assert.NilError(t, jsonv2.Unmarshal([]byte(tc.bin), &s2))

		assertState(t, u2.V, tc.state, tc.value)
		assertState(t, s1.V, tc.state, tc.value)
		assertState(t, s2.V, tc.state, tc.value)

		var (
			bin []byte
			err error
		)

		bin, err = json.Marshal(s1)
		assert.NilError(t, err)
		assert.Equal(t, string(bin), tc.bin)

		bin, err = jsonv2.Marshal(u2)
		assert.NilError(t, err)
		assert.Equal(t, string(bin), tc.bin)

		bin, err = jsonv2.Marshal(s2)
		assert.NilError(t, err)
		assert.Equal(t, string(bin), tc.bin)
	}
}

type iund[T any] interface {
	IsUndefined() bool
	IsNull() bool
	IsDefined() bool
	Value() T
}

func assertState[T iund[U], U any](t *testing.T, u T, state int, v U) {
	t.Helper()
	switch state {
	case 0:
		assert.Assert(t, u.IsUndefined())
	case 1:
		assert.Assert(t, u.IsNull())
	case 2:
		assert.Assert(t, u.IsDefined())
	}
	assert.Equal(t, u.Value(), v)
}
