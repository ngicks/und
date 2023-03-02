package serde

// This package uses modified Go programming language standard library.
// So keep it credited.
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Modified parts are governed by a license that is described in ../LICENSE.

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

	out := buf.String()
	if len(out) > 0 {
		out = out[:len(out)-1]
	}
	return reflect.StructTag(out)
}

func FakeOmitempty(t reflect.StructTag) reflect.StructTag {
	tags, err := ParseStructTag(t)
	if err != nil {
		panic(err)
	}

	hasTag := false
	for i := 0; i < len(tags); i++ {
		if tags[i].Key != "json" {
			continue
		}

		hasTag = true

		hasOmitempty := false

		// skip first opt since it is field name.
		_, rest, found := strings.Cut(tags[i].Value, ",")
		if found {
			var opt string
			for len(rest) > 0 {
				opt, rest, _ = strings.Cut(rest, ",")
				if opt == "omitempty" {
					hasOmitempty = true
					break
				}
			}
		}

		if !hasOmitempty {
			tags[i].Value += ",omitempty"
		}
		break
	}

	if !hasTag {
		tags = append(tags, Tag{Key: "json", Value: ",omitempty"})
	}

	return FlattenStructTag(tags)
}
