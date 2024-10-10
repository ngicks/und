package undgen

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
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

type ConversionMethodsSet struct {
	ToRaw   string
	ToPlain string
}

type MatchedResult map[string]MatchedPackage

type MatchedPackage struct {
	Pkg   *packages.Package
	Files map[string]MatchedFile
}

func (mt *MatchedPackage) lazyInit(pkg *packages.Package) {
	if mt.Files == nil {
		mt.Pkg = pkg
		mt.Files = make(map[string]MatchedFile)
	}
}

type MatchedFile struct {
	File  *ast.File
	Types []MatchedType
}

func (mf *MatchedFile) lazyInit(f *ast.File) {
	if mf.Types == nil {
		mf.File = f
		mf.Types = make([]MatchedType, 0)
	}
}

func (mf *MatchedFile) sortCompact() {
	slices.SortFunc(
		mf.Types,
		func(i, j MatchedType) int {
			return cmp.Compare(i.Pos, j.Pos)
		},
	)
	mf.Types = slices.CompactFunc(
		mf.Types,
		func(i, j MatchedType) bool { return i.Name == j.Name },
	)
}

type MatchedType struct {
	// 0-indexed number of appearance within the file. source code order.
	Pos int
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
type importDecls struct {
	identToImport map[string]*TargetImport
	importToIdent map[*TargetImport]string
}

func (id importDecls) Len() int {
	return max(len(id.identToImport), len(id.importToIdent))
}

func (id importDecls) HasSelector(left, right string) bool {
	ti, ok := id.identToImport[left]
	return ok && slices.Contains(ti.Types, right)
}

func (i importDecls) MatchTy(path string, name string) (TargetType, bool) {
	for _, v := range i.identToImport {
		if v.ImportPath == path && slices.Contains(v.Types, name) {
			return TargetType{v.ImportPath, name}, true
		}
	}
	return TargetType{}, false
}

// parseImports relates ident (PackageName) to TargetImport.
func parseImports(importSpecs []*ast.ImportSpec, imports []TargetImport) importDecls {
	// pre-process input
	imports = slices.Collect(
		xiter.Map(
			func(t TargetImport) TargetImport {
				t.Types = slices.Clone(t.Types)
				return t
			},
			slices.Values(imports),
		),
	)

	id := importDecls{
		identToImport: map[string]*TargetImport{},
		importToIdent: map[*TargetImport]string{},
	}

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
		targetImport := imports[idx]
		var ident string
		if is.Name != nil {
			ident = is.Name.Name
		} else {
			pkgBase := path.Base(pkgPath)
			if strings.HasPrefix(pkgBase, "v") && len(strings.TrimFunc(pkgBase[1:], isAsciiNum)) == 0 {
				pkgBase = path.Base(path.Dir(pkgPath))
			}
			ident = pkgBase
		}
		id.identToImport[ident] = &targetImport
		id.importToIdent[&targetImport] = ident
	}
	return id
}

func isAsciiNum(r rune) bool {
	return '0' <= r && r <= '9'
}

func FindMatchedTypes(pkgs []*packages.Package, imports []TargetImport, methods ConversionMethodsSet) (MatchedResult, error) {
	// 1st path, find other than implementor
	matched, err := findMatchedTypes(pkgs, imports, methods, nil)
	if err != nil {
		return matched, err
	}
	// 2nd path, find including implementor
	matched, err = findMatchedTypes(pkgs, imports, methods, matched)
	if err != nil {
		return matched, err
	}

	return matched, nil
}

func findMatchedTypes(pkgs []*packages.Package, imports []TargetImport, methods ConversionMethodsSet, matched MatchedResult) (MatchedResult, error) {
	if matched == nil {
		matched = make(MatchedResult)
	}
	for _, pkg := range pkgs {
		matchedPkg := matched[pkg.PkgPath]
		matchedPkg.lazyInit(pkg)

		for _, file := range pkg.Syntax {
			filename := pkg.Fset.Position(file.FileStart).Filename
			matchedFile := matchedPkg.Files[filename]
			matchedFile.lazyInit(file)

			importMap := parseImports(file.Imports, imports)
			// Do not return early even if importMap.Len() == 0
			// since it could still include implementor.
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

				for j, s := range genDecl.Specs {
					ts := s.(*ast.TypeSpec)
					direction, _, err := ParseUndComment(ts.Doc)
					if err != nil {
						return matched, fmt.Errorf("comment for decl %q", i, ts.Name.Name)
					}

					if direction.MustIgnore() {
						continue
					}

					obj := pkg.TypesInfo.Defs[ts.Name]
					if obj == nil {
						continue
					}
					mt, ok := parseUndType(obj, matched, importMap, methods)
					if !ok {
						continue
					}
					mt.Pos = i + j
					matchedFile.Types = append(matchedFile.Types, mt)
				}
			}

			matchedFile.sortCompact()
			matchedPkg.Files[filename] = matchedFile
		}

		matched[pkg.PkgPath] = matchedPkg
	}

	return matched, nil
}

func parseUndType(
	obj types.Object,
	total MatchedResult,
	imports importDecls,
	conversionMethod ConversionMethodsSet,
) (mt MatchedType, has bool) {
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return
	}
	switch /*underlying :=*/ obj.Type().Underlying().(type) {
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
		direct, matchedAs, matchedTy := isTargetType(f.Type(), imports, 0)
		if !direct && matchedAs == "" {
			continue
		}
		matched = append(
			matched,
			MatchedField{
				Name: named.Obj().Name(),
				As:   matchedAs,
				Type: matchedTy,
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
	return false, "", TargetType{}
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
