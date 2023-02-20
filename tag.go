package undefinedablejson

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrUnpairedKey = errors.New("unpaired key")
)

type Tag struct {
	Key   string
	Value string
}

func (t Tag) Flatten() string {
	return t.Key + ":" + strconv.Quote(t.Value)
}

func ParseStructTag(tag reflect.StructTag) ([]Tag, error) {
	var out []Tag

	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			return nil, fmt.Errorf("%w: input has no paired value, rest = %s", ErrUnpairedKey, string(tag))
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			return nil, fmt.Errorf("%w: name = %s has no paired value, rest = %s", ErrUnpairedKey, name, string(tag))
		}
		quotedValue := string(tag[:i+1])
		tag = tag[i+1:]

		value, err := strconv.Unquote(quotedValue)
		if err != nil {
			return nil, err
		}
		out = append(out, Tag{Key: name, Value: value})
	}

	return out, nil
}

func FlattenStructTag(tags []Tag) reflect.StructTag {
	var buf strings.Builder
	for _, tag := range tags {
		buf.Write([]byte(tag.Flatten()))
		buf.WriteByte(' ')
	}
	return reflect.StructTag(buf.String())
}
