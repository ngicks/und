package undgen

import (
	"fmt"
	"go/ast"
	"strings"
	"unicode"
)

const (
	UndDirectivePrefix = "undgen:"
)

const (
	UndDirectiveCommentIgnore    = "ignore"
	UndDirectiveCommentGenerated = "generated"
)

type UndDirection struct {
	ignore    bool
	generated bool
}

func (d UndDirection) MustIgnore() bool {
	return d.ignore || d.generated
}

func ParseUndComment(comments *ast.CommentGroup) (UndDirection, bool, error) {
	direction := directiveComments(comments, UndDirectivePrefix, true)

	var ud UndDirection
	if len(direction) == 0 {
		return ud, false, nil
	}

	switch direction[0] {
	default:
		return ud, true, fmt.Errorf("unknown: %v", direction)
	case UndDirectiveCommentIgnore:
		ud.ignore = true
	case UndDirectiveCommentGenerated:
		ud.generated = true
	}

	return ud, true, nil
}

func directiveComments(cg *ast.CommentGroup, directiveMarker string, allowNonDirective bool) []string {
	var stripped []string
	for _, c := range cg.List {
		text := stripMarker(c.Text)
		if allowNonDirective {
			text = strings.TrimLeftFunc(text, unicode.IsSpace)
		}
		var ok bool
		text, ok = strings.CutPrefix(text, directiveMarker)
		if !ok {
			if len(stripped) > 0 {
				break
			} else {
				continue
			}
		}
		stripped = append(stripped, text)
	}
	return stripped
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
