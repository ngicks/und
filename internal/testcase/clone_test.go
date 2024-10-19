package testcase_test

import (
	"reflect"
	"slices"
	"testing"
	"unsafe"

	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
)

var (
	_ option.Cloner[option.Option[any]]        = option.Option[any]{}
	_ option.Cloner[option.Options[any]]       = option.Options[any]{}
	_ option.Cloner[und.Und[any]]              = und.Und[any]{}
	_ option.Cloner[sliceund.Und[any]]         = sliceund.Und[any]{}
	_ option.Cloner[elastic.Elastic[any]]      = elastic.Elastic[any]{}
	_ option.Cloner[sliceelastic.Elastic[any]] = sliceelastic.Elastic[any]{}
)

type clonable []int

func (c clonable) Clone() clonable {
	return slices.Clone(c)
}

func TestCloner(t *testing.T) {
	clonerTetSet[[]int](t, false)
	clonerTetSet[clonable](t, true)
}

func clonerTetSet[T ~[]U, U int](t *testing.T, shouldDiffer bool) {
	testCloner(
		t,
		shouldDiffer,
		option.Some(T{1, 2, 3}),
		testDifferenceValuer,
	)
	testCloner(
		t,
		shouldDiffer,
		option.Options[T]{option.Some(T{1, 2, 3}), option.Some(T{4, 5, 6})},
		testDifferenceOptions,
	)
	testCloner(
		t,
		shouldDiffer,
		und.Defined(T{1, 2, 3}),
		testDifferenceValuer,
	)
	testCloner(
		t,
		shouldDiffer,
		sliceund.Defined(T{1, 2, 3}),
		testDifferenceValuer,
	)
	testCloner(
		t,
		shouldDiffer,
		elastic.FromOptions(option.Some(T{1, 2, 3}), option.Some(T{4, 5, 6})),
		func(a, b elastic.Elastic[T]) (bool, unsafe.Pointer, unsafe.Pointer) {
			return testDifferenceOptions(a.Unwrap().Value(), b.Unwrap().Value())
		},
	)
	testCloner(
		t,
		shouldDiffer,
		sliceelastic.FromOptions(option.Some(T{1, 2, 3}), option.Some(T{4, 5, 6})),
		func(a, b sliceelastic.Elastic[T]) (bool, unsafe.Pointer, unsafe.Pointer) {
			return testDifferenceOptions(a.Unwrap().Value(), b.Unwrap().Value())
		},
	)
}

func testCloner[T option.Cloner[U], U any](t *testing.T, shouldDiffer bool, initial T, isDifferent func(a T, b U) (bool, unsafe.Pointer, unsafe.Pointer)) {
	t.Helper()
	cloned := initial.Clone()
	diff, ap, bp := isDifferent(initial, cloned)
	if diff != shouldDiffer {
		t.Fatalf("not different: left = %#v, right = %#v, left pointer = %p, right pointer = %p", initial, cloned, ap, bp)
	}
}

func testDifferenceOptions[T ~[]U, U any](a, b option.Options[T]) (bool, unsafe.Pointer, unsafe.Pointer) {
	if diff, ap, bp := testDifference(a, b); !diff {
		return diff, ap, bp
	}
	if len(a) != len(b) {
		return false, nil, nil
	}
	for i := range a {
		if diff, ap, bp := testDifference(a[i].Value(), b[i].Value()); !diff {
			return false, ap, bp
		}
	}
	return true, nil, nil
}

func testDifferenceValuer[T interface{ Value() U }, U ~[]V, V any](a, b T) (bool, unsafe.Pointer, unsafe.Pointer) {
	return testDifference(a.Value(), b.Value())
}

func testDifference[T any](a, b T) (bool, unsafe.Pointer, unsafe.Pointer) {
	ap := reflect.ValueOf(a).UnsafePointer()
	bp := reflect.ValueOf(b).UnsafePointer()
	return ap != bp, ap, bp
}
