package undgen

import (
	"go/ast"
	"go/token"
	"path"
	"reflect"
	"strconv"

	"github.com/ngicks/und/internal/structtag"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type GeneratedPlainType struct {
	Pkg map[string]GeneratedTypeBuf
}

type GeneratedTypeBuf struct {
	PkgName   string
	Imports   map[string]string // maps path to identifier.
	Generated []SoleDecl
}

type SoleDecl struct {
	Fset *token.FileSet
	// TypeSpec only. The TypeSpec does not print `type` keyword its spec. So keep it as GenDecl.
	Decl *ast.GenDecl
}

func TargetTypes(pkgs []*packages.Package) (map[string]map[string]bool, error) {
	targetTypeNames := map[string]map[string]bool{}

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
						if !hasUndType(ts, imports) {
							continue
						}

						if targetTypeNames[pkg.PkgPath] == nil {
							targetTypeNames[pkg.PkgPath] = map[string]bool{}
						}
						targetTypeNames[pkg.PkgPath][ts.Name.Name] = true
					}
				}
			}
		}
	}

	return targetTypeNames, nil
}

func hasUndType(ts *ast.TypeSpec, imports UndImports) bool {
	st, ok := ts.Type.(*ast.StructType)
	if !ok {
		return false
	}
	if st.Fields == nil {
		return false
	}
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
			return true
		}
		// TODO: check it is struct type and contains und types.
		// If it is defined under code generation target package and contains und types,
		// It should have corresponding plain type.
	}
	return false
}

func GeneratePlainType(pkgs []*packages.Package) (GeneratedPlainType, error) {
	targets, err := TargetTypes(pkgs)
	if err != nil {
		return GeneratedPlainType{}, err
	}

	var generated GeneratedPlainType
	for _, pkg := range pkgs {
		if _, ok := targets[pkg.PkgPath]; !ok {
			continue
		}
		for _, f := range pkg.Syntax {
			imports, ok := parseImports(f.Imports)
			if !ok {
				continue
			}

			// clone before modification.
			f, fset := clone(pkg.Fset, f)
		DECL:
			for _, decl := range f.Decls {
				switch x := decl.(type) {
				default:
					continue DECL
				case *ast.GenDecl:
					if x.Tok != token.TYPE {
						continue
					}
					for _, s := range x.Specs {
						ts := s.(*ast.TypeSpec)
						if ts.Name == nil {
							continue
						}
						if !targets[pkg.PkgPath][ts.Name.Name] {
							continue
						}
						modifiedAny := false
						astutil.Apply(ts, func(c *astutil.Cursor) bool {
							node := c.Node()
							switch x := node.(type) {
							default:
								return true
							case *ast.TypeSpec:
								if x.Name != nil {
									x.Name.Name = x.Name.Name + "Plain"
									c.Replace(x)
								}
								return true
							case *ast.Field:
								typ, ok := x.Type.(*ast.IndexExpr)
								if !ok {
									// no traversal for now.
									return false
								}
								// It's not possible to placing und types without selection since it's outer type.
								// Do not alias.
								expr, ok := typ.X.(*ast.SelectorExpr)
								if !ok {
									return false
								}
								left, ok := expr.X.(*ast.Ident)
								if !ok || left == nil {
									return false
								}
								right := expr.Sel
								if !imports.Has(left.Name, right.Name) {
									// TODO: check it is struct type and contains und types.
									// If it is defined under code generation target package and contains und types,
									// It should have corresponding plain type.
									return false
								}

								tag := x.Tag.Value
								if tag == "" {
									return false
								}

								tag = reflect.StructTag(tag[1 : len(tag)-1]).Get(structtag.UndTag)
								if tag == "" {
									return false
								}

								undOpt, err := structtag.ParseOption(tag)
								if err != nil {
									panic(err)
								}

								modified := true
								imports.Matcher(left.Name, right.Name).Match(
									func() {
										switch s := undOpt.States.Value(); {
										case s.Def && (s.Null || s.Und):
											modified = false
											return
										case s.Def:
											x.Type = typ.Index
											c.Replace(x)
										case s.Null || s.Und:
											x.Type = &ast.StarExpr{ // *struct{}
												X: &ast.StructType{
													Fields: &ast.FieldList{Opening: 1, Closing: 2},
												},
											}
											c.Replace(x)
										}
									},
									func(_ bool) {
										switch s := undOpt.States.Value(); {
										case s.Def && s.Null && s.Und:
											modified = false
											return
										case s.Def && (s.Null || s.Und):
											expr.X = &ast.Ident{
												Name: imports.option,
											}
											expr.Sel = &ast.Ident{
												Name: "Option",
											}
											c.Replace(x)
										case s.Null && s.Und:
											typ.Index = &ast.StarExpr{ // *struct{}
												X: &ast.StructType{
													Fields: &ast.FieldList{Opening: 1, Closing: 2},
												},
											}
											expr.X = &ast.Ident{
												Name: imports.option,
											}
											expr.Sel = &ast.Ident{
												Name: "Option",
											}
											c.Replace(x)
										case s.Def:
											x.Type = typ.Index
											c.Replace(x)
										case s.Null || s.Und:
											x.Type = &ast.StarExpr{ // *struct{}
												X: &ast.StructType{
													Fields: &ast.FieldList{Opening: 1, Closing: 2},
												},
											}
											c.Replace(x)
										}
									},
									func(isSlice bool) {
										if (undOpt.States.IsSomeAnd(func(s structtag.States) bool {
											return s.Def && s.Null && s.Und
										})) && (undOpt.Len.IsNone() || undOpt.Len.IsSomeAnd(func(lv structtag.LenValidator) bool {
											return lv.Op != structtag.LenOpEqEq
										})) && (undOpt.Values.IsNone()) {
											modified = false
											return
										}

										// Generally for other cases, replace types
										typ.X = &ast.SelectorExpr{
											X: &ast.Ident{
												Name: func() string {
													if isSlice {
														return imports.sliceUnd
													} else {
														return imports.und
													}
												}(),
											},
											Sel: &ast.Ident{
												Name: "Und",
											},
										}
										typ.Index = &ast.ArrayType{
											Elt: &ast.IndexExpr{
												X: &ast.SelectorExpr{
													X: &ast.Ident{
														Name: imports.option,
													},
													Sel: &ast.Ident{
														Name: "Option",
													},
												},
												Index: typ.Index,
											},
										}

										if undOpt.Len.IsSome() {
											lv := undOpt.Len.Value()
											if lv.Op == structtag.LenOpEqEq {
												if lv.Len == 1 {
													typ.Index = typ.Index.(*ast.ArrayType).Elt
												} else {
													typ.Index.(*ast.ArrayType).Len = &ast.BasicLit{
														Kind:  token.INT,
														Value: strconv.FormatInt(int64(undOpt.Len.Value().Len), 10),
													}
												}
											}
										}
										if undOpt.Values.IsSome() {
											switch x := undOpt.Values.Value(); {
											case x.Nonnull:
												switch x := typ.Index.(type) {
												case *ast.ArrayType:
													x.Elt = x.Elt.(*ast.IndexExpr).Index
												case *ast.IndexExpr:
													typ.Index = x.Index
												default:
													panic("implementation error")
												}
											}
										}

										states := undOpt.States.Value()
										if undOpt.Len.IsSome() {
											states.Def = true
										}

										switch s := states; {
										default:
										case s.Def && (s.Null || s.Und):
											typ.X.(*ast.SelectorExpr).X = &ast.Ident{
												Name: imports.option,
											}
											typ.X.(*ast.SelectorExpr).Sel = &ast.Ident{
												Name: "Option",
											}
											c.Replace(x)
										case s.Null && s.Und:
											typ.Index = &ast.StarExpr{ // *struct{}
												X: &ast.StructType{
													Fields: &ast.FieldList{Opening: 1, Closing: 2},
												},
											}
											typ.X.(*ast.SelectorExpr).X = &ast.Ident{
												Name: imports.option,
											}
											typ.X.(*ast.SelectorExpr).Sel = &ast.Ident{
												Name: "Option",
											}
											c.Replace(x)
											return
										case s.Def:
											x.Type = typ.Index
											c.Replace(x)
										case s.Null || s.Und:
											x.Type = &ast.StarExpr{ // *struct{}
												X: &ast.StructType{
													Fields: &ast.FieldList{Opening: 1, Closing: 2},
												},
											}
											c.Replace(x)
										}
									},
								)
								if modified {
									modifiedAny = true
								}
								return false
							}
						}, nil)

						if modifiedAny {
							if generated.Pkg == nil {
								generated.Pkg = map[string]GeneratedTypeBuf{}
							}
							genBuf := generated.Pkg[pkg.PkgPath]
							genBuf.PkgName = pkg.Name
							if genBuf.Imports == nil {
								genBuf.Imports = map[string]string{}
								for _, i := range f.Imports {
									// strip " or `
									pkgPath := i.Path.Value[1 : len(i.Path.Value)-1]
									name := path.Base(pkgPath)
									if i.Name != nil {
										name = i.Name.Name
									}
									genBuf.Imports[pkgPath] = name
								}
								for pkgPath, name := range imports.Imports() {
									genBuf.Imports[pkgPath] = name
								}
							}

							genBuf.Generated = append(genBuf.Generated, SoleDecl{fset, x})

							generated.Pkg[pkg.PkgPath] = genBuf
						}
					}
				}
			}
		}
	}
	return generated, nil
}
