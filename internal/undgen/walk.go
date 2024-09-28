package undgen

import (
	"fmt"
	"go/ast"
	"go/token"
	"path"
	"slices"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
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
	Decl            *ast.GenDecl
	FieldConverters []Conversion
}

type Conversion struct {
	FieldName     string
	Converter     fieldConverter
	BackConverter fieldConverter
}

func GeneratePlainType(pkgs []*packages.Package) (GeneratedPlainType, error) {
	targets, err := TargetTypes(pkgs)
	if err != nil {
		return GeneratedPlainType{}, err
	}

	var generated GeneratedPlainType
	for _, pkg := range pkgs {
		if !targets.hasPkg(pkg.PkgPath) {
			continue
		}
		for _, f := range pkg.Syntax {
			imports, ok := parseImports(f.Imports)
			if !ok {
				continue
			}

			// clone before modification.
			clonedF, fset := clone(pkg.Fset, f)

			df, err := decorator.DecorateFile(fset, clonedF)
			if err != nil {
				return GeneratedPlainType{}, err
			}
		DECL:
			for _, decl := range df.Decls {
				switch x := decl.(type) {
				default:
					continue DECL
				case *dst.GenDecl:
					if x.Tok != token.TYPE {
						continue
					}
					for i, s := range x.Specs {
						ts := s.(*dst.TypeSpec)
						if ts.Name == nil {
							continue
						}
						ti, ok := targets.get(pkg.PkgPath, ts.Name.Name)
						if !ok {
							continue
						}
						modifiedAny := false
						var (
							modifyErr  error
							converters []Conversion
						)
						dstutil.Apply(
							ts.Type,
							func(c *dstutil.Cursor) bool {
								node := c.Node()
								switch field := node.(type) {
								default:
									return true
								case *dst.Field:
									// same error
									converter, _ := undPlainFieldConverter(field, imports, ti.fields[field.Names[0].Name])
									backConverter, _ := undRawFieldBackConverter(field, imports, ti.fields[field.Names[0].Name])
									modified, err := checkAndModifyUndField(field, imports)
									if err != nil {
										modifyErr = err
										return false
									}
									if modified {
										modifiedAny = true
										converters = append(converters, Conversion{
											FieldName:     field.Names[0].Name,
											Converter:     converter,
											BackConverter: backConverter,
										})
										return false
									} else {
										//TODO check field is other struct and implements ToPlain
										converters = append(converters, Conversion{
											FieldName:     field.Names[0].Name,
											Converter:     nil,
											BackConverter: nil,
										})
									}
									// TODO: check if field is another struct,
									// if it is generate target, then modify the field to the one suffixed with `Plain`.
									// Or if it implements ToPlain type, then modify the field to the return value of ToPlain.
									// Further more, we can expand `undgen:` directive comment format to use other `ToPlain` method name.
									return false
								}
							},
							nil,
						)

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

						restorer := decorator.NewRestorer()
						_, err := restorer.FileRestorer().RestoreFile(df)
						if err != nil {
							panic(err)
						}

						genDecl := restorer.Ast.Nodes[x].(*ast.GenDecl)
						genDecl.Specs = []ast.Spec{genDecl.Specs[i]}
						genBuf.Generated = append(genBuf.Generated, PlainType{restorer.Fset, genDecl, converters})
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
func suffixTypeName(node dst.Node, suffix string) {
	dstutil.Apply(
		node,
		func(c *dstutil.Cursor) bool {
			node := c.Node()
			switch x := node.(type) {
			default:
				return true
			case *dst.TypeSpec:
				x.Name.Name = x.Name.Name + suffix
				c.Replace(x)
				return false
			}
		},
		nil,
	)
}
