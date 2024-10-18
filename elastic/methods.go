package elastic

import (
	"encoding/json"
	"encoding/xml"
	"log/slog"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/ngicks/und/option"
)

// portable methods that can be copied from github.com/ngicks/und/elastic into github.com/ngicks/und/sliceund/elastic

func FromValue[T any](t T) Elastic[T] {
	return FromOptions(option.Options[T]{option.Some(t)})
}

// FromPointer converts nil to undefined Elastic[T],
// or defined one whose internal value is dereferenced t.
func FromPointer[T any](t *T) Elastic[T] {
	if t == nil {
		return Undefined[T]()
	}
	return FromValue(*t)
}

// FromValues converts []T into an Elastic[T].
func FromValues[T any](ts []T) Elastic[T] {
	opts := make(option.Options[T], len(ts))
	for i, value := range ts {
		opts[i] = option.Some(value)
	}
	return FromOptions(opts)
}

// FromPointers converts []*T into an Elastic[T],
// treating nil as None[T], and non-nil as Some[T].
func FromPointers[T any](ps []*T) Elastic[T] {
	opts := make(option.Options[T], len(ps))
	for i, p := range ps {
		if p == nil {
			opts[i] = option.None[T]()
		} else {
			opts[i] = option.Some(*p)
		}
	}
	return FromOptions(opts)
}

// IsZero is an alias for IsUndefined.
func (e Elastic[T]) IsZero() bool {
	return e.IsUndefined()
}

func (e Elastic[T]) UndValidate() error {
	return e.inner().Value().UndValidate()
}

func (e Elastic[T]) UndCheck() error {
	return e.inner().Value().UndCheck()
}

// MarshalJSON implements json.Marshaler.
func (u Elastic[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.inner())
}

// UnmarshalJSON implements json.Unmarshaler.
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

// MarshalJSONV2 implements jsonv2.MarshalerV2.
func (e Elastic[T]) MarshalJSONV2(enc *jsontext.Encoder, opts jsonv2.Options) error {
	return jsonv2.MarshalEncode(enc, e.inner(), opts)
}

// MarshalXML implements xml.Marshaler.
func (e Elastic[T]) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return e.Unwrap().MarshalXML(enc, start)
}

// LogValue implements slog.LogValuer.
func (e Elastic[T]) LogValue() slog.Value {
	return e.Unwrap().LogValue()
}
