package bench

import (
	"encoding/json"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

type undMap[T any] map[bool]T

func (u undMap[T]) IsZero() bool {
	return u.IsUndefined()
}

func (u undMap[T]) IsDefined() bool {
	if len(u) == 0 {
		return false
	}
	_, ok := u[true]
	return ok
}

func (u undMap[T]) IsNull() bool {
	if len(u) == 0 {
		return false
	}
	_, ok := u[false]
	return ok
}

func (u undMap[T]) IsUndefined() bool {
	return len(u) == 0
}

func (u undMap[T]) Value() T {
	if u.IsDefined() {
		return u[true]
	}
	var zero T
	return zero
}

var _ json.Marshaler = undMap[any]{}

func (u undMap[T]) MarshalJSON() ([]byte, error) {
	if !u.IsDefined() {
		return []byte(`null`), nil
	}
	return json.Marshal(u[true])
}

var _ json.Unmarshaler = (*undMap[any])(nil)

func (u *undMap[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		var zero T
		if *u == nil {
			*u = map[bool]T{false: zero}
		} else {
			delete(*u, true)
			(*u)[false] = zero
		}
		return nil
	}

	var t T
	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}

	if *u == nil {
		*u = map[bool]T{true: t}
	} else {
		delete(*u, false)
		(*u)[true] = t
	}
	return nil
}

var _ jsonv2.MarshalerV2 = undMap[any]{}

func (u undMap[T]) MarshalJSONV2(enc *jsontext.Encoder, opts jsonv2.Options) error {
	if !u.IsDefined() {
		return enc.WriteToken(jsontext.Null)
	}
	return jsonv2.MarshalEncode(enc, u.Value(), opts)
}

var _ jsonv2.UnmarshalerV2 = (*undMap[any])(nil)

func (u *undMap[T]) UnmarshalJSONV2(dec *jsontext.Decoder, opts jsonv2.Options) error {
	if dec.PeekKind() == 'n' {
		err := dec.SkipValue()
		if err != nil {
			return err
		}
		var zero T
		if *u == nil {
			*u = map[bool]T{false: zero}
		} else {
			delete(*u, true)
			(*u)[false] = zero
		}
		return nil
	}
	var t T
	err := jsonv2.UnmarshalDecode(dec, &t, opts)
	if err != nil {
		return err
	}

	if *u == nil {
		*u = map[bool]T{true: t}
	} else {
		delete(*u, false)
		(*u)[true] = t
	}
	return nil
}
