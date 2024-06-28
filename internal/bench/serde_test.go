package bench

import (
	"encoding/json"
	"testing"

	jsonv2 "github.com/go-json-experiment/json"
)

type iund[T any] interface {
	IsUndefined() bool
	IsNull() bool
	IsDefined() bool
	Value() T
}

func assertUnd[T iund[int]](t *testing.T, ex int, und T) {
	t.Helper()
	switch x := expected[ex]; {
	case x == 0:
		if !und.IsUndefined() {
			t.Fatalf("not undefined")
		}
	case x == 1:
		if !und.IsNull() {
			t.Fatalf("not null")
		}
	default:
		if und.Value() != x {
			t.Fatalf("not equal, want = %d, got = %d", x, und.Value())
		}
	}
}

func TestSerdeMap(t *testing.T) {
	for i, input := range inputs {
		var (
			sSlice sampleSliceV1
			sMap   sampleMapV1
			bin    []byte
			err    error
		)
		{
			err = json.Unmarshal([]byte(input), &sSlice)
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal([]byte(input), &sMap)
			if err != nil {
				panic(err)
			}

			assertUnd(t, i, sSlice.U)
			assertUnd(t, i, sMap.U)
			bin, err = json.Marshal(sSlice)
			if err != nil {
				panic(err)
			}
			if string(bin) != input {
				t.Fatalf("not equal, want = %s, got = %s", input, string(bin))
			}
			bin, err = json.Marshal(sMap)
			if err != nil {
				panic(err)
			}
			if string(bin) != input {
				t.Fatalf("not equal, want = %s, got = %s", input, string(bin))
			}
		}

		{
			var (
				sSliceV2 = sampleSliceV2(sSlice)
				sMapV2   = sampleMapV2(sMap)
			)
			err = jsonv2.Unmarshal([]byte(input), &sSliceV2)
			if err != nil {
				panic(err)
			}
			err = jsonv2.Unmarshal([]byte(input), &sMapV2)
			if err != nil {
				panic(err)
			}
			assertUnd(t, i, sSliceV2.U)
			assertUnd(t, i, sMapV2.U)

			bin, err = jsonv2.Marshal(sSliceV2)
			if err != nil {
				panic(err)
			}
			if string(bin) != input {
				t.Fatalf("not equal, want = %s, got = %s", input, string(bin))
			}
			bin, err = jsonv2.Marshal(sMapV2)
			if err != nil {
				panic(err)
			}
			if string(bin) != input {
				t.Fatalf("not equal, want = %s, got = %s", input, string(bin))
			}
		}
	}
}
