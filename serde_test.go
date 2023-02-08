package undefinedablejson_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ngicks/undefinedablejson"
	"github.com/stretchr/testify/assert"
)

type testCase[T any] struct {
	value        T
	unmarshalled *T
	bin          []byte
}

type regularTypes struct {
	Foo  string
	Bar  int `json:"-"`
	Baz  int `json:"-,"`
	Qux  int `json:",omitempty"`
	Quux int `json:",string"`
}

func (t regularTypes) Equal(other any) bool {
	v := other.(regularTypes)
	// regularTypes is apparently hashable.
	return t == v
}

type nullable struct {
	Foo  undefinedablejson.Nullable[string]
	Bar  undefinedablejson.Nullable[int] `json:"-"`
	Baz  undefinedablejson.Nullable[int] `json:"-," und:"string"` // und:"string" does same effect as json:"string"
	Qux  undefinedablejson.Nullable[int] `json:",omitempty"`      // it ignores omitempty
	Quux undefinedablejson.Nullable[int] `und:",string"`
}

func (t nullable) Equal(other any) bool {
	v := other.(nullable)
	return t.Foo.Equal(v.Foo) &&
		t.Bar.Equal(v.Bar) &&
		t.Baz.Equal(v.Baz) &&
		t.Qux.Equal(v.Qux) &&
		t.Quux.Equal(v.Quux)
}

type fields struct {
	Foo undefinedablejson.Undefinedable[string]
	Bar undefinedablejson.Undefinedable[int] `json:"-"`
	Baz undefinedablejson.Undefinedable[int] `json:"-," und:"string"` // und:"string" does same effect as json:"string"
	// it does ignore omitempty, since it should be set as undefined when the field should be skipped.
	Qux undefinedablejson.Undefinedable[int] `json:",omitempty"`
	// json:"string" quotes any type. it ignores type-restriction employed by encoding/json. This is just for simplicity.
	// Just do not use it. Use und:"string" instead.
	Quux undefinedablejson.Undefinedable[int] `und:",string"`
}

func (t fields) Equal(other any) bool {
	v := other.(fields)
	return t.Foo.Equal(v.Foo) &&
		t.Bar.Equal(v.Bar) &&
		t.Baz.Equal(v.Baz) &&
		t.Qux.Equal(v.Qux) &&
		t.Quux.Equal(v.Quux)
}

var regularTypeCase = []testCase[regularTypes]{
	{
		regularTypes{"foo", 123, 456, 0, 789},
		&regularTypes{"foo", 0, 456, 0, 789},
		[]byte(`{"Foo":"foo","-":456,"Quux":"789"}`),
	},
	{
		regularTypes{"foofoo", 123123, 456456, 111, 0},
		&regularTypes{"foofoo", 0, 456456, 111, 0},
		[]byte(`{"Foo":"foofoo","-":456456,"Qux":111,"Quux":"0"}`),
	},
}

var nullableCases = []testCase[nullable]{
	{
		nullable{
			undefinedablejson.NonNull("foo"),
			undefinedablejson.NonNull(123),
			undefinedablejson.NonNull(456),
			undefinedablejson.NonNull(0),
			undefinedablejson.NonNull(789),
		},
		&nullable{
			undefinedablejson.NonNull("foo"),
			undefinedablejson.Null[int](),
			undefinedablejson.NonNull(456),
			undefinedablejson.NonNull(0),
			undefinedablejson.NonNull(789),
		},
		[]byte(`{"Foo":"foo","-":"456","Qux":0,"Quux":"789"}`),
	},
	{
		nullable{
			undefinedablejson.NonNull("foofoo"),
			undefinedablejson.NonNull(123123),
			undefinedablejson.NonNull(456456),
			undefinedablejson.NonNull(111),
			undefinedablejson.NonNull(0),
		},
		&nullable{
			undefinedablejson.NonNull("foofoo"),
			undefinedablejson.Null[int](),
			undefinedablejson.NonNull(456456),
			undefinedablejson.NonNull(111),
			undefinedablejson.NonNull(0),
		},
		[]byte(`{"Foo":"foofoo","-":"456456","Qux":111,"Quux":"0"}`),
	},
	{
		nullable{
			undefinedablejson.NonNull("foofoo"),
			undefinedablejson.Null[int](),
			undefinedablejson.Null[int](),
			undefinedablejson.NonNull(111),
			undefinedablejson.NonNull(0),
		},
		&nullable{
			undefinedablejson.NonNull("foofoo"),
			undefinedablejson.Null[int](),
			undefinedablejson.Null[int](),
			undefinedablejson.NonNull(111),
			undefinedablejson.NonNull(0),
		},
		[]byte(`{"Foo":"foofoo","-":null,"Qux":111,"Quux":"0"}`),
	},
	{
		nullable{
			undefinedablejson.Null[string](),
			undefinedablejson.Null[int](),
			undefinedablejson.Null[int](),
			undefinedablejson.Null[int](), // it ignores omitempty
			undefinedablejson.Null[int](), // it ignores type
		},
		nil,
		[]byte(`{"Foo":null,"-":null,"Qux":null,"Quux":null}`),
	},
}

var fieldsCases = []testCase[fields]{
	{
		fields{
			undefinedablejson.Field("foo"),
			undefinedablejson.Field(123),
			undefinedablejson.Field(456),
			undefinedablejson.Field(0),
			undefinedablejson.Field(789),
		},
		&fields{
			undefinedablejson.Field("foo"),
			undefinedablejson.UndefinedField[int](),
			undefinedablejson.Field(456),
			undefinedablejson.Field(0),
			undefinedablejson.Field(789),
		},
		[]byte(`{"Foo":"foo","-":"456","Qux":0,"Quux":"789"}`),
	},
	{
		fields{
			undefinedablejson.Field("foofoo"),
			undefinedablejson.Field(123123),
			undefinedablejson.Field(456456),
			undefinedablejson.Field(111),
			undefinedablejson.Field(0),
		},
		&fields{
			undefinedablejson.Field("foofoo"),
			undefinedablejson.UndefinedField[int](),
			undefinedablejson.Field(456456),
			undefinedablejson.Field(111),
			undefinedablejson.Field(0),
		},
		[]byte(`{"Foo":"foofoo","-":"456456","Qux":111,"Quux":"0"}`),
	},
	{
		fields{
			undefinedablejson.NullField[string](),
			undefinedablejson.NullField[int](),
			undefinedablejson.NullField[int](),
			undefinedablejson.NullField[int](),
			undefinedablejson.NullField[int](), // it ignores type
		},
		&fields{
			undefinedablejson.NullField[string](),
			undefinedablejson.UndefinedField[int](),
			undefinedablejson.NullField[int](),
			undefinedablejson.NullField[int](),
			undefinedablejson.NullField[int](), // it ignores type
		},
		[]byte(`{"Foo":null,"-":null,"Qux":null,"Quux":null}`),
	},
	{
		fields{
			undefinedablejson.Field("aaa"),
			undefinedablejson.UndefinedField[int](),
			undefinedablejson.Field(111),
			undefinedablejson.NullField[int](),
			undefinedablejson.UndefinedField[int](),
		},
		nil,
		[]byte(`{"Foo":"aaa","-":"111","Qux":null}`),
	},
	{
		fields{
			undefinedablejson.UndefinedField[string](),
			undefinedablejson.UndefinedField[int](),
			undefinedablejson.UndefinedField[int](),
			undefinedablejson.UndefinedField[int](),
			undefinedablejson.UndefinedField[int](),
		},
		nil,
		[]byte(`{}`),
	},
}

func assertMarshalJSON[T any](t *testing.T, testCase testCase[T]) bool {
	t.Helper()
	bin, err := undefinedablejson.MarshalFieldsJSON(testCase.value)
	if err != nil {
		t.Errorf("err = %T:%+v, input = %+v", err, err, testCase.value)
		return false
	}
	if diff := cmp.Diff(string(bin), string(testCase.bin)); diff != "" {
		t.Errorf(
			"not as expected, expected = %s,\nactual = %s",
			string(testCase.bin), string(bin),
		)
		return false
	}
	return true
}

func assertUnmarshalJSON[T interface{ Equal(other any) bool }](t *testing.T, caseNumber int, testCase testCase[T]) bool {
	t.Helper()
	var v T
	err := undefinedablejson.UnmarshalFieldsJSON(testCase.bin, &v)
	if err != nil {
		t.Errorf(
			"case number = %d: err = %T:%+v, input = %+v",
			caseNumber, err, err, testCase.value,
		)
		return false
	}

	equalityTarget := testCase.value
	if testCase.unmarshalled != nil {
		equalityTarget = *testCase.unmarshalled
	}
	if !v.Equal(equalityTarget) {
		t.Errorf(
			"case number = %d: not equal. expected = %+v, actual = %+v",
			caseNumber, equalityTarget, v,
		)
		return false
	}
	return true
}

func TestMarshalJSON(t *testing.T) {
	for _, testCase := range regularTypeCase {
		assertMarshalJSON(t, testCase)
	}

	for _, testCase := range nullableCases {
		assertMarshalJSON(t, testCase)
	}

	for _, testCase := range fieldsCases {
		assertMarshalJSON(t, testCase)
	}
}

func TestMarshalJSON_shuffled_order(t *testing.T) {
	for i := 0; i < 100; i++ {
		run := false
		if i < len(regularTypeCase) {
			run = true
			assertMarshalJSON(t, regularTypeCase[i])
		}
		if i < len(nullableCases) {
			run = true
			assertMarshalJSON(t, nullableCases[i])
		}
		if i < len(fieldsCases) {
			run = true
			assertMarshalJSON(t, fieldsCases[i])
		}
		if i < len(embedCase) {
			run = true
			assertMarshalJSON(t, embedCase[i])
		}

		if !run {
			break
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {
	for idx, testCase := range regularTypeCase {
		assertUnmarshalJSON(t, idx, testCase)
	}

	for idx, testCase := range nullableCases {
		assertUnmarshalJSON(t, idx, testCase)
	}

	for idx, testCase := range fieldsCases {
		assertUnmarshalJSON(t, idx, testCase)
	}
}

type Embedded struct {
	Foo string
	Bar undefinedablejson.Nullable[string]      `json:"bar"`
	Baz undefinedablejson.Undefinedable[string] `json:"baz"`
}

type Emm string

type Emm2 struct {
	V string
}

func (e Emm2) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"foobar em em %s"`, e.V)), nil
}

type Emm3 int

func (e Emm3) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"%d":"nah"}`, e)), nil
}

type sample struct {
	Emm // embedded non struct.
	Embedded
	Corge  string
	Grault undefinedablejson.Nullable[string]
	Garply undefinedablejson.Undefinedable[string]
}

func (s sample) Equal(otherAny any) bool {
	other := otherAny.(sample)
	return s.Emm == other.Emm &&
		s.Embedded.Foo == other.Embedded.Foo &&
		s.Embedded.Bar.Equal(other.Embedded.Bar) &&
		s.Embedded.Baz.Equal(other.Embedded.Baz) &&
		s.Corge == other.Corge &&
		s.Grault.Equal(other.Grault) &&
		s.Garply.Equal(other.Garply)
}

var embedCase = []testCase[sample]{
	{
		value: sample{
			Emm: Emm("emm"),
			Embedded: Embedded{
				Foo: "aaa",
				Bar: undefinedablejson.NonNull("bar"),
				Baz: undefinedablejson.Field("baz"),
			},
			Corge:  "corge",
			Grault: undefinedablejson.Null[string](),
			Garply: undefinedablejson.UndefinedField[string](),
		},
		// non-struct embedded field is marshalled through its onw MarshalJSON implementation.
		bin: []byte(`{"Emm":"emm","Foo":"aaa","bar":"bar",` +
			`"baz":"baz","Corge":"corge","Grault":null}`),
	},
	{
		value: sample{
			Emm: Emm("emm"),
			Embedded: Embedded{
				Foo: "aaa",
				Bar: undefinedablejson.Null[string](),
				Baz: undefinedablejson.UndefinedField[string](),
			},
			Corge:  "corge",
			Grault: undefinedablejson.NonNull("grault"),
			Garply: undefinedablejson.Field("garply"),
		},
		// struct embedded field expanded as inner fields, also skips undefined fields.
		bin: []byte(`{"Emm":"emm","Foo":"aaa","bar":null,` +
			`"Corge":"corge","Grault":"grault","Garply":"garply"}`),
	},
}

func TestMarshalJSON_embedded_field(t *testing.T) {
	for _, tc := range embedCase {
		assertMarshalJSON(t, tc)
	}
}

func TestUnmarshalJSON_embedded_field(t *testing.T) {
	for idx, tc := range embedCase {
		assertUnmarshalJSON(t, idx, tc)
	}
}

func TestMarshalJSON_error(t *testing.T) {
	for _, fn := range []func() (string, error){
		func() (string, error) {
			_, err := undefinedablejson.MarshalFieldsJSON(213)
			return "if input is not a struct type.", err
		},
		func() (string, error) {
			type sample struct {
				Emm2
			}

			_, err := undefinedablejson.MarshalFieldsJSON(sample{Emm2: Emm2{V: "foo"}})
			return "if it find embedded type which implements", err
		},
		func() (string, error) {
			type sample2 struct {
				Emm3
			}

			_, err := undefinedablejson.MarshalFieldsJSON(sample2{Emm3: Emm3(15)})
			return "if it find embedded type which implements", err
		},
	} {
		if cond, err := fn(); !errors.Is(err, undefinedablejson.ErrIncorrectType) {
			t.Errorf(
				`It must return ErrIncorrect %s. type = %T, value = %+v`,
				cond, err, err,
			)
		}
	}
}

const untypedFalse = false
const untypedTrue = true
const untypedZero = 0
const untypedOne = 1

type forTypedNil struct {
}

func (t *forTypedNil) Error() string {
	return ""
}

func TestIsZero(t *testing.T) {
	assert := assert.New(t)

	// quoted from https://pkg.go.dev/encoding/json@go1.19.5:
	//
	// The "omitempty" option specifies that the field should be omitted from the encoding
	// if the field has an empty value, defined as false, 0, a nil pointer, a nil interface value,
	// and any empty array, slice, map, or string.

	for _, v := range []any{
		// false
		untypedFalse, false,
		// 0
		untypedZero,
		int8(0), int16(0), int32(0), int64(0), int(0),
		uint8(0), uint16(0), uint32(0), uint64(0), uint(0),
		float32(0), float64(0), uintptr(0),
		// empty array
		[0]int{}, [0]string{},
		// empty slice, map, or string
		[]int{}, map[string]int{}, "",
	} {
		rv := reflect.ValueOf(v)
		assert.True(undefinedablejson.IsEmpty(rv), "kind = %s, value = %+v", rv.Kind(), rv.Interface())
	}

	for _, rvFunc := range []func() reflect.Value{
		func() reflect.Value { // a nil interface
			var typedNil *forTypedNil
			var err error = typedNil
			return reflect.ValueOf(err)
		},
		func() reflect.Value { // a nil pointer
			var num *int
			return reflect.ValueOf(num)
		},
		func() reflect.Value { // a nil slice
			var sl []string
			return reflect.ValueOf(sl)
		},
		func() reflect.Value { // a nil map
			var mm map[int]string
			return reflect.ValueOf(mm)
		},
	} {
		rv := rvFunc()
		assert.True(undefinedablejson.IsEmpty(rv), "kind = %s, value = %+v", rv.Kind(), rv.Interface())
	}

	for _, v := range []any{
		untypedTrue, true,
		untypedOne,
		int8(1), int16(-2), int32(3), int64(-4), int(5),
		uint8(1), uint16(2), uint32(3), uint64(4), uint(5),
		float32(10), float64(-123), uintptr(23013209),
		[1]int{}, [2]string{},
		[]int{1}, map[string]int{"foo": 1}, "huh?",
	} {
		rv := reflect.ValueOf(v)
		assert.False(undefinedablejson.IsEmpty(rv), "kind = %s, value = %+v", rv.Kind(), rv.Interface())
	}

	for _, rvFunc := range []func() reflect.Value{
		func() reflect.Value {
			var err error = &forTypedNil{}
			return reflect.ValueOf(err)
		},
		func() reflect.Value {
			num := 213
			return reflect.ValueOf(&num)
		},
	} {
		rv := rvFunc()
		assert.False(undefinedablejson.IsEmpty(rv), "kind = %s, value = %+v", rv.Kind(), rv.Interface())
	}
}
