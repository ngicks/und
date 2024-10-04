package undgen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"path"
	"slices"
	"strconv"
	"strings"
)

type TargetImport struct {
	ImportPath string
	Types      []string
}

type MatchedType struct {
	Name  string
	Field []MatchedField
}

type MatchedField struct {
	Name   string
	Direct bool
	As     MatchedAs
	Type   TargetType
}

type MatchedAs string

const (
	MatchedAsArray  MatchedAs = "array"
	MatchedAsSlice  MatchedAs = "slice"
	MatchedAsMap    MatchedAs = "map"
	MatchedAsStruct MatchedAs = "struct"
)

type TargetType struct {
	ImportPath string
	Name       string
}

// parseImports relates ident (PackageName) to TargetImport.
func parseImports(importSpecs []*ast.ImportSpec, imports []TargetImport) map[string]TargetImport {
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

func FindMatchedTypes(file *ast.File, imports []TargetImport) ([]MatchedType, error) {
	importMap := parseImports(file.Imports, imports)
	if len(importMap) == 0 {
		return nil
	}
	for _, dec := range file.Decls {
		genDecl, ok := dec.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.TYPE {
			continue
		}

		direction, _, err := ParseUndComment(genDecl.Doc)
		if err != nil {
			return nil, fmt.Errorf("comment for %s: %w", genDecl.Name)
		}

	}

	// first path, collect generation targets
	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
			imports, ok := parseImports(f.Imports)
			if !ok {
				continue
			}

			for _, decl := range f.Decls {
				switch x := decl.(type) {
				default:
					continue
				case *ast.GenDecl:
					if x.Tok != token.TYPE {
						continue
					}
					dec, found, err := ParseComment(x.Doc)
					if err != nil {
						return nil, err
					}
					if found && dec.MustIgnore() {
						continue
					}
					for _, s := range x.Specs {
						ts := s.(*ast.TypeSpec)
						dec, found, err := ParseComment(ts.Comment)
						if err != nil {
							return nil, err
						}
						if found && dec.MustIgnore() {
							continue
						}

						// TODO: recursively check struct types.
						ti, ok := parseUndType(ts, imports)
						if !ok {
							continue
						}

						tt.add(pkg.PkgPath, ti)
					}
				}
			}
		}
	}

	return tt, nil
}

func parseUndType(ts *ast.TypeSpec, imports UndImports) (ti MatchedType, hasUndField bool) {
	st, ok := ts.Type.(*ast.StructType)
	if !ok {
		return MatchedType{}, false
	}
	if st.Fields == nil {
		return MatchedType{}, false
	}

	hasUndField = false
	for _, f := range st.Fields.List {
		typ, ok := f.Type.(*ast.IndexExpr)
		if !ok {
			// no traversal for now.
			continue
		}
		// It's not possible to placing und types without selection since it's outer type.
		// Do not alias.
		expr, ok := typ.X.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		left, ok := expr.X.(*ast.Ident)
		if !ok || left == nil {
			continue
		}
		right := expr.Sel
		if imports.Has(left.Name, right.Name) {
			hasUndField = true
			var buf bytes.Buffer
			fset := token.NewFileSet()
			err := printer.Fprint(&buf, fset, typ.Index)
			if err != nil {
				panic(err)
			}
			for _, n := range f.Names {
				if ti.fields == nil {
					ti.fields = make(map[string]TargetFieldInfo)
				}
				ti.fields[n.Name] = TargetFieldInfo{
					Kind:     imports.Kind(left.Name, right.Name),
					Name:     n.Name,
					TypeParm: buf.String(),
				}
			}
		}
		// TODO: check it is struct type and contains und types.
		// If it is defined under code generation target package and contains und types,
		// It should have corresponding plain type.
	}

	ti.Name = ts.Name.Name

	return ti, hasUndField
}
