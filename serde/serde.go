package serde

import (
	"io"
	"reflect"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

var config = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
}.Froze()

func init() {
	config.RegisterExtension(&UndefinedableExtension{})
}

type IsUndefineder interface {
	IsUndefined() bool
}

var undefinedableTy = reflect2.TypeOfPtr((*IsUndefineder)(nil)).Elem()

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

// UndefinedableExtension is the extension for jsoniter.API.
// This forces jsoniter.API to skip undefined Undefinedable[T] struct fields when marshalling.
type UndefinedableExtension struct {
}

func (extension *UndefinedableExtension) UpdateStructDescriptor(structDescriptor *jsoniter.StructDescriptor) {
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

func (extension *UndefinedableExtension) CreateMapKeyDecoder(typ reflect2.Type) jsoniter.ValDecoder {
	return nil
}

func (extension *UndefinedableExtension) CreateMapKeyEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	return nil
}

func (extension *UndefinedableExtension) CreateDecoder(typ reflect2.Type) jsoniter.ValDecoder {
	return nil
}

func (extension *UndefinedableExtension) CreateEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	return nil
}

func (extension *UndefinedableExtension) DecorateDecoder(typ reflect2.Type, decoder jsoniter.ValDecoder) jsoniter.ValDecoder {
	return decoder
}

func (extension *UndefinedableExtension) DecorateEncoder(typ reflect2.Type, encoder jsoniter.ValEncoder) jsoniter.ValEncoder {
	return encoder
}

// MarshalJSON encodes v into JSON.
// It skips fields if those are undefined Undefinedable[T].
//
// v can be any type.
func MarshalJSON(v any) ([]byte, error) {
	return config.Marshal(v)
}

func NewEncoder(w io.Writer) *jsoniter.Encoder {
	return config.NewEncoder(w)
}

// UnmarshalJSON decodes data into v.
// v must be pointer type, return error otherwise.
//
// Currently this is almost same as json.Unmarshal.
// Future releases may change behavior of this function.
// It is safe to unmarshal data through this if v has at least an Undefinedable[T] field.
func UnmarshalJSON(data []byte, v any) error {
	return config.Unmarshal(data, v)
}

func NewDecoder(r io.Reader) *jsoniter.Decoder {
	return config.NewDecoder(r)
}
