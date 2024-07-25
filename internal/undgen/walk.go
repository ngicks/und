package undgen

import (
	"fmt"
	"go/ast"
	"go/token"
	"path"
	"slices"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type GeneratedPlainType struct {
	Pkg map[string][]GeneratedTypeBuf
}

type GeneratedTypeBuf struct {
	PkgName   string
	FileName  string
	Imports   map[string]string // maps path to identifier.
	Generated []PlainType
}

type PlainType struct {
	Fset *token.FileSet
	// TypeSpec only. The TypeSpec does not print `type` keyword before its spec. So keep it as GenDecl.
	Decl    *ast.GenDecl
	ToPlain []Conversion
}

type Conversion struct {
	FieldName string
	Converter fieldConverter
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
			clonedF, fset := clone(pkg.Fset, f)
		DECL:
			for _, decl := range clonedF.Decls {
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
						var (
							modifyErr  error
							converters []Conversion
						)
						astutil.Apply(ts.Type, func(c *astutil.Cursor) bool {
							node := c.Node()
							switch field := node.(type) {
							default:
								return true
							case *ast.Field:
								// same error
								converter, _ := undPlainFieldConverter(field, imports)
								modified, err := checkAndModifyUndField(field, imports)
								if err != nil {
									modifyErr = err
									return false
								}
								if modified {
									modifiedAny = true
									converters = append(converters, Conversion{
										FieldName: field.Names[0].Name,
										Converter: converter,
									})
									return false
								} else {
									//TODO check field is other struct and implements ToPlain
									converters = append(converters, Conversion{
										FieldName: field.Names[0].Name,
										Converter: nil,
									})
								}
								// TODO: check if field is another struct,
								// if it is generate target, then modify the field to the one suffixed with `Plain`.
								// Or if it implements ToPlain type, then modify the field to the return value of ToPlain.
								// Further more, we can expand `undgen:` directive comment format to use other `ToPlain` method name.
								return false
							}
						}, nil)

						if modifyErr != nil {
							return generated, err
						}

						if !modifiedAny {
							// maybe this is temporal decision,
							// But here we call this as a malformed und tag.
							return generated, fmt.Errorf("malformed: no field is modified, %s in %s", ts.Name.Name, pkg.PkgPath)
						}

						suffixTypeName(ts, "Plain")

						if generated.Pkg == nil {
							generated.Pkg = map[string][]GeneratedTypeBuf{}
						}

						genBufs := generated.Pkg[pkg.PkgPath]

						var genBuf GeneratedTypeBuf
						genBuf.FileName = pkg.Fset.Position(f.FileStart).Filename
						idx := slices.IndexFunc(genBufs, func(b GeneratedTypeBuf) bool {
							return b.FileName == genBuf.FileName
						})
						if idx >= 0 {
							genBuf = genBufs[idx]
						}

						if genBuf.PkgName == "" {
							genBuf.PkgName = pkg.Name
						}

						if genBuf.Imports == nil {
							genBuf.Imports = map[string]string{}
							for _, i := range clonedF.Imports {
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

						genDecl := *x
						genDecl.Specs = []ast.Spec{ts}
						genBuf.Generated = append(genBuf.Generated, PlainType{fset, &genDecl, converters})

						if idx >= 0 {
							generated.Pkg[pkg.PkgPath][idx] = genBuf
						} else {
							generated.Pkg[pkg.PkgPath] = append(generated.Pkg[pkg.PkgPath], genBuf)
						}
					}
				}
			}
		}
	}
	return generated, nil
}

// suffixTypeName traverse ast tree starting from node
// until first found *ast.TypeSpec.
// It suffixes a type name with suffix.
func suffixTypeName(node ast.Node, suffix string) {
	astutil.Apply(
		node,
		func(c *astutil.Cursor) bool {
			node := c.Node()
			switch x := node.(type) {
			default:
				return true
			case *ast.TypeSpec:
				x.Name.Name = x.Name.Name + suffix
				c.Replace(x)
				return false
			}
		},
		nil,
	)
}
