package testcase_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
	"gotest.tools/v3/assert"
)

var (
	_ slog.LogValuer = option.Option[any]{}
	_ slog.LogValuer = und.Und[any]{}
	_ slog.LogValuer = sliceund.Und[any]{}
	_ slog.LogValuer = elastic.Elastic[any]{}
	_ slog.LogValuer = sliceelastic.Elastic[any]{}
)

func TestSlogValuer(t *testing.T) {
	var buf bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	logger.Info("opt", "none", option.None[string](), "some", option.Some("foo"))
	logger.Info("und", "undefined", und.Undefined[string](), "null", und.Null[string](), "defined", und.Defined("foo"))
	logger.Info("sliceund", "undefined", sliceund.Undefined[string](), "null", sliceund.Null[string](), "defined", sliceund.Defined("foo"))
	logger.Info(
		"elastic",
		"undefined", elastic.Undefined[string](),
		"null", elastic.Null[string](),
		"defined", elastic.FromValue("foo"),
		"multiple", elastic.FromPointers([]*string{nil, ptr("bar")}),
	)
	logger.Info(
		"elastic",
		"undefined", sliceelastic.Undefined[string](),
		"null", sliceelastic.Null[string](),
		"defined", sliceelastic.FromValue("foo"),
		"multiple", sliceelastic.FromPointers([]*string{nil, ptr("bar")}),
	)

	expected := []map[string]any{
		{
			"none": nil,
			"some": "foo",
		},
		{
			"undefined": nil,
			"null":      nil,
			"defined":   "foo",
		},
		{
			"undefined": nil,
			"null":      nil,
			"defined":   "foo",
		},
		{
			"undefined": nil,
			"null":      nil,
			"defined":   []any{"foo"},
			"multiple":  []any{nil, "bar"},
		},
		{
			"undefined": nil,
			"null":      nil,
			"defined":   []any{"foo"},
			"multiple":  []any{nil, "bar"},
		},
	}

	for i, line := range strings.Split(buf.String(), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		err := json.Unmarshal([]byte(line), &m)
		assert.NilError(t, err)
		for k, v := range expected[i] {
			assert.DeepEqual(t, m[k], v)
		}
	}
}
