package option

import (
	"encoding/json"
	"slices"
	"testing"
	"time"
	_ "time/tzdata"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

var (
	america, _ = time.LoadLocation("America/Anguilla")
	japan, _   = time.LoadLocation("Asia/Tokyo")
)

func TestOption_Equal(t *testing.T) {
	t.Run("comparable", func(t *testing.T) {
		n := None[int]()
		s1 := Some(10)
		s2 := Some(20)

		assert.Assert(t, n.Equal(n))
		assert.Assert(t, !n.Equal(s1))
		assert.Assert(t, s1.Equal(s1))
		assert.Assert(t, !s1.Equal(s2))
	})

	t.Run("comparable_but_Equaler", func(t *testing.T) {
		n := None[time.Time]()
		cur := time.Now()
		s1 := Some(cur)
		s2 := Some(cur)
		s3 := Some(cur.In(japan))
		s4 := Some(cur.In(america))

		assert.Equal(t, n, n)
		assert.Equal(t, s1, s2)
		assert.Assert(t, s3 != s4)

		assert.Assert(t, n.Equal(n))
		assert.Assert(t, !s1.Equal(n))
		assert.Assert(t, s1.Equal(s2))
		assert.Assert(t, s3.Equal(s4))
	})

	t.Run("uncomparable", func(t *testing.T) {
		n := None[[]string]()
		s1 := Some([]string{"foo", "bar"})
		s2 := Some([]string{"foo", "bar"})

		assert.Assert(t, n.Equal(n))
		assert.Assert(t, !n.Equal(s1))
		assert.Assert(t, cmp.Panics(func() { s1.Equal(s1) }))
		assert.Assert(t, cmp.Panics(func() { s1.Equal(s2) }))
	})

	t.Run("Equal implementor", func(t *testing.T) {
		assert.Assert(t, Some(eq1{true, true}).Equal(Some(eq1{true, true})))
		assert.Assert(t, !Some(eq1{true, true}).Equal(Some(eq1{true, false})))
		assert.Assert(t, Some(eq2{"foo", "bar"}).Equal(Some(eq2{"foo", "bar"})))
		assert.Assert(t, !Some(eq2{"foo", "foo"}).Equal(Some(eq2{"foo", "bar"})))
	})
}

type eq1 []bool

func (e eq1) Equal(e2 eq1) bool {
	return slices.Equal(e, e2)
}

type eq2 []string

func (e *eq2) Equal(e2 eq2) bool {
	return slices.Equal(*e, e2)
}

func TestOption_methods(t *testing.T) {
	n := None[string]()
	s := Some("aaa")
	s2 := Some("bbb")

	t.Run("IsSome_IsNone_IsZero", func(t *testing.T) {
		assert.Assert(t, !n.IsSome())
		assert.Assert(t, n.IsNone())
		assert.Assert(t, n.IsZero())
		assert.Assert(t, s.IsSome())
		assert.Assert(t, !s.IsNone())
		assert.Assert(t, !s.IsZero())
	})

	t.Run("IsSomeAnd", func(t *testing.T) {
		assert.Equal(t, n.IsSomeAnd(func(s string) bool { return s == "" }), false)
		assert.Equal(t, s.IsSomeAnd(func(s string) bool { return s == "aaa" }), true)
		assert.Equal(t, s.IsSomeAnd(func(s string) bool { return s == "aaaii" }), false)
	})

	t.Run("Value", func(t *testing.T) {
		assert.Equal(t, n.Value(), "")
		assert.Equal(t, s.Value(), "aaa")
	})

	t.Run("Pointer", func(t *testing.T) {
		assert.Assert(t, n.Pointer() == nil)
		assert.Assert(t, s.Pointer() != nil)
		p := s.Pointer()
		assert.Equal(t, *p, "aaa")
		*p = "bbb"
		assert.Equal(t, s.Value(), "aaa")
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		bin, err := json.Marshal(n)
		assert.NilError(t, err)
		assert.Equal(t, string(bin), "null")
		bin, err = s.MarshalJSON()
		assert.NilError(t, err)
		assert.Equal(t, string(bin), "\"aaa\"")
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		var o Option[int]
		assert.NilError(t, json.Unmarshal([]byte(`123456`), &o))
		assert.Assert(t, o.IsSome())
		assert.Equal(t, o.Value(), 123456)
		var o2 Option[string]
		assert.Error(t, json.Unmarshal([]byte(`123456`), &o2), "json: cannot unmarshal number into Go value of type string")
		assert.Assert(t, o2.IsNone())
	})

	t.Run("And", func(t *testing.T) {
		assert.Equal(t, s.And(s2), s2)
		assert.Equal(t, n.And(s2), n)
	})

	t.Run("AndThen", func(t *testing.T) {
		assert.Equal(t, s.AndThen(func(x string) Option[string] { return Some("ccc") }), Some("ccc"))
		assert.Equal(t, n.AndThen(func(x string) Option[string] { return Some("ccc") }), None[string]())
	})

	t.Run("Filter", func(t *testing.T) {
		assert.Equal(t, s.Filter(func(t string) bool { return t == "aaa" }), s)
		assert.Equal(t, s.Filter(func(t string) bool { return t == "bbb" }), None[string]())
		assert.Equal(t, n.Filter(func(t string) bool { return t == "" }), None[string]())
	})

	t.Run("FlattenOption", func(t *testing.T) {
		ss := Some(Some(float64(1.22)))
		sn := Some(None[float64]())
		nn := None[Option[float64]]()

		assert.Equal(t, FlattenOption(ss), Some(float64(1.22)))
		assert.Equal(t, FlattenOption(sn), None[float64]())
		assert.Equal(t, FlattenOption(nn), None[float64]())
	})

	t.Run("MapOption", func(t *testing.T) {
		assert.Equal(t, MapOption(s, func(o string) int { return len(o) }), Some(3))
		assert.Equal(t, MapOption(n, func(o string) bool { return true }), None[bool]())
	})

	t.Run("Map", func(t *testing.T) {
		assert.Equal(t, s.Map(func(v string) string { return v + v }), Some("aaaaaa"))
		assert.Equal(t, n.Map(func(v string) string { return "ccc" }), None[string]())
	})

	t.Run("MapOrOption", func(t *testing.T) {
		assert.Equal(t, MapOrOption(s, 123, func(t string) int { return len(t) }), 3)
		assert.Equal(t, MapOrOption(n, 123, func(t string) int { return len(t) + 4 }), 123)
	})

	t.Run("MapOr", func(t *testing.T) {
		assert.Equal(t, s.MapOr("nah", func(t string) string { return "bbb" }), "bbb")
		assert.Equal(t, n.MapOr("nah", func(s string) string { return "bbbb" }), "nah")
	})

	t.Run("MapOrElseOption", func(t *testing.T) {
		assert.Equal(t, MapOrElseOption(s, func() int { return 123 }, func(t string) int { return len(t) }), 3)
		assert.Equal(t, MapOrElseOption(n, func() int { return 123 }, func(t string) int { return len(t) + 4 }), 123)
	})

	t.Run("MapOrElse", func(t *testing.T) {
		assert.Equal(t, s.MapOrElse(func() string { return "nah" }, func(t string) string { return "bbb" }), "bbb")
		assert.Equal(t, n.MapOrElse(func() string { return "nah" }, func(s string) string { return "bbbb" }), "nah")
	})

	t.Run("Or", func(t *testing.T) {
		assert.Equal(t, s.Or(n), s)
		assert.Equal(t, n.Or(s), s)
	})

	t.Run("OrElse", func(t *testing.T) {
		assert.Equal(t, s.OrElse(func() Option[string] { return n }), s)
		assert.Equal(t, n.OrElse(func() Option[string] { return s }), s)
	})

	t.Run("Xor", func(t *testing.T) {
		assert.Equal(t, s.Xor(s), None[string]())
		assert.Equal(t, n.Xor(n), None[string]())
		assert.Equal(t, s.Xor(n), s)
		assert.Equal(t, n.Xor(s), s)
	})
}
