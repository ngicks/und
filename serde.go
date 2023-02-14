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

// UndefinedableExtension
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
// Some or all fields of v are expected to be Undefinedable[T any].
// There's no point using this function if v has no Undefinable[T] field,
// only being more expensive.
//
// It outputs `null` for a null Undefinable field, skips for an undefined Field.
//
// If v is not a struct, it returns a wrapped ErrIncorrectType error.
func MarshalFieldsJSON(v any) ([]byte, error) {
	return config.Marshal(v)
}

func UnmarshalFieldsJSON(data []byte, v any) error {
	return config.Unmarshal(data, v)
}
