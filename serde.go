package undefinedablejson

import (
	"reflect"
	"strings"
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

// undefinedableEncoder fakes Encoder so that
// the undefined Undefinedable fields are considered to be empty.
type undefinedableEncoder struct {
	ty  reflect2.Type
	org jsoniter.ValEncoder
}

func (e undefinedableEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	val := e.ty.UnsafeIndirect(ptr)
	return val.(IsUndefineder).IsUndefined()
}

func (e undefinedableEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	e.org.Encode(ptr, stream)
}

// fakingTagField implements reflect2.StructField interface,
// faking the struct tag to pretend it is always tagged with ,omitempty option.
type fakingTagField struct {
	reflect2.StructField
}

func (f fakingTagField) Tag() reflect.StructTag {
	t := f.StructField.Tag()
	if jsonTag, ok := t.Lookup("json"); !ok {
		return reflect.StructTag(`json:",omitempty"`)
	} else {
		splitted := strings.Split(jsonTag, ",")
		hasOmitempty := false
		for _, opt := range splitted {
			if opt == "omitempty" {
				hasOmitempty = true
				break
			}
		}

		if !hasOmitempty {
			return reflect.StructTag(`json:"` + strings.Join(splitted, ",") + `,omitempty"`)
		}
	}

	return t
}

// UndefinedableExtension is the extension for jsoniter.API.
// This forces jsoniter.API to skip undefined Undefinedable[T] when marshalling.
type UndefinedableExtension struct {
}

func (extension *UndefinedableExtension) UpdateStructDescriptor(structDescriptor *jsoniter.StructDescriptor) {
	if structDescriptor.Type.Implements(undefinedableTy) {
		return
	}

	for _, binding := range structDescriptor.Fields {
		if binding.Field.Type().Implements(undefinedableTy) {
			enc := binding.Encoder
			binding.Field = fakingTagField{binding.Field}
			binding.Encoder = undefinedableEncoder{ty: binding.Field.Type(), org: enc}
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

// MarshalFieldsJSON encodes v into JSON.
// It skips fields if those are undefined Undefinedable[T].
//
// v can be any type.
func MarshalFieldsJSON(v any) ([]byte, error) {
	return config.Marshal(v)
}

// UnmarshalFieldsJSON decodes data into v.
// v must be pointer type, return error otherwise.
//
// Currently this is almost same as json.Unmarshal.
// Future releases may change behavior of this function.
// It is safe to unmarshal data through this if v has at least an Undefinedable[T] field.
func UnmarshalFieldsJSON(data []byte, v any) error {
	return config.Unmarshal(data, v)
}
