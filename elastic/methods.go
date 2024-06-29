package elastic

import (
	"encoding/json"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/ngicks/und/option"
)

// portable methods that can be copied from github.com/ngicks/und/elastic into github.com/ngicks/und/sliceund/elastic

func FromValue[T any](t T) Elastic[T] {
	return FromOptions(option.Options[T]{option.Some(t)})
}

func FromPointer[T any](t *T) Elastic[T] {
	if t == nil {
		return Undefined[T]()
	}
	return FromValue(*t)
}

func FromValues[T any](ts []T) Elastic[T] {
	opts := make(option.Options[T], len(ts))
	for i, value := range ts {
		opts[i] = option.Some(value)
	}
	return FromOptions(opts)
}

func FromPointers[T any](ps []*T) Elastic[T] {
	opts := make(option.Options[T], len(ps))
	for _, p := range ps {
		if p == nil {
			opts = append(opts, option.None[T]())
		} else {
			opts = append(opts, option.Some(*p))
		}
	}
	return FromOptions(opts)
}

func (e Elastic[T]) IsZero() bool {
	return e.IsUndefined()
}

func (u Elastic[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.inner())
}

func (e *Elastic[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*e = Null[T]()
		return nil
	}

	if len(data) >= 2 && data[0] == '[' {
		var t option.Options[T]
		err := json.Unmarshal(data, &t)
		// might be T is []U, and this fails
		// since it should've been [[...data...],[...data...]]
		if err == nil {
			*e = FromOptions(t)
			return nil
		}
	}

	var t option.Option[T]
	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}
	*e = FromOptions(option.Options[T]{t})
	return nil
}

func (e Elastic[T]) MarshalJSONV2(enc *jsontext.Encoder, opts jsonv2.Options) error {
	return jsonv2.MarshalEncode(enc, e.inner(), opts)
}
