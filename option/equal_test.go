package option_test

import (
	"net/netip"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ngicks/und/v2/option"
	"github.com/stretchr/testify/assert"
)

type implementsEquality int

func (t implementsEquality) Equal(u implementsEquality) bool {
	return t%30 == u%30
}

type onlyPointerImplementsEquality int

func (t *onlyPointerImplementsEquality) Equal(u onlyPointerImplementsEquality) bool {
	return *t%30 == u%30
}

type nonComparableButImplementsEquality [][]int

func (t nonComparableButImplementsEquality) Sum() int {
	var sum int
	for _, v := range t {
		for _, vv := range v {
			sum += vv
		}
	}
	return sum
}

func (t nonComparableButImplementsEquality) Equal(u nonComparableButImplementsEquality) bool {
	return t.Sum() == u.Sum()
}

func TestEquality(t *testing.T) {
	// simple
	assert.True(t, option.None[int]().Equal(option.None[int]()))
	assert.False(t, option.None[int]().Equal(option.Some[int](0)))
	assert.False(t, option.Some[int](0).Equal(option.None[int]()))

	// comparable
	testEquality[int](t, 1, 1, true)
	testEquality[int](t, 1, 2, false)
	testEquality[netip.Addr](
		t,
		netip.AddrFrom4([4]byte{255, 255, 255, 255}),
		netip.AddrFrom4([4]byte{255, 255, 255, 255}),
		true,
	)
	testEquality[netip.Addr](
		t,
		netip.AddrFrom4([4]byte{255, 255, 255, 255}),
		netip.AddrFrom4([4]byte{255, 255, 124, 255}),
		false,
	)

	// custom equality
	testEquality[implementsEquality](t, 3, 3, true)
	testEquality[implementsEquality](t, 3, 2, false)
	testEquality[implementsEquality](t, 33, 3, true)

	// only pointer type implements custom equality
	testEquality[onlyPointerImplementsEquality](t, 3, 3, true)
	testEquality[onlyPointerImplementsEquality](t, 3, 2, false)
	testEquality[onlyPointerImplementsEquality](t, 33, 3, true)

	// non-comparable but implements equality
	testEquality[nonComparableButImplementsEquality](t, [][]int{{2, 4}, {1, 1, 2}}, [][]int{{10}}, true)
	testEquality[nonComparableButImplementsEquality](t, [][]int{{5}}, [][]int{{10}}, false)

	// non-comparable
	emptyFn := func() {}
	emptyFn2 := func() {}
	testEquality[func()](t, emptyFn, emptyFn, false)
	testEquality[func()](t, emptyFn, emptyFn2, false)

	// slice of comparable
	testEquality[[]string](t, nil, nil, true) // nil slice is considered as zero-length slice. Thus both are equal.
	testEquality[[]string](t, []string{}, []string{}, true)
	testEquality[[]string](t, []string{"foo", "bar", "baz"}, []string{"foo", "bar", "baz"}, true)
	testEquality[[]string](t, []string{"foo", "bar", "baz"}, []string{"foo", "bar"}, false)
	s1 := []string{"foofoo", "barbar"}
	testEquality[[]string](t, s1, s1, true)
	testEquality[[]string](t, s1, append(s1, "nah"), false)

	testEquality[[]int](t, nil, nil, true)
	testEquality[[]int](t, []int{}, []int{}, true)
	testEquality[[]int](t, []int{1, 2, 3}, []int{1, 2, 3}, true)
	testEquality[[]int](t, []int{2, 3, 1}, []int{1, 2, 3}, false)

	// slice of non comparable.
	testEquality[[][]string](t, nil, nil, false)
	testEquality[[][]string](t, [][]string{{"foo"}}, [][]string{{"foo"}}, false)
	testEquality[[][]string](t, [][]string{{"foo"}}, [][]string{{"bar"}}, false)

	// map with comparable value type
	testEquality[map[string]string](t, nil, nil, true)
	testEquality[map[string]string](t, map[string]string{}, map[string]string{}, true)
	testEquality[map[string]string](t, map[string]string{"foo": "bar"}, map[string]string{"foo": "bar"}, true)
	testEquality[map[string]string](t, map[string]string{"foo": "bar"}, map[string]string{"foo": "nah"}, false)
	testEquality[map[string]string](t, map[string]string{"foo": "bar"}, map[string]string{"baz": "qux"}, false)
	testEquality[map[string]string](t, map[string]string{"foo": "bar"}, map[string]string{"foo": "bar", "baz": "qux"}, false)
	m1 := map[string]string{"foo": "barbar"}
	testEquality[map[string]string](t, m1, m1, true)
	// map with non comparable value type
	// always false.
	testEquality[map[string][]string](t, nil, nil, false)
	testEquality[map[string][]string](t, map[string][]string{}, map[string][]string{}, false)
	testEquality[map[string][]string](
		t,
		map[string][]string{"foo": {"bar"}},
		map[string][]string{"foo": {"bar"}},
		false,
	)
	testEquality[map[string][]string](
		t,
		map[string][]string{"foo": {"bar"}},
		map[string][]string{"baz": {"qux"}},
		false,
	)
	testEquality[map[string][]string](
		t,
		map[string][]string{"foo": {"bar"}},
		map[string][]string{"foo": {"bar"}, "baz": {"qux"}},
		false,
	)
	m2 := map[string][]string{"foo": {"barbar"}}
	testEquality[map[string][]string](t, m2, m2, false)
}

func testEquality[T any](t *testing.T, l, r T, equal bool) {
	t.Helper()
	if option.Some(l).Equal(option.Some(r)) != equal {
		t.Errorf(
			"must be = %t but not.\nleft  = %+v\nright = %+v\ndiff = %s",
			equal, l, r, cmp.Diff(l, r),
		)
	}
}

type someComparableStruct struct {
	Foo string
	Bar int
	Baz bool
}

func TestHashable(t *testing.T) {
	assert := assert.New(t)

	var (
		some1 = option.Some[someComparableStruct](someComparableStruct{Foo: "foo"})
		some2 = option.Some[someComparableStruct](someComparableStruct{Foo: "foo"})
		some3 = option.Some[someComparableStruct](someComparableStruct{Foo: "foo", Bar: 5})
		none  = option.None[someComparableStruct]()
	)

	assert.True(some1 == some2)
	assert.False(some1 == some3)
	assert.False(some1 == none)
	assert.False(some3 == none)

	// invalid operation: option.None[[]string]() == option.None[[]string]() (struct containing []string cannot be compared)
	// assert.False(option.None[[]string]() == option.None[[]string]())
}
