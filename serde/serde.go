package serde

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

var config = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
}.Froze()

func init() {
	config.RegisterExtension(&UndefinedSkipperExtension{})
}

// Marshal encodes v into JSON.
// It skips fields if those are undefined Undefinedable[T].
//
// v can be any type.
func Marshal(v any) ([]byte, error) {
	return config.Marshal(v)
}

func NewEncoder(w io.Writer) *jsoniter.Encoder {
	return config.NewEncoder(w)
}

// Unmarshal decodes data into v.
// v must be pointer type, return error otherwise.
//
// Currently this is almost same as json.Unmarshal.
// Future releases may change behavior of this function.
// It is safe to unmarshal data through this if v has at least an Undefinedable[T] field.
func Unmarshal(data []byte, v any) error {
	return config.Unmarshal(data, v)
}

func NewDecoder(r io.Reader) *jsoniter.Decoder {
	return config.NewDecoder(r)
}
