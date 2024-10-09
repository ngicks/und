package undgen

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Finding target types.
// Targets must be defined in the packages located under the cwd.
// However it is allowed that evaluating symlinks and as a result traversing upward.
// (Basically do not do that.)
//
// Take []*(golang.org/x/tools/go/packages).Package as an argument.
// Both ast and type info are used.
// Modules other than generation target must be also evaluated and type-checked,
// because they might be implementors of UndRaw/UndPlain thus structs containing them in
// descendant fields are also target types.
//
// All found types are marked as target if they are;
//  - defined map/slice/array type where value type is target type. The key type is ignored since there's no point converting it.
//  - defined struct type where at least one field is target type
//  - Implementor of conversion method set (UndRaw, UndPlain for und types), where both are mutually convertible.

type TargetImport struct {
	ImportPath string
	Types      []string
}

type MatchedType struct {
	// Name of type without type params.
	// Just here for later reuse to look up ast.
	Name string
	// this must not be MatchedAsImplementor
	Variant MatchedAs
	// If Variants is other than "struct", this fields is empty.
	Field []MatchedField
}

type MatchedField struct {
	Name string
	As   MatchedAs
	Type TargetType
}

type MatchedAs string

const (
	MatchedAsStruct      MatchedAs = "struct"
	MatchedAsArray       MatchedAs = "array"
	MatchedAsSlice       MatchedAs = "slice"
	MatchedAsMap         MatchedAs = "map"
	MatchedAsImplementor MatchedAs = "implementor"
)

type TargetType struct {
	ImportPath string
	Name       string
}

// importDecls maps idents (PackageNames) to TargetImport
type importDecls map[string]TargetImport

func (i importDecls) HasSelector(left, right string) bool {
	ti, ok := i[left]
	return ok && slices.Contains(ti.Types, right)
}

func (i importDecls) MatchTy(path string, name string) (TargetType, bool) {
	for _, v := range i {
		if v.ImportPath == path && slices.Contains(v.Types, name) {
			return TargetType{v.ImportPath, name}, true
		}
	}
	return TargetType{}, false
}

// parseImports relates ident (PackageName) to TargetImport.
func parseImports(importSpecs []*ast.ImportSpec, imports []TargetImport) importDecls {
	var m map[string]TargetImport
	for _, is := range importSpecs {
		var pkgPath string
		// strip " or `
		if is.Path.Value[0] == '"' {
			var err error
			pkgPath, err = strconv.Unquote(is.Path.Value)
			if err != nil {
				panic(fmt.Errorf("malformed import: %w", err))
			}
		} else {
			pkgPath = is.Path.Value[1 : len(is.Path.Value)-1]
		}
		idx := slices.IndexFunc(imports, func(i TargetImport) bool { return pkgPath == i.ImportPath })
		if idx < 0 {
			continue
		}
		if m == nil {
			m = make(map[string]TargetImport)
		}
		if is.Name != nil {
			m[is.Name.Name] = imports[idx]
		} else {
			pkgBase := path.Base(pkgPath)
			if strings.HasPrefix(pkgBase, "v") && len(strings.TrimFunc(pkgBase[1:], isAsciiNum)) == 0 {
				pkgBase = path.Base(path.Dir(pkgPath))
			}
			m[pkgBase] = imports[idx]
		}
	}
	return m
}

func isAsciiNum(r rune) bool {
	return '0' <= r && r <= '9'
}

func FindMatchedTypes(pkgs []*packages.Package, imports []TargetImport) ([]MatchedType, error) {
	for _, pkg := range pkgs {
		for _, f :=range 
		var matched []MatchedType
		importMap := parseImports(file.Imports, imports)
		if len(importMap) == 0 {
			return nil, nil
		}
		for i, dec := range file.Decls {
			genDecl, ok := dec.(*ast.GenDecl)
			if !ok {
				continue
			}
			if genDecl.Tok != token.TYPE {
				continue
			}

			direction, _, err := ParseUndComment(genDecl.Doc)
			if err != nil {
				return matched, fmt.Errorf("comment for decl at %d: %w", i, file.Name, err)
			}

			if direction.MustIgnore() {
				continue
			}

			for _, s := range genDecl.Specs {
				ts := s.(*ast.TypeSpec)
				direction, _, err := ParseUndComment(ts.Doc)
				if err != nil {
					return matched, fmt.Errorf("comment for decl %q", i, ts.Name.Name)
				}

				if direction.MustIgnore() {
					continue
				}

				obj := typeInfo.Defs[ts.Name]
				if obj == nil {
					continue
				}
				mt, ok := parseUndType(obj, allTargetTypes, importMap)
				if !ok {
					continue
				}

				matched = append(matched, mt)
			}
		}
		return matched, nil
	}
}

func parseUndType(
	obj types.Object,
	allTargetTypes map[string]map[string]bool,
	imports importDecls,
) (mt MatchedType, has bool) {
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return
	}
	switch underlying := obj.Type().Underlying().(type) {
	case *types.Struct:
	case *types.Array:
	case *types.Slice:
	case *types.Map:
	}
	structTy, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		return
	}

	has = false
	var matched []MatchedField
	for i := range structTy.NumFields() {
		f := structTy.Field(i)
		direct, matchedAs, matchedTy := isTargetType(f.Type(), imports)
		if !direct && matchedAs == "" {
			continue
		}
		matched = append(
			matched,
			MatchedField{
				Name:   named.Obj().Name(),
				Direct: direct,
				As:     matchedAs,
				Type:   matchedTy,
			},
		)
	}
	return MatchedType{Name: named.Obj().Name(), Field: matched}, len(matched) > 0
}

func isTargetType(ty types.Type, imports importDecls, depth int) (direct bool, matchedAs MatchedAs, matchedTy TargetType) {
	switch x := ty.(type) {
	case *types.Named:
		if matched, ok := imports.MatchTy(x.Obj().Pkg().Path(), x.Obj().Name()); ok {
			return true, "", matched
		}
		_, matchedAs, matchedTy = isTargetType(x.Underlying(), imports, depth-1)
		return false, matchedAs, matchedTy
	case *types.Struct:
		for i := range x.NumFields() {
			_, matchedAs, matchedTy = isTargetType(x.Field(i).Type(), imports, depth-1)
		}
	}
	return false
}

func isRawImplementor(ty types.Type, imports importDecls) bool {
	ms := types.NewMethodSet(ty)
	for i := range ms.Len() {
		sel := ms.At(i)
		if sel.Obj().Name() == "UndRaw" {
			return true
		}
	}
	return false
}
