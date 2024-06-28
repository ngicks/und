package testcase

import (
	"encoding/json"
	"testing"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/ngicks/und/option"
	"gotest.tools/v3/assert"
)

type Und[T any] interface {
	DoublePointer() **T
	IsDefined() bool
	IsNull() bool
	IsUndefined() bool
	IsZero() bool
	MarshalJSON() ([]byte, error)
	MarshalJSONV2(enc *jsontext.Encoder, opts jsonv2.Options) error
	Pointer() *T
	Value() T
	Unwrap() option.Option[option.Option[T]]
}

func TestUnd[T Und[U], U any](t *testing.T, defined, null, undefined T, value U, marshaled string) {
	t.Run("DoublePointer", func(t *testing.T) {
		var pp **U

		pp = defined.DoublePointer()
		assert.Equal(t, **pp, value)

		pp = null.DoublePointer()
		assert.Equal(t, *pp, (*U)(nil))

		pp = undefined.DoublePointer()
		assert.Equal(t, pp, (**U)(nil))
	})

	t.Run("IsDefined", func(t *testing.T) {
		assert.Assert(t, defined.IsDefined())
		assert.Assert(t, !null.IsDefined())
		assert.Assert(t, !undefined.IsDefined())
	})

	t.Run("IsNull", func(t *testing.T) {
		assert.Assert(t, !defined.IsNull())
		assert.Assert(t, null.IsNull())
		assert.Assert(t, !undefined.IsNull())
	})

	t.Run("IsUndefined", func(t *testing.T) {
		assert.Assert(t, !defined.IsUndefined())
		assert.Assert(t, !null.IsUndefined())
		assert.Assert(t, undefined.IsUndefined())
	})

	t.Run("IsZero", func(t *testing.T) {
		assert.Assert(t, !defined.IsZero())
		assert.Assert(t, !null.IsZero())
		assert.Assert(t, undefined.IsZero())
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		var (
			bin []byte
			err error
		)
		bin, err = json.Marshal(defined)
		assert.NilError(t, err)
		assert.Equal(t, string(bin), marshaled)

		bin, err = json.Marshal(null)
		assert.NilError(t, err)
		assert.Equal(t, string(bin), "null")

		bin, err = json.Marshal(undefined)
		assert.NilError(t, err)
		assert.Equal(t, string(bin), "null")
	})

	t.Run("MarshalJSONV2", func(t *testing.T) {
		var (
			bin []byte
			err error
		)
		bin, err = jsonv2.Marshal(defined)
		assert.NilError(t, err)
		assert.Equal(t, string(bin), marshaled)

		bin, err = jsonv2.Marshal(null)
		assert.NilError(t, err)
		assert.Equal(t, string(bin), "null")

		bin, err = jsonv2.Marshal(undefined)
		assert.NilError(t, err)
		assert.Equal(t, string(bin), "null")
	})

	t.Run("Pointer", func(t *testing.T) {
		var p *U
		p = defined.Pointer()
		assert.Equal(t, *p, value)
		p = null.Pointer()
		assert.Equal(t, p, (*U)(nil))
		p = undefined.Pointer()
		assert.Equal(t, p, (*U)(nil))
	})

	t.Run("Unwrap", func(t *testing.T) {
		assert.Equal(t, defined.Unwrap().Value().Value(), value)
		assert.Assert(t, null.Unwrap().IsSome())
		assert.Assert(t, null.Unwrap().Value().IsNone())
		assert.Assert(t, undefined.Unwrap().IsNone())
	})

	t.Run("Value", func(t *testing.T) {
		assert.Equal(t, defined.Value(), value)
		var zero U
		assert.Equal(t, null.Value(), zero)
		assert.Equal(t, undefined.Value(), zero)
	})
}
