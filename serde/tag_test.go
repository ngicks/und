package serde_test

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ngicks/und/serde"
)

func FuzzTag(f *testing.F) {
	f.Add(`json:"foo,string,omitempty"`)
	f.Fuzz(func(t *testing.T, tag string) {
		tags, err := serde.ParseStructTag(reflect.StructTag(tag))
		if err != nil {
			t.Skip()
		}

		tags2, err := serde.ParseStructTag(serde.FlattenStructTag(tags))
		if err != nil {
			t.Fatalf("must not return error. err = %+v", err)
		}
		if diff := cmp.Diff(tags, tags2); diff != "" {
			t.Fatalf("must Parse back to same tags but diff is not empty, diff = %s", diff)
		}
	})
}

func TestTag(t *testing.T) {
	tags, err := serde.ParseStructTag(reflect.StructTag(`foo:"bar" baz:",qux,,," json:"foo,omitempty,string"`))

	if err != nil {
		t.Fatalf("must not return error. err = %+v, input = %+v", err, tags)
	}
	if diff := cmp.Diff(
		tags,
		[]serde.Tag{
			{Key: "foo", Value: "bar"},
			{Key: "baz", Value: ",qux,,,"},
			{Key: "json", Value: "foo,omitempty,string"},
		},
	); diff != "" {
		t.Fatalf("not equal. diff = %s", diff)
	}
}

func TestTag_error(t *testing.T) {
	for _, tags := range []reflect.StructTag{
		`json:"foo`,
		`json`,
		`json::`,
	} {
		_, err := serde.ParseStructTag(tags)
		if err == nil {
			t.Fatalf("must return error for incorrectly formatted struct tag. input = %s", tags)
		}
	}
}

func TestFakeOmitempty(t *testing.T) {
	for _, tagPair := range [][2]reflect.StructTag{
		{`json:"foo,string,omitempty"`, `json:"foo,string,omitempty"`},
		{`json:"foo,omitempty,string"`, `json:"foo,omitempty,string"`},
		{`json:"foo,string"`, `json:"foo,string,omitempty"`},
		{`json:"foo"`, `json:"foo,omitempty"`},
		{`json:"omitempty"`, `json:"omitempty,omitempty"`},
		{``, `json:",omitempty"`},
		{`foo:"bar" baz:"qux" json:"foo" quux:"corge"`, `foo:"bar" baz:"qux" json:"foo,omitempty" quux:"corge"`},
		{`foo:"bar" baz:"qux" quux:"corge"`, `foo:"bar" baz:"qux" quux:"corge" json:",omitempty"`},
	} {
		faked := serde.FakeOmitempty(tagPair[0])
		if faked != tagPair[1] {
			t.Errorf(
				"not equal. expected = %s, actual = %s. diff = %s",
				tagPair[1], faked, cmp.Diff([]byte(tagPair[1]), []byte(faked)),
			)
		}
	}
}
