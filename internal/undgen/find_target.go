package undgen

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"

	"golang.org/x/tools/go/packages"
)

// TargetTypes holds found target types.
// Types should be one of
//
//   - array/slice of target type or map whose value type contains target type.
//   - struct that contains target type.
//   - struct that contains the struct which implements conversion method or also a target type.
type TargetTypes map[string]map[string]TargetType

func (tt *TargetTypes) lazyInit() {
	if *tt == nil {
		*tt = make(TargetTypes)
	}
}

func (tt *TargetTypes) set(pkg string, tyName string, target TargetType) {
	tt.lazyInit()
	if (*tt)[pkg] == nil {
		(*tt)[pkg] = make(map[string]TargetType)
	}
	(*tt)[pkg][tyName] = target
}

func (tt *TargetTypes) HasPkg(pkg string) bool {
	tt.lazyInit()
	_, ok := (*tt)[pkg]
	return ok
}

// TargetType
type TargetType struct {
	// fields is non-nil if the type is struct.
	fields map[string]TargetFieldInfo
	Name   string
}

type TargetFieldInfo struct {
	Kind     string
	Name     string
	TypeParm string // maybe empty if the ti is remote type.
}

func newTargetTypePackages() *targetTypePackages {
	return &targetTypePackages{
		m: map[string]*targetTypesPackage{},
	}
}

func (pkgs *targetTypePackages) lazyInit() {
	if pkgs.m == nil {
		pkgs.m = make(map[string]*targetTypesPackage)
	}
}

func (pkgs *targetTypePackages) hasPkg(pkgPath string) bool {
	// reading from nil map does not panic.
	_, ok := pkgs.m[pkgPath]
	return ok
}

func (pkgs *targetTypePackages) get(pkgPath string, name string) (TargetType, bool) {
	pkg, ok := pkgs.m[pkgPath]
	if !ok {
		return TargetType{}, false
	}
	mm, ok := pkg.m[name]
	return mm, ok
}

func (pkgs *targetTypePackages) add(pkgPath string, ti TargetType) {
	pkgs.lazyInit()
	pkgs.m[pkgPath].set(ti)
}

func FindTargetTypes(pkgs []*packages.Package) (tt *targetTypePackages, err error) {
	tt = new(targetTypePackages)

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

func parseUndType(ts *ast.TypeSpec, imports UndImports) (ti TargetType, hasUndField bool) {
	st, ok := ts.Type.(*ast.StructType)
	if !ok {
		return TargetType{}, false
	}
	if st.Fields == nil {
		return TargetType{}, false
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
