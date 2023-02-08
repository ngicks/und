package undefinedablejson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/mailru/easyjson/jwriter"
	syncparam "github.com/ngicks/type-param-common/sync-param"
)

var ErrIncorrectType = errors.New("incorrect")

// MarshalFieldsJSON encodes v into JSON.
// Some or all fields of v are expected to be Undefinedable[T any].
// There's no point using this function if v has no Undefinable[T] field,
// only being more expensive.
//
// It outputs `null` for a null Undefinable field, skips for an undefined Field.
//
// If v is not a struct, it returns a wrapped ErrIncorrectType error.
func MarshalFieldsJSON(v any) ([]byte, error) {
	rv := reflect.ValueOf(v)
	rt := rv.Type()

	if rt.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w. want = struct, but is %s", ErrIncorrectType, rt.Kind())
	}

	marshalFields, err := loadOrCreateSerdeMeta(rv)
	if err != nil {
		return nil, err
	}

	writer := jwriter.Writer{}
	writer.RawByte('{')

	appendComma := false

	for _, fieldName := range marshalFields.layout {
		fieldInfo := marshalFields.fields[fieldName]

		frv := rv.Field(fieldInfo.index)
		valueInterface := frv.Interface()

		if fieldInfo.implementsIsUndefined {
			if valueInterface.(interface{ IsUndefined() bool }).IsUndefined() {
				continue
			}
		} else if fieldInfo.taggedOmitempty {
			if IsEmpty(frv) {
				// skip this field.
				continue
			}
		}

		if appendComma {
			writer.RawString(",")
		}
		appendComma = true

		shouldSkipFieldName := false
		marshalled, err := fieldInfo.marshaller(valueInterface)
		if fieldInfo.embedded && !fieldInfo.taggedFieldName {
			// Unwrap opening and closing braces only if field is struct.
			if frv.Kind() == reflect.Struct &&
				(marshalled[0] == '{' || marshalled[len(marshalled)-1] == '}') {
				shouldSkipFieldName = true
				marshalled = marshalled[1 : len(marshalled)-1]
			}
		}

		// skip field name only if not tagged and is struct.
		if !shouldSkipFieldName {
			writer.String(fieldInfo.name)
			writer.RawString(":")
		}

		shouldQuote := fieldInfo.quote && string(marshalled) != string(nullByte)
		if shouldQuote {
			writer.RawString("\"")
		}
		writer.Raw(marshalled, err)
		if shouldQuote {
			writer.RawString("\"")
		}
	}

	writer.RawString("}")

	if writer.Error != nil {
		return nil, writer.Error
	}

	var buf bytes.Buffer
	buf.Grow(writer.Size())
	if _, err := writer.DumpTo(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type serdeMeta struct {
	layout           []string // ordering is first written field to the last; layout order. contents are one defined by json:"" tag or field name.
	fields           map[string]serdeFieldInfo
	untaggedEmbedded []string
}

type serdeFieldInfo struct {
	index                 int
	name                  string
	taggedFieldName       bool
	taggedOmitempty       bool
	quote                 bool
	embedded              bool
	implementsIsUndefined bool
	marshaller            func(v any) ([]byte, error)
}

var serdeInfoCache syncparam.Map[reflect.Type, serdeMeta]

func loadOrCreateSerdeMeta(v reflect.Value) (serdeMeta, error) {
	fields, ok := serdeInfoCache.Load(v.Type())
	if ok {
		return fields, nil
	}
	meta, err := readFieldInfo(v)
	if err != nil {
		return serdeMeta{}, err
	}
	fields, _ = serdeInfoCache.LoadOrStore(v.Type(), meta)
	return fields, nil
}

func readFieldInfo(rv reflect.Value) (serdeMeta, error) {
	rt := rv.Type()

	layout := make([]string, 0, rv.NumField())
	fields := make(map[string]serdeFieldInfo, rv.NumField())
	untaggedEmbedded := make([]string, 0, rv.NumField())

	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)

		if !field.IsExported() {
			continue
		}

		var fieldInfo serdeFieldInfo

		fieldInfo.index = i

		fieldName, options, tagged, shouldSkip := GetFieldName(field)
		frv := rv.Field(i)
		valueInterface := frv.Interface()

		if shouldSkip {
			// tagged as "-"
			continue
		}

		fieldInfo.name = fieldName
		fieldInfo.taggedFieldName = tagged
		fieldInfo.taggedOmitempty = OptContain(options, "omitempty")
		_, fieldInfo.implementsIsUndefined = valueInterface.(interface {
			IsUndefined() bool
		})
		fieldInfo.embedded = field.Anonymous

		if field.Anonymous {
			// If the embedded (Anonymous) field implements json.Marshaler:
			// json.Marshal finds that MarshalJSON because the Go's method look up mechanism works in that way,
			// meaning the result of marshalling is to be that method's result.
			//
			// This function explicitly forbids that.
			if _, ok := valueInterface.(json.Marshaler); ok {
				return serdeMeta{}, fmt.Errorf("%w. embedded field implements json.Marshaler", ErrIncorrectType)
			}

			if field.Type.Kind() == reflect.Struct {
				_, err := loadOrCreateSerdeMeta(frv)
				if err != nil {
					return serdeMeta{}, err
				}

				fieldInfo.marshaller = func(v any) ([]byte, error) {
					// the embedded struct field receive same treatment.
					return MarshalFieldsJSON(v)
				}

				if !tagged {
					untaggedEmbedded = append(untaggedEmbedded, fieldName)
				}
			}
		}

		if fieldInfo.marshaller == nil {
			fieldInfo.marshaller = func(v any) ([]byte, error) {
				return json.Marshal(v)
			}
		}

		fieldInfo.quote = (OptContain(options, "string") ||
			OptContain(field.Tag.Get("und"), "string")) &&
			shouldQuote(field.Type, valueInterface)

		layout = append(layout, fieldInfo.name)
		fields[fieldInfo.name] = fieldInfo
	}

	return serdeMeta{layout, fields, untaggedEmbedded}, nil
}

func UnmarshalFieldsJSON(data []byte, v any) error {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &json.InvalidUnmarshalError{Type: rv.Type()}
	}

	return unmarshalFieldsJSON(data, rv)
}

func unmarshalFieldsJSON(data []byte, rv reflect.Value) error {
	rv = reflect.Indirect(rv)
	rt := rv.Type()

	if rt.Kind() != reflect.Struct {
		return fmt.Errorf("%w. want = struct, but is %s", ErrIncorrectType, rt.Kind())
	}

	serdeInfo, err := loadOrCreateSerdeMeta(rv)
	if err != nil {
		return err
	}

	// TODO: We only need to do this if we find keys of embedded field in input data.
	for _, fieldName := range serdeInfo.untaggedEmbedded {
		// Recursion is needed because embedded fields may have another embedded field.
		// Flatten them is a bit harder to implement.
		// TODO: elaborate more?
		frv := rv.Field(serdeInfo.fields[fieldName].index)
		err := unmarshalFieldsJSON(data, frv)
		if err != nil {
			return err
		}
	}

	return jsonparser.ObjectEach(
		data,
		func(key, value []byte, dataType jsonparser.ValueType, offset int) error {
			info, has := serdeInfo.fields[string(key)]
			if !has {
				return nil
			}

			if dataType == jsonparser.String {
				// jsonparser trims wrapping double quotations. Get it back here.
				value = data[offset-len(value)-2 : offset]
			}

			if info.quote && string(value) != string(nullByte) {
				value = bytes.Trim(value, "\"")
			}

			frv := rv.Field(info.index)

			v := frv
			if v.Kind() != reflect.Pointer && v.Type().Name() != "" && v.CanAddr() {
				// adder this value so that we can find method of *T, not only ones for T.
				v = v.Addr()
			}

			if info.embedded && frv.Type().Kind() == reflect.Struct {
				// tagged embedded field.
				err := unmarshalFieldsJSON(value, v)
				if err != nil {
					return err
				}
			}

			if decoder, ok := v.Interface().(json.Unmarshaler); ok {
				err := decoder.UnmarshalJSON(value)
				if err != nil {
					return err
				}
			} else {
				internalValue := v.Interface()
				err := json.Unmarshal(value, internalValue)
				if err != nil {
					return err
				}
			}

			return nil
		},
	)

}

func GetFieldName(field reflect.StructField) (fieldName string, options string, tagged bool, shouldSkip bool) {
	tagged = true
	fieldName, options, shouldSkip = GetJsonTag(field)
	if len(fieldName) == 0 {
		tagged = false
		fieldName = field.Name
	}
	return fieldName, options, tagged, shouldSkip
}

func GetJsonTag(field reflect.StructField) (name string, opt string, shouldSkip bool) {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", "", true
	}
	name, opt, _ = strings.Cut(tag, ",")
	return name, opt, false
}

func OptContain(options string, target string) bool {
	if len(options) == 0 {
		return false
	}
	var opt string
	for len(options) != 0 {
		opt, options, _ = strings.Cut(options, ",")
		if opt == target {
			return true
		}
	}
	return false
}

func shouldQuote(ty reflect.Type, value any) bool {
	if IsQuotable(ty) {
		return true
	}
	if quotable, ok := value.(interface{ IsQuotable() bool }); ok {
		return quotable.IsQuotable()
	}
	return false
}

func IsQuotable(ty reflect.Type) bool {
	// string options work only for
	// string, floating point, integer, or boolean types.
	switch ty.Kind() {
	case reflect.String,
		reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Bool:
		return true
	}

	return false
}

// IsEmpty reports true if v should be skipped when tagged with omitempty, false otherwise.
func IsEmpty(v reflect.Value) bool {
	switch v.Kind() {
	// false
	case reflect.Bool:
		return !v.Bool()
		// 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return math.Float64bits(v.Float()) == 0
		// empty array
	case reflect.Array:
		return v.Len() == 0
		// nil interface or pointer
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
		// empty map, slice, string
	case reflect.Map, reflect.Slice:
		return v.IsNil() || v.Len() == 0
	case reflect.String:
		return v.Len() == 0
	}
	return false
}
