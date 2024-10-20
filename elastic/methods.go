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

// FromValue returns Elastic[T] with single some value.
func FromValue[T any](t T) Elastic[T] {
	return FromOptions(option.Some(t))
}

// FromPointer converts nil to undefined Elastic[T],
// or defined one whose internal value is dereferenced t.
//
// If you need to keep t as pointer, use [WrapPointer] instead.
func FromPointer[T any](t *T) Elastic[T] {
	if t == nil {
		return Undefined[T]()
	}
	return FromValue(*t)
}

// WrapPointer converts *T into Elastic[*T].
// The elastic value is defined if t is non nil, undefined otherwise.
//
// If you want t to be dereferenced, use [FromPointer] instead.
func WrapPointer[T any](t *T) Elastic[*T] {
	if t == nil {
		return Undefined[*T]()
	}
	return FromValue(t)
}

// FromValues converts variadic T values into an Elastic[T].
func FromValues[T any](ts ...T) Elastic[T] {
	opts := make(option.Options[T], len(ts))
	for i, value := range ts {
		opts[i] = option.Some(value)
	}
	return FromOptions(opts...)
}

// FromPointers converts variadic *T values into an Elastic[T],
// treating nil as None[T], and non-nil as Some[T].
//
// If you need to keep t-s as pointer, use [WrapPointers] instead.
func FromPointers[T any](ps ...*T) Elastic[T] {
	opts := make(option.Options[T], len(ps))
	for i, p := range ps {
		opts[i] = option.FromPointer(p)
	}
	return FromOptions(opts...)
}

// FromPointers converts variadic *T values into an Elastic[*T],
// treating nil as None[*T], and non-nil as Some[*T].
//
// If you need t-s to be dereferenced, use [FromPointers] instead.
func WrapPointers[T any](ps ...*T) Elastic[*T] {
	opts := make(option.Options[*T], len(ps))
	for i, p := range ps {
		opts[i] = option.WrapPointer(p)
	}
	return FromOptions(opts...)
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
			*e = FromOptions(t...)
			return nil
		}
	}

	var t option.Option[T]
	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}
	*e = FromOptions(t)
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
