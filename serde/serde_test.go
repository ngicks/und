package serde_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ngicks/und/serde"
	und "github.com/ngicks/und/undefinedable"
)

type skippable struct {
	Foo string
	Bar und.Undefinedable[string]
	Baz und.Undefinedable[string] `json:",omitempty"`
	Qux und.Undefinedable[string] `json:"qux,omitempty"`
}

func TestSerde(t *testing.T) {
	type testCase struct {
		parsed     skippable
		serialized string
	}

	for _, tc := range []testCase{
		{
			parsed: skippable{
				Foo: "foo",
				Bar: und.Defined("bar"),
				Baz: und.Defined("baz"),
				Qux: und.Defined("qux"),
			},
			serialized: `{"Foo":"foo","Bar":"bar","Baz":"baz","qux":"qux"}`,
		},
		{
			parsed: skippable{
				Bar: und.Null[string](),
				Baz: und.Null[string](),
				Qux: und.Null[string](),
			},
			serialized: `{"Foo":"","Bar":null,"Baz":null,"qux":null}`,
		},
		{
			parsed:     skippable{},
			serialized: `{"Foo":""}`,
		},
	} {
		// MarshalJSON
		{
			bin, err := serde.MarshalJSON(tc.parsed)
			if err != nil {
				t.Errorf("must not error: %+v", err)
			}

			if diff := cmp.Diff(tc.serialized, string(bin)); diff != "" {
				t.Errorf("not equal. diff = %s", diff)
			}

			var s skippable
			err = serde.UnmarshalJSON([]byte(tc.serialized), &s)
			if err != nil {
				t.Errorf("must not error: %+v", err)
			}

			if s != tc.parsed {
				t.Errorf("not equal: expected = %+v, actual = %+v", tc.parsed, s)
			}
		}
		// Stream
		{
			buf := new(bytes.Buffer)
			enc := serde.NewEncoder(buf)
			err := enc.Encode(tc.parsed)
			if err != nil {
				t.Errorf("must not error: %+v", err)
			}

			bin := buf.Bytes()
			if diff := cmp.Diff(tc.serialized+"\n", buf.String()); diff != "" {
				t.Errorf("not equal. diff = %s", diff)
			}

			dec := serde.NewDecoder(bytes.NewBuffer(bin))
			var s skippable
			err = dec.Decode(&s)
			if err != nil {
				t.Errorf("must not error: %+v", err)
			}

			if s != tc.parsed {
				t.Errorf("not equal: expected = %+v, actual = %+v", tc.parsed, s)
			}
		}
	}
}

func TestSerde_nested(t *testing.T) {
	type nested struct {
		Bar und.Undefinedable[int] `json:"bar"`
		Baz int                    `json:",omitempty"`
	}
	type skippableNested struct {
		Foo     und.Undefinedable[string]
		Nested  nested
		Nested2 und.Undefinedable[nested]
	}
	type testCase struct {
		parsed     skippableNested
		serialized string
	}

	for _, tc := range []testCase{
		{
			parsed: skippableNested{
				Foo: und.Defined("foo"),
				Nested: nested{
					Bar: und.Defined(0),
					Baz: 1,
				},
				Nested2: und.Defined(nested{
					Bar: und.Defined(123),
					Baz: 333,
				}),
			},
			serialized: `{"Foo":"foo","Nested":{"bar":0,"Baz":1},"Nested2":{"bar":123,"Baz":333}}`,
		},
		{
			parsed: skippableNested{
				Nested: nested{
					Bar: und.Null[int](),
					Baz: 0,
				},
				Nested2: und.Defined(nested{
					Bar: und.Null[int](),
					Baz: 0,
				}),
			},
			serialized: `{"Nested":{"bar":null},"Nested2":{"bar":null}}`,
		},
		{
			parsed: skippableNested{
				Nested: nested{
					Bar: und.Undefined[int](),
				},
				Nested2: und.Null[nested](),
			},
			serialized: `{"Nested":{},"Nested2":null}`,
		},
		{
			parsed:     skippableNested{},
			serialized: `{"Nested":{}}`,
		},
	} {
		bin, err := serde.MarshalJSON(tc.parsed)
		if err != nil {
			t.Errorf("must not error: %+v", err)
		}

		if diff := cmp.Diff(tc.serialized, string(bin)); diff != "" {
			t.Errorf("not equal. diff = %s", diff)
		}

		var s skippableNested
		err = serde.UnmarshalJSON([]byte(tc.serialized), &s)
		if err != nil {
			t.Errorf("must not error: %+v", err)
		}

		if s != tc.parsed {
			t.Errorf("not equal: expected = %+v, actual = %+v", tc.parsed, s)
		}
	}
}
