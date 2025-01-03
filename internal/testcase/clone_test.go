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

func TestCloner(t *testing.T) {
	clonerTetSet(t)
}

func clonerTetSet(t *testing.T) {
	testCloner(
		t,
		option.Some([]int{1, 2, 3}),
		slices.Clone,
		testDifferenceValuer,
	)
	testCloner(
		t,
		option.Options[[]int]{option.Some([]int{1, 2, 3}), option.Some([]int{4, 5, 6})},
		slices.Clone,
		testDifferenceOptions,
	)
	testCloner(
		t,
		und.Defined([]int{1, 2, 3}),
		slices.Clone,
		testDifferenceValuer,
	)
	testCloner(
		t,
		sliceund.Defined([]int{1, 2, 3}),
		slices.Clone,
		testDifferenceValuer,
	)
	testCloner(
		t,
		elastic.FromOptions(option.Some([]int{1, 2, 3}), option.Some([]int{4, 5, 6})),
		slices.Clone,
		func(a, b elastic.Elastic[[]int]) (bool, unsafe.Pointer, unsafe.Pointer) {
			return testDifferenceOptions(a.Unwrap().Value(), b.Unwrap().Value())
		},
	)
	testCloner(
		t,
		sliceelastic.FromOptions(option.Some([]int{1, 2, 3}), option.Some([]int{4, 5, 6})),
		slices.Clone,
		func(a, b sliceelastic.Elastic[[]int]) (bool, unsafe.Pointer, unsafe.Pointer) {
			return testDifferenceOptions(a.Unwrap().Value(), b.Unwrap().Value())
		},
	)
}

type cloner[T, U any] interface {
	CloneFunc(cloneU func(U) U) T
}

func testCloner[C cloner[T, U], T, U any](t *testing.T, initial C, cloneU func(U) U, isDifferent func(a C, b T) (bool, unsafe.Pointer, unsafe.Pointer)) {
	t.Helper()
	cloned := initial.CloneFunc(cloneU)
	diff, ap, bp := isDifferent(initial, cloned)
	if !diff {
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
