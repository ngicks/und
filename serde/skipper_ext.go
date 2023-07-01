package serde

import (
	"reflect"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

type IsUndefineder interface {
	IsUndefined() bool
}

var undefinedableTy = reflect2.TypeOfPtr((*IsUndefineder)(nil)).Elem()

// UndefinedSkipperExtension is the extension for jsoniter.API.
// When marshaling, this extension forces jsoniter.API to skip undefined struct fields.
// A field is considered undefined if its type implements interface{ IsUndefined() bool }
// and if it returns true.
type UndefinedSkipperExtension struct {
}

func (extension *UndefinedSkipperExtension) UpdateStructDescriptor(structDescriptor *jsoniter.StructDescriptor) {
	if structDescriptor.Type.Implements(undefinedableTy) {
		return
	}

	for _, binding := range structDescriptor.Fields {
		if binding.Field.Type().Implements(undefinedableTy) {
			enc := binding.Encoder
			binding.Field = NewFakedOmitemptyField(binding.Field)
			binding.Encoder = UndefinedableEncoder{ty: binding.Field.Type(), org: enc}
		}
	}
}

func (extension *UndefinedSkipperExtension) CreateMapKeyDecoder(typ reflect2.Type) jsoniter.ValDecoder {
	return nil
}

func (extension *UndefinedSkipperExtension) CreateMapKeyEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	return nil
}

func (extension *UndefinedSkipperExtension) CreateDecoder(typ reflect2.Type) jsoniter.ValDecoder {
	return nil
}

func (extension *UndefinedSkipperExtension) CreateEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	return nil
}

func (extension *UndefinedSkipperExtension) DecorateDecoder(typ reflect2.Type, decoder jsoniter.ValDecoder) jsoniter.ValDecoder {
	return decoder
}

func (extension *UndefinedSkipperExtension) DecorateEncoder(typ reflect2.Type, encoder jsoniter.ValEncoder) jsoniter.ValEncoder {
	return encoder
}

// UndefinedableEncoder fakes the Encoder so that
// undefined Undefinedable[T] fields are skipped.
type UndefinedableEncoder struct {
	ty  reflect2.Type
	org jsoniter.ValEncoder
}

func (e UndefinedableEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	val := e.ty.UnsafeIndirect(ptr)
	return val.(IsUndefineder).IsUndefined()
}

func (e UndefinedableEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	e.org.Encode(ptr, stream)
}

// FakedOmitemptyField implements reflect2.StructField interface,
// faking the struct tag to pretend it is always tagged with ,omitempty option.
//
// The Zero value is not ready for use. Make it with NewFakedOmitemptyField.
type FakedOmitemptyField struct {
	reflect2.StructField
	fakedTag reflect.StructTag
}

func NewFakedOmitemptyField(f reflect2.StructField) FakedOmitemptyField {
	return FakedOmitemptyField{
		StructField: f,
		fakedTag:    FakeOmitempty(f.Tag()),
	}
}

func (f FakedOmitemptyField) Tag() reflect.StructTag {
	return f.fakedTag
}
