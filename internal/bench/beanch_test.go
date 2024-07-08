package bench

import (
	"encoding/json"
	"fmt"
	"testing"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/ngicks/und"
	"github.com/ngicks/und/sliceund"
	"github.com/oapi-codegen/nullable"
)

var inputs = []string{
	`{"Pad2":123}`,
	`{"U":null}`,
	`{"Pad1":445,"U":123}`,
}

var (
	unmarshaledNullableV1 []sampleNullableV1
	unmarshaledNullableV2 []sampleNullableV2
	unmarshaledMapV1      []sampleMapV1
	unmarshaledMapV2      []sampleMapV2
	unmarshaledSliceV1    []sampleSliceV1
	unmarshaledSliceV2    []sampleSliceV2
	unmarshaledNonSliceV2 []sampleNonSliceV2
)

func init() {
	preheat(&unmarshaledNullableV1)
	preheat(&unmarshaledNullableV2)
	preheat(&unmarshaledMapV1)
	preheat(&unmarshaledMapV2)
	preheat(&unmarshaledSliceV1)
	preheat(&unmarshaledSliceV2)
	preheat(&unmarshaledNonSliceV2)
}

func preheat[T any](storage *[]T) {
	for _, input := range inputs {
		var s T
		err := json.Unmarshal([]byte(input), &s)
		if err != nil {
			panic(err)
		}
		*storage = append(*storage, s)
	}
}

type sampleNullableV1 struct {
	Pad1 int                    `json:",omitempty"`
	U    nullable.Nullable[int] `json:",omitempty"`
	Pad2 int                    `json:",omitempty"`
}

type sampleNullableV2 struct {
	Pad1 int                    `json:",omitzero"`
	U    nullable.Nullable[int] `json:",omitzero"`
	Pad2 int                    `json:",omitzero"`
}

type sampleMapV1 struct {
	Pad1 int         `json:",omitempty"`
	U    undMap[int] `json:",omitempty"`
	Pad2 int         `json:",omitempty"`
}

type sampleMapV2 struct {
	Pad1 int         `json:",omitzero"`
	U    undMap[int] `json:",omitzero"`
	Pad2 int         `json:",omitzero"`
}

type sampleSliceV1 struct {
	Pad1 int               `json:",omitempty"`
	U    sliceund.Und[int] `json:",omitempty"`
	Pad2 int               `json:",omitempty"`
}

type sampleSliceV2 struct {
	Pad1 int               `json:",omitzero"`
	U    sliceund.Und[int] `json:",omitzero"`
	Pad2 int               `json:",omitzero"`
}

type sampleNonSliceV2 struct {
	Pad1 int          `json:",omitzero"`
	U    und.Und[int] `json:",omitzero"`
	Pad2 int          `json:",omitzero"`
}

func benchMarshalV1[T any](b *testing.B, inputs []T) {
	b.Helper()
	for range b.N {
		for _, input := range inputs {
			_, err := json.Marshal(input)
			if err != nil {
				panic(err)
			}
		}
	}
}

func benchUnmarshalV1[T any](b *testing.B) {
	b.Helper()
	for range b.N {
		for _, input := range inputs {
			var s T
			err := json.Unmarshal([]byte(input), &s)
			if err != nil {
				panic(err)
			}
		}
	}
}

func benchSerdeV1[T any](b *testing.B) {
	b.Helper()
	for range b.N {
		for _, input := range inputs {
			var s T
			err := json.Unmarshal([]byte(input), &s)
			if err != nil {
				panic(err)
			}
			bin, err := json.Marshal(s)
			if err != nil {
				panic(err)
			}
			if input != string(bin) {
				panic(fmt.Sprintf("not equal, left = %s, right = %s", input, bin))
			}
		}
	}
}

func benchMarshalV2[T any](b *testing.B, inputs []T) {
	b.Helper()
	for range b.N {
		for _, input := range inputs {
			_, err := jsonv2.Marshal(input)
			if err != nil {
				panic(err)
			}
		}
	}
}

func benchUnmarshalV2[T any](b *testing.B) {
	b.Helper()
	for range b.N {
		for _, input := range inputs {
			var s T
			err := jsonv2.Unmarshal([]byte(input), &s)
			if err != nil {
				panic(err)
			}
		}
	}
}

func benchSerdeV2[T any](b *testing.B) {
	b.Helper()
	for range b.N {
		for _, input := range inputs {
			var s T
			err := jsonv2.Unmarshal([]byte(input), &s)
			if err != nil {
				panic(err)
			}
			bin, err := jsonv2.Marshal(s)
			if err != nil {
				panic(err)
			}
			if input != string(bin) {
				panic(fmt.Sprintf("not equal, left = %s, right = %s", input, bin))
			}
		}
	}
}

func BenchmarkUnd_Marshal(b *testing.B) {
	b.Run("NullableV1", func(b *testing.B) {
		benchMarshalV1[sampleNullableV1](b, unmarshaledNullableV1)
	})
	b.Run("MapV1", func(b *testing.B) {
		benchMarshalV1[sampleMapV1](b, unmarshaledMapV1)
	})
	b.Run("SliceV1", func(b *testing.B) {
		benchMarshalV1[sampleSliceV1](b, unmarshaledSliceV1)
	})
	b.Run("NullableV2", func(b *testing.B) {
		benchMarshalV2[sampleNullableV2](b, unmarshaledNullableV2)
	})
	b.Run("MapV2", func(b *testing.B) {
		benchMarshalV2[sampleMapV2](b, unmarshaledMapV2)
	})
	b.Run("SliceV2", func(b *testing.B) {
		benchMarshalV2[sampleSliceV2](b, unmarshaledSliceV2)
	})
	b.Run("NonSliceV2", func(b *testing.B) {
		benchMarshalV2[sampleNonSliceV2](b, unmarshaledNonSliceV2)
	})
}
func BenchmarkUnd_Unmarshal(b *testing.B) {
	b.Run("NullableV1", func(b *testing.B) {
		benchUnmarshalV1[sampleNullableV1](b)
	})
	b.Run("MapV1", func(b *testing.B) {
		benchUnmarshalV1[sampleMapV1](b)
	})
	b.Run("SliceV1", func(b *testing.B) {
		benchUnmarshalV1[sampleSliceV1](b)
	})
	b.Run("NullableV2", func(b *testing.B) {
		benchUnmarshalV2[sampleNullableV2](b)
	})
	b.Run("MapV2", func(b *testing.B) {
		benchUnmarshalV2[sampleMapV2](b)
	})
	b.Run("SliceV2", func(b *testing.B) {
		benchUnmarshalV2[sampleSliceV2](b)
	})
	b.Run("NonSliceV2", func(b *testing.B) {
		benchUnmarshalV2[sampleNonSliceV2](b)
	})
}

func BenchmarkUnd_Serde(b *testing.B) {
	b.Run("NullableV1", func(b *testing.B) {
		benchSerdeV1[sampleNullableV1](b)
	})
	b.Run("MapV1", func(b *testing.B) {
		benchSerdeV1[sampleMapV1](b)
	})
	b.Run("SliceV1", func(b *testing.B) {
		benchSerdeV1[sampleSliceV1](b)
	})
	b.Run("NullableV2", func(b *testing.B) {
		benchSerdeV2[sampleNullableV2](b)
	})
	b.Run("MapV2", func(b *testing.B) {
		benchSerdeV2[sampleMapV2](b)
	})
	b.Run("SliceV2", func(b *testing.B) {
		benchSerdeV2[sampleSliceV2](b)
	})
	b.Run("NonSliceV2", func(b *testing.B) {
		benchSerdeV2[sampleNonSliceV2](b)
	})
}
