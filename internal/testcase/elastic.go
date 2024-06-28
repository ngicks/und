package testcase

import (
	"encoding/json"
	"testing"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/ngicks/und/option"
	"gotest.tools/v3/assert"
)

type Ela[T any] interface {
	// Equal(other Elastic[T]) bool
	IsDefined() bool
	IsNull() bool
	IsUndefined() bool
	IsZero() bool
	// Map(f func(und.Und[option.Options[T]]) und.Und[option.Options[T]]) Elastic[T]
	MarshalJSON() ([]byte, error)
	MarshalJSONV2(enc *jsontext.Encoder, opts jsonv2.Options) error
	Pointer() *T
	Pointers() []*T
	// UnmarshalJSON(data []byte) error
	// Unwrap() und.Und[option.Options[T]]
	Value() T
	Values() []T
}

func TestElastic_non_addressable[T Ela[U], U comparable](
	t *testing.T,
	defined, null, undefined T,
	values []option.Option[U], marshaled string,
) {
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
		assert.Equal(t, *p, values[0].Value())
		p = null.Pointer()
		assert.Assert(t, p == nil)
		p = undefined.Pointer()
		assert.Assert(t, p == nil)
	})

	t.Run("Pointers", func(t *testing.T) {
		pp := defined.Pointers()
		assert.Assert(t, len(pp) == len(values))
		for i, p := range pp {
			opt := values[i]
			if opt.IsNone() {
				assert.Assert(t, p == nil)
			} else {
				assert.Equal(t, *p, opt.Value())
			}
		}
		assert.Assert(t, null.Pointers() == nil)
		assert.Assert(t, undefined.Pointers() == nil)
	})

	t.Run("Value", func(t *testing.T) {
		assert.Equal(t, defined.Value(), values[0].Value())
		var zero U
		assert.Equal(t, null.Value(), zero)
		assert.Equal(t, undefined.Value(), zero)
	})

	t.Run("Values", func(t *testing.T) {
		vals := defined.Values()
		assert.Assert(t, len(vals) == len(values))
		for i, v := range vals {
			assert.Equal(t, v, values[i].Value())
		}
		var zero U
		assert.Equal(t, null.Value(), zero)
		assert.Equal(t, undefined.Value(), zero)
	})
}
