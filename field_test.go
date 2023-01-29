package undefinedablejson_test

import (
	"fmt"
	"testing"

	"github.com/ngicks/type-param-common/util"
	"github.com/ngicks/undefinedablejson"
)

type CustomizedEquality struct {
	V *int
}

func (e CustomizedEquality) Equal(other CustomizedEquality) bool {
	// Does it look something like ring buffer index?
	return *e.V%30 == *other.V%30
}

type NonComparableButEquality []string

func (e NonComparableButEquality) Equal(other NonComparableButEquality) bool {
	if len(e) != len(other) {
		return false
	}
	for idx := range e {
		if e[idx] != other[idx] {
			return false
		}
	}
	return true
}

type pairNullable[T any] struct {
	l, r                undefinedablejson.Nullable[T]
	equal, sameInstance bool
}

func runNullableTests[T any](t *testing.T, pairs []pairNullable[T]) bool {
	t.Helper()
	didError := false
	for idx, testCase := range pairs {
		isEqual := testCase.l.Equal(testCase.r)
		if isEqual != testCase.equal {
			didError = true
			t.Errorf(
				"case number = %d. not equal: type = %T left = %v, right = %v",
				idx, testCase.l, formatValue[T](testCase.l), formatValue[T](testCase.r),
			)
		}
	}

	return !didError
}

type pairUndefinedable[T any] struct {
	l, r           undefinedablejson.Undefinedable[T]
	equal, ignored bool
}

func runUndefinedableTests[T any](t *testing.T, pairs []pairUndefinedable[T]) bool {
	t.Helper()
	didError := false
	for idx, testCase := range pairs {
		// same instance test need to be ignored: conversion may not correctly place same instance.
		if testCase.ignored {
			continue
		}

		isEqual := testCase.l.Equal(testCase.r)
		if isEqual != testCase.equal {
			didError = true
			t.Errorf(
				"case number = %d. not equal: type = %T left = %v, right = %v",
				idx, testCase.l, formatValue[T](testCase.l), formatValue[T](testCase.r),
			)
		}
	}

	return !didError
}

func formatValue[T any](v interface{ Value() *T }) string {
	if val := v.Value(); val == nil {
		return `<nil>`
	} else {
		return fmt.Sprintf("%+v", *val)
	}
}

// case 1: comparable.
var nullableInt = undefinedablejson.NonNull(0)
var caseComparable = []pairNullable[int]{
	{
		nullableInt, nullableInt,
		true, true,
	},
	{
		undefinedablejson.NonNull(123), undefinedablejson.NonNull(123),
		true, false,
	},
	{
		undefinedablejson.NonNull(123), undefinedablejson.NonNull(224),
		false, false,
	},
	{
		undefinedablejson.Null[int](), undefinedablejson.Null[int](),
		true, false,
	},
	{
		undefinedablejson.NonNull(123), undefinedablejson.Null[int](),
		false, false,
	},
	{
		undefinedablejson.Null[int](), undefinedablejson.NonNull(123),
		false, false,
	},
}

var nullableSlice = undefinedablejson.NonNull([]string{})

// case 2: non comparable
var caseNonComparable = []pairNullable[[]string]{
	{
		nullableSlice, nullableSlice,
		true, true,
	},
	{
		undefinedablejson.NonNull([]string{"foo"}), undefinedablejson.NonNull([]string{"foo"}),
		false, false,
	},
	{
		undefinedablejson.NonNull([]string{"foo"}), undefinedablejson.NonNull([]string{"bar"}),
		false, false,
	},
	{
		undefinedablejson.Null[[]string](), undefinedablejson.Null[[]string](),
		true, false,
	},
	{
		undefinedablejson.NonNull([]string{"foo"}), undefinedablejson.Null[[]string](),
		false, false,
	},
	{
		undefinedablejson.Null[[]string](), undefinedablejson.NonNull([]string{"foo"}),
		false, false,
	},
}

var nullableP = undefinedablejson.NonNull(&[]string{})
var sampleSlice = []string{"foo", "bar", "baz"}

// case 3: pointer value
var casePointer = []pairNullable[*[]string]{
	{
		nullableP, nullableP, // same instance (always true)
		true, true,
	},
	{
		undefinedablejson.NonNull(&[]string{"foo"}), undefinedablejson.NonNull(&[]string{"foo"}),
		false, false,
	},
	{
		undefinedablejson.NonNull(&[]string{"foo"}), undefinedablejson.NonNull(&[]string{"bar"}),
		false, false,
	},
	{ // same pointer = true (of course).
		undefinedablejson.NonNull(&sampleSlice), undefinedablejson.NonNull(&sampleSlice),
		true, false,
	},
	{
		undefinedablejson.Null[*[]string](), undefinedablejson.Null[*[]string](),
		true, false,
	},
	{
		undefinedablejson.NonNull(&[]string{"foo"}), undefinedablejson.Null[*[]string](),
		false, false,
	},
	{
		undefinedablejson.Null[*[]string](), undefinedablejson.NonNull(&[]string{"foo"}),
		false, false,
	},
}

var nullableNonComparableButEquality = undefinedablejson.NonNull(NonComparableButEquality{})

// case 4: non comparable but implements Equality.
var caseNonComparableButCustomEquality = []pairNullable[NonComparableButEquality]{
	{
		nullableNonComparableButEquality, nullableNonComparableButEquality, // same instance (always true)
		true, true,
	},
	{
		undefinedablejson.NonNull(NonComparableButEquality{"foo"}), undefinedablejson.NonNull(NonComparableButEquality{"foo"}),
		true, false,
	},
	{
		undefinedablejson.NonNull(NonComparableButEquality{"foo"}), undefinedablejson.NonNull(NonComparableButEquality{"bar"}),
		false, false,
	},
	{
		undefinedablejson.Null[NonComparableButEquality](), undefinedablejson.Null[NonComparableButEquality](),
		true, false,
	},
	{
		undefinedablejson.NonNull(NonComparableButEquality{"foo"}), undefinedablejson.Null[NonComparableButEquality](),
		false, false,
	},
	{
		undefinedablejson.Null[NonComparableButEquality](), undefinedablejson.NonNull(NonComparableButEquality{"foo"}),
		false, false,
	},
}

var nullableCustomizedEquality = undefinedablejson.NonNull(CustomizedEquality{util.Escape(123)})

// case 5: comparable but has customized equality.
var caseComparableButCustomEquality = []pairNullable[CustomizedEquality]{
	{
		nullableCustomizedEquality, nullableCustomizedEquality, // same instance (always true)
		true, true,
	},
	{
		undefinedablejson.NonNull(CustomizedEquality{util.Escape(123)}), undefinedablejson.NonNull(CustomizedEquality{util.Escape(123)}),
		true, false,
	},
	{ // uses customized equality method
		undefinedablejson.NonNull(CustomizedEquality{util.Escape(1)}), undefinedablejson.NonNull(CustomizedEquality{util.Escape(31)}),
		true, false,
	},
	{
		undefinedablejson.NonNull(CustomizedEquality{util.Escape(123)}), undefinedablejson.NonNull(CustomizedEquality{util.Escape(124)}),
		false, false,
	},
	{
		undefinedablejson.Null[CustomizedEquality](), undefinedablejson.Null[CustomizedEquality](),
		true, false,
	},
	{
		undefinedablejson.NonNull(CustomizedEquality{util.Escape(123)}), undefinedablejson.Null[CustomizedEquality](),
		false, false,
	},
	{
		undefinedablejson.Null[CustomizedEquality](), undefinedablejson.NonNull(CustomizedEquality{util.Escape(123)}),
		false, false,
	},
}

func TestFields_equality(t *testing.T) {
	runNullableTests(t, caseComparable)
	runNullableTests(t, caseNonComparable)
	runNullableTests(t, casePointer)
	runNullableTests(t, caseNonComparableButCustomEquality)
	runNullableTests(t, caseComparableButCustomEquality)

	runUndefinedableTests(t, convertNullableCasesToUndefined(caseComparable))
	runUndefinedableTests(t, convertNullableCasesToUndefined(caseNonComparable))
	runUndefinedableTests(t, convertNullableCasesToUndefined(casePointer))
	runUndefinedableTests(t, convertNullableCasesToUndefined(caseNonComparableButCustomEquality))
	runUndefinedableTests(t, convertNullableCasesToUndefined(caseComparableButCustomEquality))

	runUndefinedableTests(t, []pairUndefinedable[int]{
		{ // undefined - undefined
			undefinedablejson.UndefinedField[int](), undefinedablejson.UndefinedField[int](),
			true, false,
		},
		// undefined - value
		{
			undefinedablejson.Field(123), undefinedablejson.UndefinedField[int](),
			false, false,
		}, {
			undefinedablejson.UndefinedField[int](), undefinedablejson.Field(123),
			false, false,
		},
		// undefined - null
		{
			undefinedablejson.UndefinedField[int](), undefinedablejson.NullField[int](),
			false, false,
		},
		{
			undefinedablejson.NullField[int](), undefinedablejson.UndefinedField[int](),
			false, false,
		},
	})
}
func convertNullableCasesToUndefined[T any](cases []pairNullable[T]) []pairUndefinedable[T] {
	ret := make([]pairUndefinedable[T], len(cases))

	for idx, testCase := range cases {
		var l undefinedablejson.Undefinedable[T]
		if v := testCase.l.Value(); v == nil {
			l = undefinedablejson.NullField[T]()
		} else {
			l = undefinedablejson.Field(*v)
		}

		var r undefinedablejson.Undefinedable[T]
		if v := testCase.r.Value(); v == nil {
			r = undefinedablejson.NullField[T]()
		} else {
			r = undefinedablejson.Field(*v)
		}

		ret[idx] = pairUndefinedable[T]{
			l, r,
			testCase.equal, testCase.sameInstance,
		}
	}
	return ret
}
