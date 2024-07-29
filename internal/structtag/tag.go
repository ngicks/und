package structtag

// This file uses modified Go programming language standard library.
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
	"io"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	ErrUnpairedKey = errors.New("unpaired key")
	ErrNotFound    = errors.New("not found")
)

type Tag struct {
	Key   string
	Value string
}

func (t Tag) Flatten() string {
	return t.Key + ":" + strconv.Quote(t.Value)
}

func ParseStructTag(tag reflect.StructTag) (Tags, error) {
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

func StructTagOf(tags []Tag) reflect.StructTag {
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

type Tags []Tag

func (t Tags) mapper(tagName string, option string, fn func(v string, n, m int, err error) (string, error)) (Tags, error) {
	tt := slices.Clone(t)

	for i, tag := range tt {
		if tag.Key != tagName {
			continue
		}
		n, m, _, err := getRange(tag.Value, option)
		tag.Value, err = fn(tag.Value, n, m, err)
		if err != nil {
			return tt, err
		}
		tt[i] = tag
		return tt, nil
	}

	return tt, ErrNotFound
}

func (t Tags) DeleteOption(tagName string, option string) (Tags, error) {
	return t.mapper(tagName, option, func(v string, n, m int, err error) (string, error) {
		if err != nil {
			return v, err
		}

		if option != "" && len(v) > m && v[m] == ',' {
			m++
		}

		v = v[:n] + v[m:]
		// maybe an empty string and a single comma is left behind. cut it.
		v, _ = strings.CutSuffix(v, ",")

		return v, nil
	})
}

func (t Tags) AddOption(tagName string, option, value string) (Tags, error) {
	tt, err := t.mapper(tagName, option, func(v string, n, m int, err error) (string, error) {
		if err == nil {
			return v, nil
		}
		if !errors.Is(err, ErrNotFound) {
			return v, err
		}

		// option does not exist in the tagName:"" section. Add an option.

		if option == "" {
			return value + v, nil
		}

		v = v + "," + option
		if value != "" {
			v += ":" + value
		}

		return v, nil
	})
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return tt, err
		}

		// tagName:"" itself does not exist.
		var newValue string
		if option == "" {
			newValue = value
		} else {
			newValue = "," + option
			if value != "" {
				newValue += ":" + value
			}
		}

		tt = append(tt, Tag{
			Key:   tagName,
			Value: newValue,
		})
	}

	return tt, nil
}

func (t Tags) StructTag() reflect.StructTag {
	return StructTagOf(t)
}

func getRange(tag string, targetOption string) (n, m int, unescaped string, err error) {
	// first, skip name.
	if len(tag) > 0 && !strings.HasPrefix(tag, ",") {
		n := len(tag) - len(strings.TrimLeftFunc(tag, func(r rune) bool {
			return !strings.ContainsRune(",\\'\"`", r) // reserve comma, backslash, and quotes
		}))
		if n == 0 {
			unescaped, n, err = readTagOption(tag)
			if err != nil {
				return -1, -1, "", err
			}
			if targetOption == "" {
				return 0, n, unescaped, nil
			}
		}
		m = n
		tag = tag[n:]
	}

	for len(tag) > 0 {
		n = m + 1
		if tag[0] != ',' {
			return -1, -1, "", ErrNotFound
		} else {
			tag = tag[1:]
			if len(tag) == 0 {
				return -1, -1, "", ErrNotFound
			}
		}

		var (
			opt string
			mm  int
		)
		opt, mm, err = readTagOption(tag)
		if err != nil {
			return -1, -1, "", err
		}

		m = n + mm

		tag = tag[mm:]
		if len(tag) > 0 && tag[0] == ':' {
			tag = tag[len(":"):]
			_, mm, err = readTagOption(tag)
			if err != nil {
				return
			}
			tag = tag[mm:]
			m += len(":") + mm
		}

		if opt == targetOption {
			return n, m, "", nil
		}
	}

	return -1, -1, "", ErrNotFound
}

// AddTagOption returns a new StructTag which has option added for tag.
// It assumes tag options are formatted as `tag:"name,opt,opt"` style.
// The name is allowed to be quoted by single quotation marks.
func AddTagOption(t reflect.StructTag, tag string, option string) (reflect.StructTag, error) {
	tags, err := ParseStructTag(t)
	if err != nil {
		return "", err
	}

	hasTag := false
	for i := 0; i < len(tags); i++ {
		if tags[i].Key != tag {
			continue
		}

		hasTag = true

		hasValue := false

		value := tags[i].Value
		// first, skip name.
		if len(value) > 0 && !strings.HasPrefix(value, ",") {
			n := len(value) - len(strings.TrimLeftFunc(value, func(r rune) bool {
				return !strings.ContainsRune(",\\'\"`", r) // reserve comma, backslash, and quotes
			}))
			if n == 0 {
				_, n, err = readTagOption(value)
				if err != nil {
					return "", err
				}
			}
			value = value[n:]
		}

		for len(value) > 0 {
			if value[0] != ',' {
				return "", fmt.Errorf("malformed option, %s", tags[i].Value)
			} else {
				value = value[1:]
				if len(value) == 0 {
					return "", fmt.Errorf("malformed option, %s", tags[i].Value)
				}
			}

			opt, n, err := readTagOption(value)
			if err != nil {
				return "", err
			}

			value = value[n:]
			if len(value) > 0 && value[0] == ':' {
				if strings.HasPrefix(option, opt+":") {
					hasValue = true
					break
				}
				value = value[len(":"):]
				_, n, err := readTagOption(value)
				if err != nil {
					return "", err
				}
				value = value[n:]
			}

			if option == opt {
				hasValue = true
				break
			}
		}

		if !hasValue {
			if !strings.HasPrefix(option, ",") {
				tags[i].Value += ","
			}
			tags[i].Value += option
		}
		break
	}

	if !hasTag {
		tags = append(tags, Tag{Key: tag, Value: option})
	}

	return StructTagOf(tags), nil
}

func readTagOption(s string) (opt string, n int, err error) {
	if len(s) == 0 {
		return "", 0, io.ErrUnexpectedEOF
	}

	switch r, _ := utf8.DecodeRuneInString(s); {
	case r == '_' || unicode.IsLetter(r): // Go ident
		n = len(s) - len(strings.TrimLeftFunc(s, func(r rune) bool {
			return r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r)
		}))
		return s[:n], n, nil
	case r == '\'': // escaped
		return unescape(s)
	default:
		return "", 0, fmt.Errorf("invalid character: %s", s)
	}
}

func unescape(s string) (unescaped string, n int, err error) {
	i := 0
	if s[0] == '\'' {
		i = 1
	}

	escaping := false
	escaped := []byte{'"'}
	for i < len(s) {
		r, rn := utf8.DecodeRuneInString(s[i:])
		switch {
		case escaping:
			if r == '\'' {
				escaped = escaped[:len(escaped)-1]
			}
			escaping = false
		case r == '\\':
			escaping = true
		case r == '"':
			escaped = append(escaped, '\\')
		case r == '\'':
			escaped = append(escaped, '"')
			i += 1
			out, err := strconv.Unquote(string(escaped))
			if err != nil {
				return "", 0, fmt.Errorf("invalid escaped string: string must be escaped by single quotes, input = %s", s)
			}
			return out, i, nil
		}
		escaped = append(escaped, s[i:][:rn]...)
		i += rn
	}
	return "", 0, fmt.Errorf("invalid escaped string: single-quoted string missing terminating single-quote: %s", s)
}
