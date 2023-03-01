package undefinedable_test

import (
	"fmt"
	"testing"

	"github.com/ngicks/type-param-common/util"
	"github.com/ngicks/und/nullable"
	"github.com/ngicks/und/undefinedable"
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
	l, r  nullable.Nullable[T]
	equal bool
}

func runNullableTests[T any](t *testing.T, pairs []pairNullable[T]) bool {
	t.Helper()
	didError := false
	for idx, testCase := range pairs {
		isEqual := testCase.l.Equal(testCase.r)
		if isEqual != testCase.equal {
			var shouldBe string
			if testCase.equal {
				shouldBe = "be equal"
			} else {
				shouldBe = "not be equal"
			}
			didError = true
			t.Errorf(
				"case number = %d. should %s: type = %T left = %v, right = %v",
				idx, shouldBe, testCase.l, formatValue[T](testCase.l), formatValue[T](testCase.r),
			)
		}
	}

	return !didError
}

type pairUndefinedable[T any] struct {
	l, r  undefinedable.Undefinedable[T]
	equal bool
}

func runUndefinedableTests[T any](t *testing.T, pairs []pairUndefinedable[T]) bool {
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

func formatValue[T any](v interface {
	IsNull() bool
	Value() T
}) string {
	if und, ok := v.(interface{ IsUndefined() bool }); ok && und.IsUndefined() {
		return `<undefined>`
	}
	if v.IsNull() {
		return `<null>`
	} else {
		return fmt.Sprintf("%+v", v.Value())
	}
}

// case 1: comparable.
var caseComparable = []pairNullable[int]{
	{
		nullable.NonNull(123), nullable.NonNull(123),
		true,
	},
	{
		nullable.NonNull(123), nullable.NonNull(224),
		false,
	},
	{
		nullable.Null[int](), nullable.Null[int](),
		true,
	},
	{
		nullable.NonNull(123), nullable.Null[int](),
		false,
	},
	{
		nullable.Null[int](), nullable.NonNull(123),
		false,
	},
}

// case 2: non comparable
var caseNonComparable = []pairNullable[[]string]{
	{
		nullable.NonNull([]string{"foo"}), nullable.NonNull([]string{"foo"}),
		false,
	},
	{
		nullable.NonNull([]string{"foo"}), nullable.NonNull([]string{"bar"}),
		false,
	},
	{
		nullable.Null[[]string](), nullable.Null[[]string](),
		true,
	},
	{
		nullable.NonNull([]string{"foo"}), nullable.Null[[]string](),
		false,
	},
	{
		nullable.Null[[]string](), nullable.NonNull([]string{"foo"}),
		false,
	},
}

var sampleSlice = []string{"foo", "bar", "baz"}

// case 3: pointer value
var casePointer = []pairNullable[*[]string]{
	{
		nullable.NonNull(&[]string{"foo"}), nullable.NonNull(&[]string{"foo"}),
		false,
	},
	{
		nullable.NonNull(&[]string{"foo"}), nullable.NonNull(&[]string{"bar"}),
		false,
	},
	{ // same pointer = true (of course).
		nullable.NonNull(&sampleSlice), nullable.NonNull(&sampleSlice),
		true,
	},
	{
		nullable.Null[*[]string](), nullable.Null[*[]string](),
		true,
	},
	{
		nullable.NonNull(&[]string{"foo"}), nullable.Null[*[]string](),
		false,
	},
	{
		nullable.Null[*[]string](), nullable.NonNull(&[]string{"foo"}),
		false,
	},
}

// case 4: non comparable but implements Equality.
var caseNonComparableButCustomEquality = []pairNullable[NonComparableButEquality]{
	{
		nullable.NonNull(NonComparableButEquality{"foo"}), nullable.NonNull(NonComparableButEquality{"foo"}),
		true,
	},
	{
		nullable.NonNull(NonComparableButEquality{"foo"}), nullable.NonNull(NonComparableButEquality{"bar"}),
		false,
	},
	{
		nullable.Null[NonComparableButEquality](), nullable.Null[NonComparableButEquality](),
		true,
	},
	{
		nullable.NonNull(NonComparableButEquality{"foo"}), nullable.Null[NonComparableButEquality](),
		false,
	},
	{
		nullable.Null[NonComparableButEquality](), nullable.NonNull(NonComparableButEquality{"foo"}),
		false,
	},
}

// case 5: comparable but has customized equality.
var caseComparableButCustomEquality = []pairNullable[CustomizedEquality]{
	{
		nullable.NonNull(CustomizedEquality{util.Escape(123)}), nullable.NonNull(CustomizedEquality{util.Escape(123)}),
		true,
	},
	{ // uses customized equality method
		nullable.NonNull(CustomizedEquality{util.Escape(1)}), nullable.NonNull(CustomizedEquality{util.Escape(31)}),
		true,
	},
	{
		nullable.NonNull(CustomizedEquality{util.Escape(123)}), nullable.NonNull(CustomizedEquality{util.Escape(124)}),
		false,
	},
	{
		nullable.Null[CustomizedEquality](), nullable.Null[CustomizedEquality](),
		true,
	},
	{
		nullable.NonNull(CustomizedEquality{util.Escape(123)}), nullable.Null[CustomizedEquality](),
		false,
	},
	{
		nullable.Null[CustomizedEquality](), nullable.NonNull(CustomizedEquality{util.Escape(123)}),
		false,
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
			undefinedable.Undefined[int](), undefinedable.Undefined[int](),
			true,
		},
		// undefined - value
		{
			undefinedable.Defined(123), undefinedable.Undefined[int](),
			false,
		}, {
			undefinedable.Undefined[int](), undefinedable.Defined(123),
			false,
		},
		// undefined - null
		{
			undefinedable.Undefined[int](), undefinedable.Null[int](),
			false,
		},
		{
			undefinedable.Null[int](), undefinedable.Undefined[int](),
			false,
		},
	})
}
func convertNullableCasesToUndefined[T any](cases []pairNullable[T]) []pairUndefinedable[T] {
	ret := make([]pairUndefinedable[T], len(cases))

	for idx, testCase := range cases {
		var l undefinedable.Undefinedable[T]
		if testCase.l.IsNull() {
			l = undefinedable.Null[T]()
		} else {
			l = undefinedable.Defined(testCase.l.Value())
		}

		var r undefinedable.Undefinedable[T]
		if testCase.r.IsNull() {
			r = undefinedable.Null[T]()
		} else {
			r = undefinedable.Defined(testCase.r.Value())
		}

		ret[idx] = pairUndefinedable[T]{
			l, r,
			testCase.equal,
		}
	}
	return ret
}
