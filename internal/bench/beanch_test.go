package bench

import (
	"encoding/json"
	"testing"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/ngicks/und/sliceund"
	"github.com/oapi-codegen/nullable"
)

var inputs = []string{
	`{"Pad2":123}`,
	`{"U":null}`,
	`{"Pad1":445,"U":123}`,
}

var expected = []int{
	0,   // undefined
	1,   // null
	123, // value
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

func BenchmarkSerdeNullableV1(b *testing.B) {
	for range b.N {
		for _, input := range inputs {
			var s sampleNullableV1
			err := json.Unmarshal([]byte(input), &s)
			if err != nil {
				panic(err)
			}
			_, err = json.Marshal(s)
			if err != nil {
				panic(err)
			}
		}
	}
}

func BenchmarkSerdeMapV1(b *testing.B) {
	for range b.N {
		for _, input := range inputs {
			var s sampleMapV1
			err := json.Unmarshal([]byte(input), &s)
			if err != nil {
				panic(err)
			}
			_, err = json.Marshal(s)
			if err != nil {
				panic(err)
			}
		}
	}
}

func BenchmarkSerdeSliceV1(b *testing.B) {
	for range b.N {
		for _, input := range inputs {
			var s sampleSliceV1
			err := json.Unmarshal([]byte(input), &s)
			if err != nil {
				panic(err)
			}
			_, err = json.Marshal(s)
			if err != nil {
				panic(err)
			}
		}
	}
}

func BenchmarkSerdeNullableV2(b *testing.B) {
	for range b.N {
		for _, input := range inputs {
			var s sampleNullableV2
			err := jsonv2.Unmarshal([]byte(input), &s)
			if err != nil {
				panic(err)
			}
			_, err = json.Marshal(s)
			if err != nil {
				panic(err)
			}
		}
	}
}

func BenchmarkSerdeMapV2(b *testing.B) {
	for range b.N {
		for _, input := range inputs {
			var s sampleMapV2
			err := jsonv2.Unmarshal([]byte(input), &s)
			if err != nil {
				panic(err)
			}
			_, err = jsonv2.Marshal(s)
			if err != nil {
				panic(err)
			}
		}
	}
}

func BenchmarkSerdeSliceV2(b *testing.B) {
	for range b.N {
		for _, input := range inputs {
			var s sampleSliceV2
			err := jsonv2.Unmarshal([]byte(input), &s)
			if err != nil {
				panic(err)
			}
			_, err = jsonv2.Marshal(s)
			if err != nil {
				panic(err)
			}
		}
	}
}
