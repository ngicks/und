package undgen

import (
	"go/ast"
	"path"
	"strconv"
)

type UndImports struct {
	option       string
	und          string
	elastic      string
	sliceUnd     string
	sliceElastic string
	conversion   string
}

func parseImports(specs []*ast.ImportSpec) (nameMap UndImports, ok bool) {
	idents := make(map[string]bool)

	for _, s := range specs {
		// strip " or `
		pkgPath := s.Path.Value[1 : len(s.Path.Value)-1]
		switch pkgPath {
		case "github.com/ngicks/und/option", "github.com/ngicks/und",
			"github.com/ngicks/und/elastic", "github.com/ngicks/und/sliceund",
			"github.com/ngicks/und/sliceund/elastic",
			"github.com/ngicks/und/conversion":
			ok = true
		}
		var f *string
		switch pkgPath {
		case "github.com/ngicks/und/option":
			f = &nameMap.option
		case "github.com/ngicks/und":
			f = &nameMap.und
		case "github.com/ngicks/und/elastic":
			f = &nameMap.elastic
		case "github.com/ngicks/und/sliceund":
			f = &nameMap.sliceUnd
		case "github.com/ngicks/und/sliceund/elastic":
			f = &nameMap.sliceElastic
		case "github.com/ngicks/und/conversion":
			f = &nameMap.conversion
		}
		if s.Name != nil {
			*f = s.Name.Name
			idents[s.Name.Name] = true
		} else {
			defaultIdent := path.Base(pkgPath)
			*f = defaultIdent
			idents[defaultIdent] = true
		}
	}

	nameMap = nameMap.fill(idents)

	return
}

func (i UndImports) fill(idents map[string]bool) UndImports {
	setFallingBack := func(tgt *string, name string) {
		if *tgt != "" {
			return
		}
		if !idents[name] {
			*tgt = name
			return
		}
		for i := 0; ; i++ {
			fallenBack := name + "_" + strconv.FormatInt(int64(i), 10)
			if !idents[fallenBack] {
				*tgt = fallenBack
				return
			}
		}
	}
	setFallingBack(&i.option, "option")
	setFallingBack(&i.und, "und")
	setFallingBack(&i.elastic, "elastic")
	setFallingBack(&i.sliceUnd, "sliceund")
	setFallingBack(&i.sliceElastic, "sliceelastic")
	setFallingBack(&i.conversion, "conversion")
	return i
}

func (i UndImports) Has(x string, sel string) bool {
	return i.Kind(x, sel) > 0
}

type UndTypeKind int

const (
	TypeKindOption UndTypeKind = iota + 1
	TypeKindUnd
	TypeKindSliceUnd
	TypeKindElastic
	TypeKindSliceElastic
)

func (i UndImports) Kind(x string, sel string) UndTypeKind {
	switch x { // conversion does not have type.
	case i.option:
		if sel == "Option" {
			return TypeKindOption
		}
	case i.und:
		if sel == "Und" {
			return TypeKindUnd
		}
	case i.sliceUnd:
		if sel == "Und" {
			return TypeKindSliceUnd
		}
	case i.elastic:
		if sel == "Elastic" {
			return TypeKindElastic
		}
	case i.sliceElastic:
		if sel == "Elastic" {
			return TypeKindSliceElastic
		}
	}
	return 0
}

func (i UndImports) Matcher(x string, sel string) matcher {
	return matcher{i, x, sel}
}

type matcher struct {
	names  UndImports
	x, sel string
}

func (m matcher) Match(onOption func(), onUnd func(isSlice bool), onElastic func(isSlice bool)) {
	switch m.x {
	case m.names.option:
		if m.sel == "Option" {
			onOption()
		}
	case m.names.und, m.names.sliceUnd:
		if m.sel == "Und" {
			onUnd(m.x == m.names.sliceUnd)
		}
	case m.names.elastic, m.names.sliceElastic:
		if m.sel == "Elastic" {
			onElastic(m.x == m.names.sliceElastic)
		}
	}
}

func (i UndImports) Imports() map[string]string {
	return map[string]string{
		"github.com/ngicks/und/option":           i.option,
		"github.com/ngicks/und":                  i.und,
		"github.com/ngicks/und/elastic":          i.elastic,
		"github.com/ngicks/und/sliceund":         i.sliceUnd,
		"github.com/ngicks/und/sliceund/elastic": i.sliceElastic,
		"github.com/ngicks/und/conversion":       i.conversion,
	}
}

func (i UndImports) Und(isSlice bool) string {
	if isSlice {
		return i.sliceUnd
	} else {
		return i.und
	}
}

func (i UndImports) Elastic(isSlice bool) string {
	if isSlice {
		return i.sliceElastic
	} else {
		return i.elastic
	}
}
