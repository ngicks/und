package bench

import (
	"encoding/json"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/ngicks/und"
)

type undSlice[T any] []und.Option[T]

func (u undSlice[T]) IsZero() bool {
	return u.IsUndefined()
}

func (u undSlice[T]) IsDefined() bool {
	return len(u) > 0 && u[0].IsSome()
}

func (u undSlice[T]) IsNull() bool {
	return len(u) > 0 && u[0].IsNone()
}

func (u undSlice[T]) IsUndefined() bool {
	return len(u) == 0
}

func (u undSlice[T]) Value() T {
	if u.IsDefined() {
		return u[0].Value()
	}
	var zero T
	return zero
}

var _ json.Marshaler = undSlice[any]{}

func (u undSlice[T]) MarshalJSON() ([]byte, error) {
	if !u.IsDefined() {
		return []byte(`null`), nil
	}
	return json.Marshal(u[0].Value())
}

var _ json.Unmarshaler = (*undSlice[any])(nil)

func (u *undSlice[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		if len(*u) == 0 {
			*u = []und.Option[T]{und.None[T]()}
		} else {
			(*u)[0] = und.None[T]()
		}
		return nil
	}

	var t T
	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}

	if len(*u) == 0 {
		*u = []und.Option[T]{und.Some(t)}
	} else {
		(*u)[0] = und.Some(t)
	}
	return nil
}

var _ jsonv2.MarshalerV2 = undSlice[any]{}

func (u undSlice[T]) MarshalJSONV2(enc *jsontext.Encoder, opts jsonv2.Options) error {
	if !u.IsDefined() {
		return enc.WriteToken(jsontext.Null)
	}
	return jsonv2.MarshalEncode(enc, u.Value(), opts)
}

var _ jsonv2.UnmarshalerV2 = (*undSlice[any])(nil)

func (u *undSlice[T]) UnmarshalJSONV2(dec *jsontext.Decoder, opts jsonv2.Options) error {
	if dec.PeekKind() == 'n' {
		err := dec.SkipValue()
		if err != nil {
			return err
		}
		if len(*u) == 0 {
			*u = []und.Option[T]{und.None[T]()}
		} else {
			(*u)[0] = und.None[T]()
		}
		return nil
	}
	var t T
	err := jsonv2.UnmarshalDecode(dec, &t, opts)
	if err != nil {
		return err
	}

	if len(*u) == 0 {
		*u = []und.Option[T]{und.Some(t)}
	} else {
		(*u)[0] = und.Some(t)
	}
	return nil
}
