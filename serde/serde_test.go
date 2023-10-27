package serde_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ngicks/und/v2/serde"
)

type undefinedableInt struct {
	V int
}

func (u undefinedableInt) IsUndefined() bool {
	return u.V == 0
}

func und(v int) undefinedableInt {
	return undefinedableInt{V: v}
}

type skippable struct {
	Foo string
	Bar undefinedableInt
	Baz undefinedableInt `json:",omitempty"`
	Qux undefinedableInt `json:"qux,omitempty"`
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
				Bar: und(1),
				Baz: und(2),
				Qux: und(3),
			},
			serialized: `{"Foo":"foo","Bar":{"V":1},"Baz":{"V":2},"qux":{"V":3}}`,
		},
		{
			parsed:     skippable{},
			serialized: `{"Foo":""}`,
		},
	} {
		// MarshalJSON
		{
			bin, err := serde.Marshal(tc.parsed)
			if err != nil {
				t.Errorf("must not error: %+v", err)
			}

			if diff := cmp.Diff(tc.serialized, string(bin)); diff != "" {
				t.Errorf("not equal. diff = %s", diff)
			}

			var s skippable
			err = serde.Unmarshal([]byte(tc.serialized), &s)
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
		Bar undefinedableInt `json:"bar"`
		Baz int              `json:",omitempty"`
	}
	type skippableNested struct {
		Foo    undefinedableInt
		Nested nested
	}
	type testCase struct {
		parsed     skippableNested
		serialized string
	}

	for _, tc := range []testCase{
		{
			parsed: skippableNested{
				Foo: und(1),
				Nested: nested{
					Bar: und(2),
					Baz: 1,
				},
			},
			serialized: `{"Foo":{"V":1},"Nested":{"bar":{"V":2},"Baz":1}}`,
		},
		{
			parsed: skippableNested{
				Nested: nested{
					Bar: und(0),
				},
			},
			serialized: `{"Nested":{}}`,
		},
	} {
		bin, err := serde.Marshal(tc.parsed)
		if err != nil {
			t.Errorf("must not error: %+v", err)
		}

		if diff := cmp.Diff(tc.serialized, string(bin)); diff != "" {
			t.Errorf("not equal. diff = %s", diff)
		}

		var s skippableNested
		err = serde.Unmarshal([]byte(tc.serialized), &s)
		if err != nil {
			t.Errorf("must not error: %+v", err)
		}

		if s != tc.parsed {
			t.Errorf("not equal: expected = %+v, actual = %+v, diff = %s", tc.parsed, s, cmp.Diff(tc.parsed, s))
		}
	}
}
