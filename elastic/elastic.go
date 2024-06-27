package elastic

import (
	"encoding/json"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
)

type Elastic[T any] struct {
	v und.Und[option.Options[T]]
}

func Null[T any]() Elastic[T] {
	return Elastic[T]{
		v: und.Null[option.Options[T]](),
	}
}

func Undefined[T any]() Elastic[T] {
	return Elastic[T]{
		v: und.Undefined[option.Options[T]](),
	}
}

func FromOptions[T any, Opts ~[]option.Option[T]](options Opts) Elastic[T] {
	return Elastic[T]{
		v: und.Defined(option.Options[T](options)),
	}
}

func FromPointer[T any](t *T) Elastic[T] {
	if t == nil {
		return Undefined[T]()
	}
	return FromValue(*t)
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

func FromValue[T any](t T) Elastic[T] {
	return FromOptions(option.Options[T]{option.Some(t)})
}

func FromValues[T any](ts []T) Elastic[T] {
	opts := make(option.Options[T], len(ts))
	for i, value := range ts {
		opts[i] = option.Some(value)
	}
	return FromOptions(opts)
}

func (e Elastic[T]) inner() und.Und[option.Options[T]] {
	return e.v
}

func (e Elastic[T]) IsZero() bool {
	return e.IsUndefined()
}

func (e Elastic[T]) IsDefined() bool {
	return e.v.IsDefined()
}

func (e Elastic[T]) IsNull() bool {
	return e.v.IsNull()
}

func (e Elastic[T]) IsUndefined() bool {
	return e.v.IsUndefined()
}

func (e Elastic[T]) Equal(other Elastic[T]) bool {
	return e.v.Equal(other.v)
}

func (e Elastic[T]) Value() T {
	if e.IsDefined() {
		vs := e.v.Value()
		if len(vs) > 0 {
			return vs[0].Value()
		}
	}
	var zero T
	return zero
}

func (e Elastic[T]) Values() []T {
	if !e.IsDefined() {
		return []T(nil)
	}
	opts := e.v.Value()
	vs := make([]T, len(opts))
	for i, opt := range opts {
		vs[i] = opt.Value()
	}
	return vs
}

func (e Elastic[T]) Pointer() *T {
	if e.IsDefined() {
		vs := e.v.Value()
		if len(vs) > 0 {
			v := vs[0].Value()
			return &v
		}
	}
	return nil
}

func (e Elastic[T]) Pointers() []*T {
	if !e.IsDefined() {
		return []*T(nil)
	}
	opts := e.v.Value()
	ptrs := make([]*T, len(opts))
	for i, opt := range opts {
		ptrs[i] = opt.Pointer()
	}
	return ptrs
}

func (u Elastic[T]) Unwrap() und.Und[option.Options[T]] {
	return u.v
}

func (e Elastic[T]) Map(f func(und.Und[option.Options[T]]) und.Und[option.Options[T]]) Elastic[T] {
	return Elastic[T]{v: f(e.v)}
}

func (u Elastic[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.inner())
}

func (e *Elastic[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*e = Null[T]()
		return nil
	}

	if len(data) > 2 && data[0] == '[' {
		var t option.Options[T]
		err := json.Unmarshal(data, &t)
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

func (u *Elastic[T]) UnmarshalJSONV2(dec *jsontext.Decoder, opts jsonv2.Options) error {
	if dec.PeekKind() == 'n' {
		err := dec.SkipValue()
		if err != nil {
			return err
		}
		*u = Null[T]()
		return nil
	}

	var t option.Options[T]
	err := jsonv2.UnmarshalDecode(dec, &t, opts)
	if err != nil {
		return err
	}

	*u = FromOptions(t)

	return nil
}
