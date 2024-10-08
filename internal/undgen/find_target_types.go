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

type TargetImport struct {
	ImportPath string
	Types      []string
}

type MatchedType struct {
	Name    string
	Variant MatchedTypeVariant
	Field   []MatchedField
}

type MatchedTypeVariant string

const (
	MatchedTypeVariantStruct MatchedTypeVariant = "struct"
	MatchedTypeVariantArray  MatchedTypeVariant = "array"
	MatchedTypeVariantSlice  MatchedTypeVariant = "slice"
	MatchedTypeVariantMap    MatchedTypeVariant = "map"
)

type MatchedField struct {
	Name   string
	Direct bool
	As     MatchedAs
	Type   TargetType
}

type MatchedAs string

const (
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
	allTypes, err := findAllTypesInPackages(pkgs)
	if err != nil {
		return nil, err
	}
	for _, pkg := range pkgs {
	}
}

func findAllTypesInPackages(pkgs []*packages.Package) (map[string]map[string]bool, error) {
	allTypesInPackages := map[string]map[string]bool{}
	for _, pkg := range pkgs {
		for i, f := range pkg.Syntax {
			for _, dec := range f.Decls {
				genDecl, ok := dec.(*ast.GenDecl)
				if !ok {
					continue
				}
				if genDecl.Tok != token.TYPE {
					continue
				}

				direction, _, err := ParseUndComment(genDecl.Doc)
				if err != nil {
					return allTypesInPackages, fmt.Errorf("comment for decl %s at %d: %w", f.Name, i, err)
				}

				if direction.MustIgnore() {
					continue
				}

				for _, s := range genDecl.Specs {
					ts := s.(*ast.TypeSpec)
					direction, _, err := ParseUndComment(ts.Doc)
					if err != nil {
						return allTypesInPackages, fmt.Errorf("comment for decl %q", ts.Name.Name)
					}

					if direction.MustIgnore() {
						continue
					}

					if allTypesInPackages[pkg.PkgPath] == nil {
						allTypesInPackages[pkg.PkgPath] = map[string]bool{}
					}
					allTypesInPackages[pkg.PkgPath][ts.Name.String()] = true
				}
			}
		}
	}
	return allTypesInPackages, nil
}

func findMatchedTypesFile(
	file *ast.File,
	typeInfo *types.Info,
	allTargetTypes map[string]map[string]bool,
	imports []TargetImport,
) ([]MatchedType, error) {
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
