package option

import (
	"database/sql"
	"encoding/json"
	"slices"
	"testing"
	"time"
	_ "time/tzdata"

	"gotest.tools/v3/assert"
)

var (
	america, _ = time.LoadLocation("America/Anguilla")
	japan, _   = time.LoadLocation("Asia/Tokyo")
)

func TestOption_new_functions(t *testing.T) {
	num := 15
	{
		some := FromPointer(&num)
		assert.Assert(t, some.IsSome())
		assert.Equal(t, 15, some.Value())

		none := FromPointer((*int)(nil))
		assert.Assert(t, none.IsNone())
		assert.Equal(t, 0, none.Value())
	}
	{
		some := WrapPointer(&num)
		assert.Assert(t, some.IsSome())
		assert.Equal(t, &num, some.Value())

		none := WrapPointer((*int)(nil))
		assert.Assert(t, none.IsNone())
		assert.Equal(t, (*int)(nil), none.Value())
	}
	{
		some := FromSqlNull(sql.Null[int]{Valid: true, V: 15})
		assert.Assert(t, some.IsSome())
		assert.Equal(t, 15, some.Value())

		none := FromSqlNull(sql.Null[int]{Valid: false, V: 15})
		assert.Assert(t, none.IsNone())
		assert.Equal(t, 0, none.Value())
	}
}

func TestOption_Equal(t *testing.T) {
	t.Run("comparable", func(t *testing.T) {
		n := None[int]()
		s1 := Some(10)
		s2 := Some(20)

		assert.Assert(t, Equal(n, n))
		assert.Assert(t, !Equal(n, s1))
		assert.Assert(t, Equal(s1, s1))
		assert.Assert(t, !Equal(s1, s2))
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

		assert.Assert(t, Equal(n, n))
		assert.Assert(t, !Equal(s1, n))
		assert.Assert(t, Equal(s1, s2))

		assert.Assert(t, !Equal(s3, s4)) // not uses Equal implementation
		assert.Assert(t, EqualEqualer(s3, s4))
	})

	t.Run("EqualFunc", func(t *testing.T) {
		assert.Assert(
			t,
			Some([]bool{true, true}).
				EqualFunc(
					Some([]bool{true, true}),
					slices.Equal,
				),
		)
		assert.Assert(
			t,
			!Some([]bool{true, true}).
				EqualFunc(
					Some([]bool{true, true}),
					func(i, j []bool) bool { return false },
				),
		)
		assert.Assert(
			t,
			!Some([]bool{true, true}).
				EqualFunc(
					Some([]bool{true, false}),
					slices.Equal,
				),
		)
	})
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

		assert.Equal(t, Flatten(ss), Some(float64(1.22)))
		assert.Equal(t, Flatten(sn), None[float64]())
		assert.Equal(t, Flatten(nn), None[float64]())
	})

	t.Run("Get", func(t *testing.T) {
		ss := Some(float64(1.22))
		sn := None[float64]()

		assertGet := func(t *testing.T, o Option[float64], v float64, ok bool) {
			t.Helper()
			gotV, gotOk := o.Get()
			assert.Equal(t, gotV, v)
			assert.Equal(t, gotOk, ok)
		}

		assertGet(t, ss, 1.22, true)
		assertGet(t, sn, 0, false)
	})

	t.Run("MapOption", func(t *testing.T) {
		assert.Equal(t, Map(s, func(o string) int { return len(o) }), Some(3))
		assert.Equal(t, Map(n, func(o string) bool { return true }), None[bool]())
	})

	t.Run("Map", func(t *testing.T) {
		assert.Equal(t, s.Map(func(v string) string { return v + v }), Some("aaaaaa"))
		assert.Equal(t, n.Map(func(v string) string { return "ccc" }), None[string]())
	})

	t.Run("MapOrOption", func(t *testing.T) {
		assert.Equal(t, MapOr(s, 123, func(t string) int { return len(t) }), 3)
		assert.Equal(t, MapOr(n, 123, func(t string) int { return len(t) + 4 }), 123)
	})

	t.Run("MapOr", func(t *testing.T) {
		assert.Equal(t, s.MapOr("nah", func(t string) string { return "bbb" }), "bbb")
		assert.Equal(t, n.MapOr("nah", func(s string) string { return "bbbb" }), "nah")
	})

	t.Run("MapOrOpt", func(t *testing.T) {
		assert.Equal(t, s.MapOrOpt("nah", func(t string) string { return "bbb" }), Some("bbb"))
		assert.Equal(t, n.MapOrOpt("nah", func(s string) string { return "bbbb" }), Some("nah"))
	})

	t.Run("MapOrElseOption", func(t *testing.T) {
		assert.Equal(t, MapOrElse(s, func() int { return 123 }, func(t string) int { return len(t) }), 3)
		assert.Equal(t, MapOrElse(n, func() int { return 123 }, func(t string) int { return len(t) + 4 }), 123)
	})

	t.Run("MapOrElse", func(t *testing.T) {
		assert.Equal(t, s.MapOrElse(func() string { return "nah" }, func(t string) string { return "bbb" }), "bbb")
		assert.Equal(t, n.MapOrElse(func() string { return "nah" }, func(s string) string { return "bbbb" }), "nah")
	})

	t.Run("MapOrElseOpt", func(t *testing.T) {
		assert.Equal(t, s.MapOrElseOpt(func() string { return "nah" }, func(t string) string { return "bbb" }), Some("bbb"))
		assert.Equal(t, n.MapOrElseOpt(func() string { return "nah" }, func(s string) string { return "bbbb" }), Some("nah"))
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
