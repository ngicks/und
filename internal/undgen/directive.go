package undgen

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"strings"
)

const (
	DirectivePrefix = "undgen:"
)

const (
	DirectiveCommentIgnore    = "ignore"
	DirectiveCommentGenerated = "generated"
)

type Directive struct {
	ignore    bool
	generated bool
	json      map[string]any
}

func ParseComment(cg *ast.CommentGroup) (directive Directive, has bool, err error) {
	if cg == nil {
		return Directive{}, false, nil
	}

	for i, c := range cg.List {
		text := stripMarker(c.Text)
		text = strings.TrimSpace(text)
		if !strings.HasPrefix(text, DirectivePrefix) {
			continue
		}
		text = text[len(DirectivePrefix):]
		if len(text) == 0 {
			return Directive{}, true, fmt.Errorf("empty")
		}
		switch {
		case strings.HasPrefix(text, DirectiveCommentIgnore):
			return Directive{ignore: true}, true, nil
		case strings.HasPrefix(text, DirectiveCommentGenerated):
			return Directive{generated: true}, true, nil
		}
		if text[0] != '{' {
			return Directive{}, true, fmt.Errorf("must be a JSON object or specific text without any enclosure: malformed: %s", text)
		}
		var rest strings.Builder
		rest.WriteString(text)
		if i+1 < len(cg.List) {
			for _, c := range cg.List[i+1:] {
				rest.WriteString(stripMarker(c.Text))
			}
		}
		var m map[string]any
		err = json.NewDecoder(strings.NewReader(rest.String())).Decode(&m)
		if err != nil {
			return Directive{}, true, fmt.Errorf("malformed: %w", err)
		}
		return Directive{json: m}, true, nil
	}

	return Directive{}, false, nil
}

func stripMarker(text string) string {
	if len(text) < 2 {
		return text
	}
	switch text[1] {
	case '/':
		return text[2:]
	case '*':
		return text[2 : len(text)-2]
	}
	return text
}

func (d Directive) MustIgnore() bool {
	return d.generated || d.ignore
}
