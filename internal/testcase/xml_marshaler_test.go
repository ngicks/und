package testcase_test

import (
	"encoding/xml"
	"testing"

	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
	"gotest.tools/v3/assert"
)

var (
	_ xml.Marshaler = option.Option[any]{}
	_ xml.Marshaler = und.Und[any]{}
	_ xml.Marshaler = sliceund.Und[any]{}
	_ xml.Marshaler = elastic.Elastic[any]{}
	_ xml.Marshaler = sliceelastic.Elastic[any]{}
)

var (
	_ xml.Unmarshaler = (*option.Option[any])(nil)
	_ xml.Unmarshaler = (*und.Und[any])(nil)
	_ xml.Unmarshaler = (*sliceund.Und[any])(nil)
	_ xml.Unmarshaler = (*elastic.Elastic[any])(nil)
	_ xml.Unmarshaler = (*sliceelastic.Elastic[any])(nil)
)

type valueSet[T any] struct {
	Opt      option.Option[T]
	Und      und.Und[T]
	SliceUnd sliceund.Und[T]
	Ela      elastic.Elastic[T]
	SliceEla sliceelastic.Elastic[T]
}

func (v valueSet[T]) EqualFunc(t *testing.T, v2 valueSet[T], cmp func(i, j T) bool) {
	assert.Assert(t, v2.Opt.EqualFunc(v.Opt, cmp), "left = %+v, right %+v", v2.SliceUnd, v.SliceUnd)
	assert.Assert(t, v2.Und.EqualFunc(v.Und, cmp), "left = %+v, right %+v", v2.SliceUnd, v.SliceUnd)
	assert.Assert(t, v2.SliceUnd.EqualFunc(v.SliceUnd, cmp), "left = %+v, right %+v", v2.SliceUnd, v.SliceUnd)
	assert.Assert(t, v2.Ela.EqualFunc(v.Ela, cmp), "left = %+v, right %+v", v2.Ela, v.Ela)
	assert.Assert(t, v2.SliceEla.EqualFunc(v.SliceEla, cmp), "left = %+v, right %+v", v2.SliceEla, v.SliceEla)
}

type xmlMarshaler[T any] struct {
	XMLName  xml.Name                `xml:"test"`
	Pad1     int                     `xml:"pad1,omitempty"`
	Opt      option.Option[T]        `xml:"opt"`
	Pad2     int                     `xml:"pad2,omitempty"`
	Und      und.Und[T]              `xml:"und"`
	Pad3     int                     `xml:"pad3,omitempty"`
	SliceUnd sliceund.Und[T]         `xml:"sliceund"`
	Pad4     int                     `xml:"pad4,omitempty"`
	Ela      elastic.Elastic[T]      `xml:"ela"`
	Pad5     int                     `xml:"pad5,omitempty"`
	SliceEla sliceelastic.Elastic[T] `xml:"sliceela"`
	Pad6     int                     `xml:"pad6,omitempty"`
}

func (x xmlMarshaler[T]) into() valueSet[T] {
	return valueSet[T]{x.Opt, x.Und, x.SliceUnd, x.Ela, x.SliceEla}
}

type xmlSerdeTestCase[T any] struct {
	bin    string
	values valueSet[T]
}

func TestXmlMarshaler(t *testing.T) {
	for _, tc := range []xmlSerdeTestCase[int]{
		{
			`<test><pad1>1</pad1><opt>1</opt><pad2>2</pad2><und>1</und><pad3>3</pad3>` +
				`<sliceund>1</sliceund><pad4>4</pad4><ela>1</ela><ela>2</ela><pad5>5</pad5>` +
				`<sliceela>3</sliceela><sliceela>4</sliceela><pad6>6</pad6></test>`,
			valueSet[int]{
				option.Some(1),
				und.Defined(1),
				sliceund.Defined(1),
				elastic.FromValues(1, 2),
				sliceelastic.FromValues(3, 4),
			},
		},
		{
			`<test></test>`,
			valueSet[int]{},
		},
		{
			`<test><ela>2</ela><pad5>5</pad5><sliceela>3</sliceela></test>`,
			valueSet[int]{
				Ela:      elastic.FromValue(2),
				SliceEla: sliceelastic.FromValue(3),
			},
		},
	} {
		var s xmlMarshaler[int]
		err := xml.Unmarshal([]byte(tc.bin), &s)
		assert.NilError(t, err)
		tc.values.EqualFunc(t, s.into(), func(i, j int) bool { return i == j })
		bin, err := xml.Marshal(s)
		assert.NilError(t, err)
		assert.Equal(t, string(bin), tc.bin)
	}
}

type nested struct {
	Opt      option.Option[int]        `xml:"opt"`
	Und      und.Und[int]              `xml:"und"`
	SliceUnd sliceund.Und[int]         `xml:"sliceund"`
	Ela      elastic.Elastic[int]      `xml:"ela"`
	SliceEla sliceelastic.Elastic[int] `xml:"sliceela"`
}

func (n nested) Equal(v nested) bool {
	return option.Equal(n.Opt, v.Opt) &&
		und.Equal(n.Und, v.Und) &&
		sliceund.Equal(n.SliceUnd, v.SliceUnd) &&
		elastic.Equal(n.Ela, v.Ela) &&
		sliceelastic.Equal(n.SliceEla, v.SliceEla)
}

func TestXmlMarshaler_nested(t *testing.T) {
	for _, tc := range []xmlSerdeTestCase[nested]{
		{
			`<test></test>`,
			valueSet[nested]{},
		},
		{
			`<test><opt></opt><und></und>` +
				`<sliceund></sliceund><ela></ela><ela></ela>` +
				`<sliceela></sliceela><sliceela></sliceela></test>`,
			valueSet[nested]{
				option.Some(nested{}),
				und.Defined(nested{}),
				sliceund.Defined(nested{}),
				elastic.FromValues([]nested{{}, {}}...),
				sliceelastic.FromValues([]nested{{}, {}}...),
			},
		},
		{
			`<test><opt><sliceela>5</sliceela><sliceela>5</sliceela><sliceela>5</sliceela></opt>` +
				`<und><opt>444</opt></und>` +
				`<sliceund><und>789</und></sliceund>` +
				`<ela><sliceund>57</sliceund></ela><ela><ela>5</ela><ela>7</ela></ela>` +
				`<sliceela><ela>5</ela><ela>7</ela></sliceela><sliceela><sliceela>65</sliceela></sliceela></test>`,
			valueSet[nested]{
				option.Some(nested{SliceEla: sliceelastic.FromValues(5, 5, 5)}),
				und.Defined(nested{Opt: option.Some(444)}),
				sliceund.Defined(nested{Und: und.Defined(789)}),
				elastic.FromValues([]nested{{SliceUnd: sliceund.Defined(57)}, {Ela: elastic.FromValues(5, 7)}}...),
				sliceelastic.FromValues([]nested{{Ela: elastic.FromValues(5, 7)}, {SliceEla: sliceelastic.FromValue(65)}}...),
			},
		},
	} {
		var s xmlMarshaler[nested]
		err := xml.Unmarshal([]byte(tc.bin), &s)
		assert.NilError(t, err)
		tc.values.EqualFunc(t, s.into(), func(i, j nested) bool { return i.Equal(j) })
		bin, err := xml.Marshal(s)
		assert.NilError(t, err)
		assert.Equal(t, string(bin), tc.bin)
	}

}
