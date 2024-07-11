package validate_test

import (
	"testing"

	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund/elastic"
	"github.com/ngicks/und/validate"
	"gotest.tools/v3/assert"
)

type All struct {
	OptRequired       option.Option[string] `und:"required"`
	OptNullish        option.Option[string] `und:"nullish"`
	OptDef            option.Option[string] `und:"def"`
	OptNull           option.Option[string] `und:"null"`
	OptUnd            option.Option[string] `und:"und"`
	OptDefOrUnd       option.Option[string] `und:"def,und"`
	OptDefOrNull      option.Option[string] `und:"def,null"`
	OptNullOrUnd      option.Option[string] `und:"null,und"`
	OptDefOrNullOrUnd option.Option[string] `und:"def,null,und"`

	UndRequired       und.Und[string] `und:"required"`
	UndNullish        und.Und[string] `und:"nullish"`
	UndDef            und.Und[string] `und:"def"`
	UndNull           und.Und[string] `und:"null"`
	UndUnd            und.Und[string] `und:"und"`
	UndDefOrUnd       und.Und[string] `und:"def,und"`
	UndDefOrNull      und.Und[string] `und:"def,null"`
	UndNullOrUnd      und.Und[string] `und:"null,und"`
	UndDefOrNullOrUnd und.Und[string] `und:"def,null,und"`

	ElaRequired       elastic.Elastic[string] `und:"required"`
	ElaNullish        elastic.Elastic[string] `und:"nullish"`
	ElaDef            elastic.Elastic[string] `und:"def"`
	ElaNull           elastic.Elastic[string] `und:"null"`
	ElaUnd            elastic.Elastic[string] `und:"und"`
	ElaDefOrUnd       elastic.Elastic[string] `und:"def,und"`
	ElaDefOrNull      elastic.Elastic[string] `und:"def,null"`
	ElaNullOrUnd      elastic.Elastic[string] `und:"null,und"`
	ElaDefOrNullOrUnd elastic.Elastic[string] `und:"def,null,und"`

	ElaEqEq elastic.Elastic[string] `und:"len==1"`
	ElaGr   elastic.Elastic[string] `und:"len>1"`
	ElaGrEq elastic.Elastic[string] `und:"len>=1"`
	ElaLe   elastic.Elastic[string] `und:"len<1"`
	ElaLeEq elastic.Elastic[string] `und:"len<=1"`

	ElaEqEquRequired elastic.Elastic[string] `und:"required,len==2"`
	ElaEqEquNullish  elastic.Elastic[string] `und:"nullish,len==2"`
	ElaEqEquDef      elastic.Elastic[string] `und:"def,len==2"`
	ElaEqEquNull     elastic.Elastic[string] `und:"null,len==2"`
	ElaEqEquUnd      elastic.Elastic[string] `und:"und,len==2"`

	ElaEqEqNonNull elastic.Elastic[string] `und:"values:nonnull,len==3"`
}

type (
	Nested struct {
		A option.Option[ChildA] `und:"required"`
	}

	ChildA struct {
		B option.Option[ChildB] `und:"required"`
	}

	ChildB struct {
		C option.Option[string] `und:"required"`
	}
)

type (
	Embedded struct {
		Foo string
		Sub
		Bar string
	}

	Sub struct {
		C option.Option[string] `und:"required"`
	}
)

// invalid, multiple mutually exclusive options
type (
	invalidMultiple1 struct {
		A option.Option[string] `und:"required,required"`
	}
	invalidMultiple2 struct {
		A option.Option[string] `und:"nullish,nullish"`
	}
	invalidMultiple3 struct {
		A option.Option[string] `und:"def,def"`
	}
	invalidMultiple4 struct {
		A option.Option[string] `und:"null,null"`
	}
	invalidMultiple5 struct {
		A option.Option[string] `und:"und,und"`
	}
	invalidMultiple6 struct {
		A option.Option[string] `und:"def,und,def"`
	}
	invalidMultiple7 struct {
		A option.Option[string] `und:"def,null,null"`
	}
	invalidMultiple8 struct {
		A option.Option[string] `und:"null,und,null"`
	}
	invalidMultiple9 struct {
		A option.Option[string] `und:"def,null,und,null"`
	}
	invalidMultiple10 struct {
		A option.Option[string] `und:"required,null"`
	}
	invalidMultiple11 struct {
		A option.Option[string] `und:"len==1,len==2"`
	}
	invalidMultiple12 struct {
		A option.Option[string] `und:"values:nonnull,values:nonnull"`
	}
)

type (
	invalidMalformedLen1 struct {
		A elastic.Elastic[string] `und:"len123"`
	}
	invalidMalformedLen2 struct {
		A elastic.Elastic[string] `und:"len==-123"`
	}
)

type (
	invalidMalformedValues1 struct {
		A elastic.Elastic[string] `und:"values:non-null"`
	}
)

type (
	invalidWrongOptionLenOnOpt struct {
		A option.Option[string] `und:"len==1"`
	}
	invalidWrongOptionValuesOnOpt struct {
		A option.Option[string] `und:"values:nonnull"`
	}
)

type (
	invalidNested struct {
		B option.Option[invalidMalformedLen1] `und:"required"`
	}
)

type (
	validRecursive struct {
		Intermediate
	}

	Intermediate struct {
		Bar option.Option[int] `und:"required"`
		Baz *validRecursive
	}
)

type (
	validTree struct {
		Node *ValidTreeNode
	}

	ValidTreeNode struct {
		V    ValidTreeValues
		L, R *ValidTreeNode
	}

	ValidTreeValues struct {
		A int
		B option.Option[string] `und:"required"`
		C bool
	}
)

var (
	valid = All{
		OptRequired:       option.Some("foo"),
		OptNullish:        option.None[string](),
		OptDef:            option.Some("bar"),
		OptNull:           option.None[string](),
		OptUnd:            option.None[string](),
		OptDefOrUnd:       option.Some("baz"),
		OptDefOrNull:      option.Some("qux"),
		OptNullOrUnd:      option.None[string](),
		OptDefOrNullOrUnd: option.Some("quux"),

		UndRequired:       und.Defined("corge"),
		UndNullish:        und.Null[string](),
		UndDef:            und.Defined("grault"),
		UndNull:           und.Null[string](),
		UndUnd:            und.Undefined[string](),
		UndDefOrUnd:       und.Defined("garply"),
		UndDefOrNull:      und.Defined("waldo"),
		UndNullOrUnd:      und.Null[string](),
		UndDefOrNullOrUnd: und.Defined("fred"),

		ElaRequired:       elastic.FromValue("plugh"),
		ElaNullish:        elastic.Null[string](),
		ElaDef:            elastic.FromValue("xyzzy"),
		ElaNull:           elastic.Null[string](),
		ElaUnd:            elastic.Undefined[string](),
		ElaDefOrUnd:       elastic.FromValue("thud"),
		ElaDefOrNull:      elastic.FromValue("foofoo"),
		ElaNullOrUnd:      elastic.Null[string](),
		ElaDefOrNullOrUnd: elastic.FromValue("barbar"),

		ElaEqEq: elastic.FromValue("bazbaz"),
		ElaGr:   elastic.FromValues([]string{"quxqux", "quuxquux"}),
		ElaGrEq: elastic.FromValue("corgecorge"),
		ElaLe:   elastic.FromValues([]string{}),
		ElaLeEq: elastic.FromValue("graultgrault"),

		ElaEqEquRequired: elastic.FromValues([]string{"foofoo", "barbar"}),
		ElaEqEquNullish:  elastic.FromValues([]string{"foofoo", "barbar"}),
		ElaEqEquDef:      elastic.FromValues([]string{"foofoo", "barbar"}),
		ElaEqEquNull:     elastic.FromValues([]string{"foofoo", "barbar"}),
		ElaEqEquUnd:      elastic.FromValues([]string{"foofoo", "barbar"}),

		ElaEqEqNonNull: elastic.FromValues([]string{"a", "b", "c"}),
	}
)

func TestValidate_all(t *testing.T) {
	assert.NilError(t, validate.CheckUnd(valid))
	assert.NilError(t, validate.ValidateUnd(valid))
}

func TestValidate_all_invalid(t *testing.T) {
	fo := option.Some("foo")
	fu := und.Defined("foo")
	fe := elastic.FromValue("foo")
	for _, patcher := range []func(v All) All{
		func(v All) All { v.OptRequired = option.None[string](); return v },
		func(v All) All { v.OptNullish = fo; return v },
		func(v All) All { v.OptDef = option.None[string](); return v },
		func(v All) All { v.OptNull = fo; return v },
		func(v All) All { v.OptUnd = fo; return v },
		func(v All) All { v.OptNullOrUnd = fo; return v },
		func(v All) All { v.UndRequired = und.Undefined[string](); return v },
		func(v All) All { v.UndNullish = fu; return v },
		func(v All) All { v.UndDef = und.Null[string](); return v },
		func(v All) All { v.UndNull = und.Undefined[string](); return v },
		func(v All) All { v.UndUnd = fu; return v },
		func(v All) All { v.UndDefOrUnd = und.Null[string](); return v },
		func(v All) All { v.UndDefOrNull = und.Undefined[string](); return v },
		func(v All) All { v.UndNullOrUnd = fu; return v },
		func(v All) All { v.ElaRequired = elastic.Null[string](); return v },
		func(v All) All { v.ElaNullish = fe; return v },
		func(v All) All { v.ElaDef = elastic.Undefined[string](); return v },
		func(v All) All { v.ElaNull = fe; return v },
		func(v All) All { v.ElaUnd = fe; return v },
		func(v All) All { v.ElaDefOrUnd = elastic.Null[string](); return v },
		func(v All) All { v.ElaDefOrNull = elastic.Undefined[string](); return v },
		func(v All) All { v.ElaNullOrUnd = fe; return v },
		func(v All) All { v.ElaEqEq = elastic.FromValues([]string{}); return v },
		func(v All) All { v.ElaGr = fe; return v },
		func(v All) All { v.ElaGrEq = elastic.FromValues([]string{}); return v },
		func(v All) All { v.ElaLe = fe; return v },
		func(v All) All {
			v.ElaLeEq = elastic.FromOptions([]option.Option[string]{option.None[string](), option.None[string]()})
			return v
		},
		func(v All) All {
			v.ElaEqEqNonNull = elastic.FromOptions([]option.Option[string]{option.Some("a"), option.None[string](), option.Some("c")})
			return v
		},
	} {
		patched := patcher(valid)
		assert.NilError(t, validate.CheckUnd(patched))
		err := validate.ValidateUnd(patched)
		t.Logf("%v", err)
		assert.Assert(t, err != nil)
	}
}

func TestValidate_nested(t *testing.T) {
	assert.NilError(t, validate.CheckUnd(Nested{}))
	assert.NilError(t, validate.ValidateUnd(Nested{
		A: option.Some(ChildA{
			B: option.Some(ChildB{
				C: option.Some("foo"),
			}),
		}),
	}))
	for _, n := range []Nested{
		{
			A: option.Some(ChildA{
				B: option.Some(ChildB{
					C: option.None[string](),
				}),
			}),
		},
		{
			A: option.Some(ChildA{
				B: option.None[ChildB](),
			}),
		},
		{
			A: option.None[ChildA](),
		},
	} {
		err := validate.ValidateUnd(n)
		t.Logf("err = %v", err)
		assert.Assert(t, err != nil)
	}
}

func TestValidate_embedded(t *testing.T) {
	assert.NilError(t, validate.CheckUnd(Embedded{}))
	assert.NilError(t, validate.ValidateUnd(Embedded{
		Foo: "foo",
		Sub: Sub{
			C: option.Some("sub"),
		},
		Bar: "bar",
	}))
	err := validate.ValidateUnd(Embedded{})
	t.Logf("err = %v", err)
	assert.Assert(t, err != nil)
}

func TestValidate_invalid_options(t *testing.T) {
	for _, tt := range []any{
		invalidMultiple1{},
		invalidMultiple2{},
		invalidMultiple3{},
		invalidMultiple4{},
		invalidMultiple5{},
		invalidMultiple6{},
		invalidMultiple7{},
		invalidMultiple8{},
		invalidMultiple9{},
		invalidMultiple10{},
		invalidMultiple11{},
		invalidMultiple12{},
	} {
		err := validate.CheckUnd(tt)
		t.Logf("err = %v", err)
		assert.ErrorIs(t, err, validate.ErrMultipleOption)
	}

	for _, tt := range []any{
		invalidMalformedLen1{},
		invalidMalformedLen2{},
	} {
		err := validate.CheckUnd(tt)
		t.Logf("err = %v", err)
		assert.ErrorIs(t, err, validate.ErrMalformedLen)
	}

	for _, tt := range []any{
		invalidMalformedValues1{},
	} {
		err := validate.CheckUnd(tt)
		t.Logf("err = %v", err)
		assert.ErrorIs(t, err, validate.ErrMalformedValues)
	}

	for _, tt := range []any{
		invalidWrongOptionLenOnOpt{},
		invalidWrongOptionValuesOnOpt{},
	} {
		err := validate.CheckUnd(tt)
		t.Logf("err = %v", err)
		assert.Assert(t, err != nil)
	}

	err := validate.CheckUnd(invalidNested{})
	t.Logf("err = %v", err)
	assert.ErrorIs(t, err, validate.ErrMalformedLen)
	assert.ErrorContains(t, err, "B.A:")
}

func TestValidate_recursion_embedded(t *testing.T) {
	assert.NilError(t, validate.CheckUnd(validRecursive{}))
	assert.NilError(t, validate.ValidateUnd(validRecursive{Intermediate{Bar: option.Some(5)}}))
	assert.Assert(t, validate.ValidateUnd(validRecursive{Intermediate{}}) != nil)
	assert.NilError(t, validate.ValidateUnd(validRecursive{Intermediate{Bar: option.Some(5), Baz: &validRecursive{Intermediate{Bar: option.Some[int](15)}}}}))
	assert.Assert(t, validate.ValidateUnd(validRecursive{Intermediate{Bar: option.Some(5), Baz: &validRecursive{Intermediate{Bar: option.None[int]()}}}}) != nil)
}

func TestValidate_recursion(t *testing.T) {
	assert.NilError(t, validate.CheckUnd(validTree{}))
	assert.NilError(t, validate.ValidateUnd(validTree{
		Node: &ValidTreeNode{
			V: ValidTreeValues{
				A: 5,
				B: option.Some("foo"),
				C: true,
			},
		},
	}))
	assert.Assert(t, validate.ValidateUnd(validTree{
		Node: &ValidTreeNode{
			V: ValidTreeValues{
				A: 5,
				B: option.None[string](),
				C: true,
			},
		},
	}) != nil)
	assert.NilError(t, validate.ValidateUnd(validTree{
		Node: &ValidTreeNode{
			V: ValidTreeValues{
				A: 5,
				B: option.Some("foo"),
				C: true,
			},
			L: &ValidTreeNode{
				V: ValidTreeValues{
					B: option.Some("bar"),
				},
			},
		},
	}))
	assert.Assert(t, validate.ValidateUnd(validTree{
		Node: &ValidTreeNode{
			V: ValidTreeValues{
				A: 5,
				B: option.Some("foo"),
				C: true,
			},
			L: &ValidTreeNode{
				V: ValidTreeValues{
					B: option.Some("bar"),
				},
			},
			R: &ValidTreeNode{
				V: ValidTreeValues{
					B: option.Some("baz"),
				},
				R: &ValidTreeNode{
					V: ValidTreeValues{
						B: option.None[string](),
					},
				},
			},
		},
	}) != nil)
}
