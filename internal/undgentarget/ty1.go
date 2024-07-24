package undgentarget

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
)

//undgen:ignore

type All struct {
	Foo string
	Bar *string
	Baz *struct{}
	Qux []string

	UntouchedOpt      option.Option[int] `json:",omitzero"`
	UntouchedUnd      und.Und[int]       `json:",omitzero"`
	UntouchedSliceUnd sliceund.Und[int]  `json:",omitzero"`

	OptRequired       option.Option[string] `json:",omitzero" und:"required"`
	OptNullish        option.Option[string] `json:",omitzero" und:"nullish"`
	OptDef            option.Option[string] `json:",omitzero" und:"def"`
	OptNull           option.Option[string] `json:",omitzero" und:"null"`
	OptUnd            option.Option[string] `json:",omitzero" und:"und"`
	OptDefOrUnd       option.Option[string] `json:",omitzero" und:"def,und"`
	OptDefOrNull      option.Option[string] `json:",omitzero" und:"def,null"`
	OptNullOrUnd      option.Option[string] `json:",omitzero" und:"null,und"`
	OptDefOrNullOrUnd option.Option[string] `json:",omitzero" und:"def,null,und"`

	UndRequired       und.Und[string] `json:",omitzero" und:"required"`
	UndNullish        und.Und[string] `json:",omitzero" und:"nullish"`
	UndDef            und.Und[string] `json:",omitzero" und:"def"`
	UndNull           und.Und[string] `json:",omitzero" und:"null"`
	UndUnd            und.Und[string] `json:",omitzero" und:"und"`
	UndDefOrUnd       und.Und[string] `json:",omitzero" und:"def,und"`
	UndDefOrNull      und.Und[string] `json:",omitzero" und:"def,null"`
	UndNullOrUnd      und.Und[string] `json:",omitzero" und:"null,und"`
	UndDefOrNullOrUnd und.Und[string] `json:",omitzero" und:"def,null,und"`

	ElaRequired       elastic.Elastic[string] `json:",omitzero" und:"required"`
	ElaNullish        elastic.Elastic[string] `json:",omitzero" und:"nullish"`
	ElaDef            elastic.Elastic[string] `json:",omitzero" und:"def"`
	ElaNull           elastic.Elastic[string] `json:",omitzero" und:"null"`
	ElaUnd            elastic.Elastic[string] `json:",omitzero" und:"und"`
	ElaDefOrUnd       elastic.Elastic[string] `json:",omitzero" und:"def,und"`
	ElaDefOrNull      elastic.Elastic[string] `json:",omitzero" und:"def,null"`
	ElaNullOrUnd      elastic.Elastic[string] `json:",omitzero" und:"null,und"`
	ElaDefOrNullOrUnd elastic.Elastic[string] `json:",omitzero" und:"def,null,und"`

	ElaEqEq elastic.Elastic[string] `json:",omitzero" und:"len==1"`
	ElaGr   elastic.Elastic[string] `json:",omitzero" und:"len>1"`
	ElaGrEq elastic.Elastic[string] `json:",omitzero" und:"len>=1"`
	ElaLe   elastic.Elastic[string] `json:",omitzero" und:"len<1"`
	ElaLeEq elastic.Elastic[string] `json:",omitzero" und:"len<=1"`

	ElaEqEquRequired elastic.Elastic[string] `json:",omitzero" und:"required,len==2"`
	ElaEqEquNullish  elastic.Elastic[string] `json:",omitzero" und:"nullish,len==2"`
	ElaEqEquDef      elastic.Elastic[string] `json:",omitzero" und:"def,len==2"`
	ElaEqEquNull     elastic.Elastic[string] `json:",omitzero" und:"null,len==2"`
	ElaEqEquUnd      elastic.Elastic[string] `json:",omitzero" und:"und,len==2"`

	ElaEqEqNonNullSlice  elastic.Elastic[string] `json:",omitzero" und:"values:nonnull"`
	ElaEqEqNonNullSingle elastic.Elastic[string] `json:",omitzero" und:"values:nonnull,len==1"`
	ElaEqEqNonNull       elastic.Elastic[string] `json:",omitzero" und:"values:nonnull,len==3"`
}

//undgen:ignore
type Ignored struct {
	Foo string
	Bar int
	Baz option.Option[int] `json:",omitzero" und:"required"`
}

type Ignored2 struct {
	Foo string
	Bar int
}

type WithTypeParam[T any] struct {
	Foo string
	Bar T
	Baz option.Option[T] `json:",omitzero" und:"required"`
}
