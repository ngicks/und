package structtag

import (
	"reflect"
	"strconv"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestTag(t *testing.T) {
	type testCase struct {
		input    reflect.StructTag
		tag      string
		option   string
		value    string
		expected string
	}

	for _, tc := range []testCase{
		{
			input:    `json:"foo"`,
			tag:      `json`,
			option:   `omitempty`,
			expected: `json:"foo,omitempty"`,
		},
		{
			input:    `json:"'\\xde\\xad\\xbe\\xef'"`,
			tag:      `json`,
			option:   `omitzero`,
			expected: `json:"'\\xde\\xad\\xbe\\xef',omitzero"`,
		},
		{
			input:    `json:",omitzero"`,
			tag:      `json`,
			option:   `omitempty`,
			expected: `json:",omitzero,omitempty"`,
		},
		{
			input:    `json:",omitzero"`,
			tag:      `json`,
			option:   `omitzero`,
			expected: `json:",omitzero"`,
		},
		{
			input:    `json:",omitempty"`,
			tag:      `json`,
			option:   `format`,
			value:    `booboo`,
			expected: `json:",omitempty,format:booboo"`,
		},
		{
			input:    `json:",format:fizzbuzz"`,
			tag:      `json`,
			option:   `format`,
			value:    `booboo`,
			expected: `json:",format:fizzbuzz"`,
		},
		{
			input:    `json:",format:fizzbuzz"`,
			tag:      `json`,
			option:   `omitempty`,
			expected: `json:",format:fizzbuzz,omitempty"`,
		},
		{
			input:    `json:"foo"`,
			tag:      `bar`,
			option:   `baz`,
			expected: `json:"foo" bar:",baz"`,
		},
		{
			input:    `json:"foo" bar:",foo"`,
			tag:      `bar`,
			option:   `baz`,
			expected: `json:"foo" bar:",foo,baz"`,
		},
	} {
		tags, err := ParseStructTag(tc.input)
		assert.NilError(t, err)
		added, err := tags.AddOption(tc.tag, tc.option, tc.value)
		assert.NilError(t, err)
		assert.Assert(t, cmp.Equal(reflect.StructTag(tc.expected), added.StructTag()))
	}
}

func Test_getRange(t *testing.T) {
	tag := "'\\xde\\xad\\xbe\\xef',omitzero,format:RFC3339,omitempty"

	var (
		n, m      int
		unescaped string
		err       error
	)
	n, m, unescaped, err = getRange(tag, "")
	s, _ := strconv.Unquote("\"\\xde\\xad\\xbe\\xef\"")
	assert.NilError(t, err)
	assert.Equal(t, n, 0)
	assert.Equal(t, m, 18)
	assert.Equal(t, unescaped, s)

	n, m, unescaped, err = getRange(tag, "omitzero")
	assert.NilError(t, err)
	assert.Equal(t, n, 19)
	assert.Equal(t, m, 19+len("omitzero"))
	assert.Equal(t, unescaped, "")

	n, m, unescaped, err = getRange(tag, "format")
	assert.NilError(t, err)
	assert.Equal(t, n, 28)
	assert.Equal(t, m, 28+len("format:RFC3339"))
	assert.Equal(t, unescaped, "")

	n, m, unescaped, err = getRange(tag, "omitempty")
	assert.NilError(t, err)
	assert.Equal(t, n, 43)
	assert.Equal(t, m, 43+len("omitempty"))
	assert.Equal(t, unescaped, "")

	n, m, unescaped, err = getRange(tag, "foo")
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Equal(t, n, -1)
	assert.Equal(t, m, -1)
	assert.Equal(t, unescaped, "")
}

func TestTags_DeleteOption(t *testing.T) {
	tag := `json:"'\\xde\\xad\\xbe\\xef',omitzero,format:RFC3339,omitempty"`

	type testCase struct {
		opt    string
		output string
	}

	for _, tc := range []testCase{
		{
			opt:    "",
			output: `json:",omitzero,format:RFC3339,omitempty"`,
		},
		{
			opt:    "omitzero",
			output: `json:"'\\xde\\xad\\xbe\\xef',format:RFC3339,omitempty"`,
		},
		{
			opt:    "format",
			output: `json:"'\\xde\\xad\\xbe\\xef',omitzero,omitempty"`,
		},
		{
			opt:    "omitempty",
			output: `json:"'\\xde\\xad\\xbe\\xef',omitzero,format:RFC3339"`,
		},
	} {
		t.Run(tc.opt, func(t *testing.T) {
			tags, err := ParseStructTag(reflect.StructTag(tag))
			assert.NilError(t, err)
			tt, err := tags.DeleteOption("json", tc.opt)
			assert.NilError(t, err)
			assert.Equal(t, string(tt.StructTag()), tc.output)
		})
	}
}

func TestTags_AddOption(t *testing.T) {
	type testCase struct {
		option   string
		value    string
		input    string
		expected string
	}

	for _, tc := range []testCase{
		{
			option:   "",
			value:    "'\\xde\\xad\\xbe\\xef'",
			input:    ``,
			expected: `json:"'\\xde\\xad\\xbe\\xef'"`,
		},
		{
			option:   "",
			value:    "'\\xde\\xad\\xbe\\xef'",
			input:    `json:",omitzero,format:RFC3339,omitempty"`,
			expected: `json:"'\\xde\\xad\\xbe\\xef',omitzero,format:RFC3339,omitempty"`,
		},
		{
			option:   "omitzero",
			input:    `json:"'\\xde\\xad\\xbe\\xef',format:RFC3339,omitempty"`,
			expected: `json:"'\\xde\\xad\\xbe\\xef',format:RFC3339,omitempty,omitzero"`,
		},
		{
			option:   "format",
			value:    "RFC3339",
			input:    `json:"'\\xde\\xad\\xbe\\xef',omitzero,omitempty"`,
			expected: `json:"'\\xde\\xad\\xbe\\xef',omitzero,omitempty,format:RFC3339"`,
		},
		{
			option:   "omitempty",
			input:    `json:"'\\xde\\xad\\xbe\\xef',omitzero,format:RFC3339"`,
			expected: `json:"'\\xde\\xad\\xbe\\xef',omitzero,format:RFC3339,omitempty"`,
		},
		{
			option:   "format",
			input:    `json:"'\\xde\\xad\\xbe\\xef',omitzero,format:RFC3339"`,
			expected: `json:"'\\xde\\xad\\xbe\\xef',omitzero,format:RFC3339"`,
		},
	} {
		t.Run(tc.option+":"+tc.value, func(t *testing.T) {
			tags, err := ParseStructTag(reflect.StructTag(tc.input))
			assert.NilError(t, err)
			tt, err := tags.AddOption("json", tc.option, tc.value)
			assert.NilError(t, err)
			assert.Equal(t, string(tt.StructTag()), tc.expected)
		})
	}
}
