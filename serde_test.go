package undefinedablejson_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ngicks/undefinedablejson"
)

type skippable struct {
	Foo string
	Bar undefinedablejson.Undefinedable[string]
	Baz undefinedablejson.Undefinedable[string] `json:",omitempty"`
	Qux undefinedablejson.Undefinedable[string] `json:"qux,omitempty"`
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
				Bar: undefinedablejson.Field("bar"),
				Baz: undefinedablejson.Field("baz"),
				Qux: undefinedablejson.Field("qux"),
			},
			serialized: `{"Foo":"foo","Bar":"bar","Baz":"baz","qux":"qux"}`,
		},
		{
			parsed: skippable{
				Bar: undefinedablejson.NullField[string](),
				Baz: undefinedablejson.NullField[string](),
				Qux: undefinedablejson.NullField[string](),
			},
			serialized: `{"Foo":"","Bar":null,"Baz":null,"qux":null}`,
		},
		{
			parsed:     skippable{},
			serialized: `{"Foo":""}`,
		},
	} {
		bin, err := undefinedablejson.MarshalFieldsJSON(tc.parsed)
		if err != nil {
			t.Errorf("must not error: %+v", err)
		}

		if diff := cmp.Diff(tc.serialized, string(bin)); diff != "" {
			t.Errorf("not equal. diff = %s", diff)
		}

		var s skippable
		err = undefinedablejson.UnmarshalFieldsJSON([]byte(tc.serialized), &s)
		if err != nil {
			t.Errorf("must not error: %+v", err)
		}

		if s != tc.parsed {
			t.Errorf("not equal: expected = %+v, actual = %+v", tc.parsed, s)
		}
	}
}

func TestSerde_nested(t *testing.T) {
	type nested struct {
		Bar undefinedablejson.Undefinedable[int] `json:"bar"`
		Baz int                                  `json:",omitempty"`
	}
	type skippableNested struct {
		Foo     undefinedablejson.Undefinedable[string]
		Nested  nested
		Nested2 undefinedablejson.Undefinedable[nested]
	}
	type testCase struct {
		parsed     skippableNested
		serialized string
	}

	for _, tc := range []testCase{
		{
			parsed: skippableNested{
				Foo: undefinedablejson.Field("foo"),
				Nested: nested{
					Bar: undefinedablejson.Field(0),
					Baz: 1,
				},
				Nested2: undefinedablejson.Field(nested{
					Bar: undefinedablejson.Field(123),
					Baz: 333,
				}),
			},
			serialized: `{"Foo":"foo","Nested":{"bar":0,"Baz":1},"Nested2":{"bar":123,"Baz":333}}`,
		},
		{
			parsed: skippableNested{
				Nested: nested{
					Bar: undefinedablejson.NullField[int](),
					Baz: 0,
				},
				Nested2: undefinedablejson.Field(nested{
					Bar: undefinedablejson.NullField[int](),
					Baz: 0,
				}),
			},
			serialized: `{"Nested":{"bar":null},"Nested2":{"bar":null}}`,
		},
		{
			parsed: skippableNested{
				Nested: nested{
					Bar: undefinedablejson.UndefinedField[int](),
				},
				Nested2: undefinedablejson.NullField[nested](),
			},
			serialized: `{"Nested":{},"Nested2":null}`,
		},
		{
			parsed:     skippableNested{},
			serialized: `{"Nested":{}}`,
		},
	} {
		bin, err := undefinedablejson.MarshalFieldsJSON(tc.parsed)
		if err != nil {
			t.Errorf("must not error: %+v", err)
		}

		if diff := cmp.Diff(tc.serialized, string(bin)); diff != "" {
			t.Errorf("not equal. diff = %s", diff)
		}

		var s skippableNested
		err = undefinedablejson.UnmarshalFieldsJSON([]byte(tc.serialized), &s)
		if err != nil {
			t.Errorf("must not error: %+v", err)
		}

		if s != tc.parsed {
			t.Errorf("not equal: expected = %+v, actual = %+v", tc.parsed, s)
		}
	}
}
