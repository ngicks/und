package undgen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"path"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/ngicks/und/internal/structtag"
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
								// TODO: let checkANdModifyUndField also return field converter.
								modified, converter, err := checkAndModifyUndField(field, imports)
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

type fieldType int

const (
	optionFieldType = iota
	undFieldType
	undSliceFieldTye
	elasticFieldType
	elasticSliceFieldType
)

func checkAndModifyUndField(
	f *ast.Field,
	imports UndImports,
) (modified bool, fieldConverter fieldConverter, err error) {
	fieldTy, x, left, right, undOpt, ok, err := isUndField(f, imports)
	if err != nil {
		return false, nil, err
	}
	if !ok {
		return false, nil, nil
	}
	typeParam := fieldTy.Index
	var ty fieldType
	imports.
		Matcher(left.Name, right.Name).
		Match(
			func() {
				ty = optionFieldType
			},
			func(isSlice bool) {
				if !isSlice {
					ty = undFieldType
				} else {
					ty = undSliceFieldTye
				}
			},
			func(isSlice bool) {
				if !isSlice {
					ty = elasticFieldType
				} else {
					ty = elasticSliceFieldType
				}
			},
		)
	modified = modifyUndField(f, imports, fieldTy, x, left, right, undOpt)
	fieldConverter = converter(ty, undOpt, imports, typeParam)
	return modified, fieldConverter, nil
}

// fieldTy is a type-asserted field type. It cannot be other than IndexedExpr since it is expected to be like option.Option[T].
// x is fieldTy's former part, e.g. option.Option.
// left and right is splitted x's elector, e.g. left is option and right is Option.
// undOpt is parsed struct tag.
// ok is true only when result is valid and target und type field.
// err is non-nil only when und struct tag is malformed.
func isUndField(field *ast.Field, imports UndImports) (
	fieldTy *ast.IndexExpr,
	x *ast.SelectorExpr,
	left, right *ast.Ident,
	undOpt structtag.UndOpt,
	ok bool,
	err error,
) {
	fieldTy, ok = field.Type.(*ast.IndexExpr)
	if !ok {
		// no traversal for now.
		return
	}
	// It's not possible to placing und types without selection since it's outer type.
	// Do not alias.
	x, ok = fieldTy.X.(*ast.SelectorExpr)
	if !ok {
		return
	}
	left, ok = x.X.(*ast.Ident)
	if !ok || left == nil {
		return
	}
	right = x.Sel
	if !imports.Has(left.Name, right.Name) {
		return
	}

	tag := field.Tag.Value
	if tag == "" {
		return
	}

	tag = reflect.StructTag(tag[1 : len(tag)-1]).Get(structtag.UndTag)
	if tag == "" {
		return
	}

	undOpt, err = structtag.ParseOption(tag)
	if err != nil {
		return
	}

	ok = true
	return
}

func modifyUndField(
	field *ast.Field,
	imports UndImports,
	fieldTy *ast.IndexExpr,
	x *ast.SelectorExpr,
	left *ast.Ident, right *ast.Ident,
	undOpt structtag.UndOpt,
) (modified bool) {
	astutil.Apply(
		field,
		func(c *astutil.Cursor) bool {
			f, ok := c.Node().(*ast.Field)
			if !ok {
				return true
			}
			modified = true
			imports.Matcher(left.Name, right.Name).Match(
				func() {
					switch s := undOpt.States.Value(); {
					default:
						modified = false
					case s.Def && (s.Null || s.Und):
						modified = false
					case s.Def:
						f.Type = fieldTy.Index // unwrap, simply T.
						c.Replace(f)
					case s.Null || s.Und:
						f.Type = startStructExpr() // *struct{}
						c.Replace(f)
					}
				},
				func(_ bool) {
					switch s := undOpt.States.Value(); {
					case s.Def && s.Null && s.Und:
						modified = false
					case s.Def && (s.Null || s.Und):
						*x = *optionExpr(imports)
						c.Replace(f)
					case s.Null && s.Und:
						fieldTy.Index = startStructExpr()
						*x = *optionExpr(imports)
						c.Replace(f)
					case s.Def:
						f.Type = fieldTy.Index
						c.Replace(f)
					case s.Null || s.Und:
						f.Type = startStructExpr()
						c.Replace(f)
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
					// und.Und[[]option.Option[T]]
					if isSlice {
						fieldTy.X = sliceUndExpr(imports)
					} else {
						fieldTy.X = undExpr(imports)
					}
					fieldTy.Index = &ast.ArrayType{
						Elt: &ast.IndexExpr{
							X:     optionExpr(imports),
							Index: fieldTy.Index,
						},
					}

					if undOpt.Len.IsSome() {
						lv := undOpt.Len.Value()
						if lv.Op == structtag.LenOpEqEq {
							if lv.Len == 1 {
								// und.Und[[]option.Option[T]] -> und.Und[option.Option[T]]
								fieldTy.Index = fieldTy.Index.(*ast.ArrayType).Elt
							} else {
								// und.Und[[]option.Option[T]] -> und.Und[[n]option.Option[T]]
								fieldTy.Index.(*ast.ArrayType).Len = &ast.BasicLit{
									Kind:  token.INT,
									Value: strconv.FormatInt(int64(undOpt.Len.Value().Len), 10),
								}
							}
						}
					}

					if undOpt.Values.IsSome() {
						switch x := undOpt.Values.Value(); {
						case x.Nonnull:
							switch x := fieldTy.Index.(type) {
							case *ast.ArrayType:
								// und.Und[[n]option.Option[T]] -> und.Und[[n]T]
								x.Elt = x.Elt.(*ast.IndexExpr).Index
							case *ast.IndexExpr:
								// und.Und[option.Option[T]] -> und.Und[T]
								fieldTy.Index = x.Index
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
					case s.Def && s.Null && s.Und:
						// no conversion
					case s.Def && (s.Null || s.Und):
						// und.Und[[]option.Option[T]] -> option.Option[[]option.Option[T]]
						fieldTy.X = optionExpr(imports)
						c.Replace(f)
					case s.Null && s.Und:
						// option.Option[*struct{}]
						fieldTy.Index = startStructExpr()
						fieldTy.X = optionExpr(imports)
						c.Replace(f)
					case s.Def:
						// und.Und[[]option.Option[T]] -> []option.Option[T]
						f.Type = fieldTy.Index
						c.Replace(f)
					case s.Null || s.Und:
						f.Type = startStructExpr()
						c.Replace(f)
					}
				},
			)
			return false
		},
		nil,
	)
	return
}

func converter(ty fieldType, undOpt structtag.UndOpt, imports UndImports, typeParam ast.Node) fieldConverter {
	switch ty {
	case optionFieldType:
		c := optionConverter(undOpt)
		if c != nil {
			return c
		}
	case undFieldType, undSliceFieldTye:
		c := undConverter(undOpt.States.Value(), imports)
		if c != nil {
			return c
		}
	case elasticFieldType, elasticSliceFieldType:
		c := newElasticConverter(undOpt, imports, ty == elasticSliceFieldType, typeParam)
		if c != nil {
			return c
		}
	}
	return nil
}

type fieldConverter interface {
	Expr(field string) string
}

type genericConverter struct {
	Selector string
	// AdditionalImports map[string]string
	Method string
	Nil    bool
	// added before field.
	Args     []string
	TypePram []string
}

func nullishConverter(imports UndImports) *genericConverter {
	return &genericConverter{
		Selector: imports.conversion,
		Method:   "UndNullish",
	}
}

func optionConverter(undOpt structtag.UndOpt) *genericConverter {
	switch s := undOpt.States.Value(); {
	default:
		return nil
	case s.Def && (s.Null || s.Und):
		return nil
	case s.Def:
		return &genericConverter{
			Method: "Value",
		}
	case s.Null || s.Und:
		return &genericConverter{
			Nil: true,
		}
	}
}

func undConverter(states structtag.States, imports UndImports) *genericConverter {
	switch s := states; {
	default:
		return nil
	case s.Def && s.Null && s.Und:
		return nil
	case s.Def && (s.Null || s.Und):
		return &genericConverter{
			Method: "Unwrap().Value",
		}
	case s.Null && s.Und:
		return nullishConverter(imports)
	case s.Def:
		return &genericConverter{
			Method: "Value",
		}
	case s.Null || s.Und:
		return &genericConverter{
			Nil: true,
		}
	}
}

func (m *genericConverter) Expr(
	field string,
) string {
	if m.Nil {
		return "nil"
	}
	if m.Selector == "" {
		return field + "." + m.Method + "()"
	}

	var instantiation string
	if len(m.TypePram) > 0 {
		instantiation = "[" + strings.Join(m.TypePram, ",") + "]"
	}
	ident := m.Selector
	if ident == "." {
		ident = ""
	} else {
		ident += "."
	}
	var args []string
	if len(m.Args) > 0 {
		args = append(args, m.Args...)
	}
	args = append(args, field)
	return ident + m.Method + instantiation + "(" + strings.Join(args, ",") + ")"
}

type elasticConverter struct {
	g        *genericConverter
	wrappers []fieldConverter
}

func newElasticConverter(
	undOpt structtag.UndOpt,
	imports UndImports,
	isSlice bool,
	typeParam ast.Node,
) *elasticConverter {
	// very really simple case.
	if undOpt.States.IsSome() && undOpt.Len.IsNone() && undOpt.Values.IsNone() {
		switch s := undOpt.States.Value(); {
		default: //case s.Def && s.Null && s.Und:
			return nil
		case s.Def && (s.Null || s.Und):
			return &elasticConverter{
				g: &genericConverter{
					Selector: imports.conversion,
					Method:   suffixSlice("UnwrapElastic", isSlice),
				},
				wrappers: []fieldConverter{
					&genericConverter{
						Method: "Unwrap().Value",
					},
				},
			}
		case s.Null && s.Und:
			return &elasticConverter{g: nullishConverter(imports)}
		case s.Def:
			return &elasticConverter{
				g: &genericConverter{ // []option.Option[T]
					Method: "Unwrap().Value",
				},
			}
		case s.Null || s.Und:
			return &elasticConverter{
				g: &genericConverter{
					Nil: true,
				},
			}
		}
	}

	states := undOpt.States.Value()
	if undOpt.Len.IsSome() {
		states.Def = true
	}
	if !states.Def {
		// return early.
		switch s := states; {
		case s.Null && s.Und:
			return &elasticConverter{
				g: nullishConverter(imports),
			}
		case s.Null || s.Und:
			return &elasticConverter{
				g: &genericConverter{
					Nil: true,
				},
			}
		}
	}
	// fist, converts Elastic[T] -> Und[[]option.Option[T]]
	c := &elasticConverter{
		g: &genericConverter{
			Selector: imports.conversion,
			Method:   suffixSlice("UnwrapElastic", isSlice),
		},
	}
	if undOpt.Len.IsSome() {
		// if len is set, map it into Und[[n]option.Option[T]]
		lv := undOpt.Len.Value()
		var wrapper fieldConverter
		switch lv.Op {
		case structtag.LenOpEqEq:
			// to [n]option.Option[T]
			wrapper = &templateConverter{
				t: undFixedSize,
				p: newTemplateParams(imports, isSlice, "", typeParam, lv.Len),
			}
			// other then trim down or append it to the size at most or at least.
		case structtag.LenOpGr:
			wrapper = &genericConverter{
				Selector: imports.conversion,
				Method:   suffixSlice("LenNAtLeast", isSlice),
				Args:     []string{strconv.FormatInt(int64(lv.Len+1), 10)},
			}
		case structtag.LenOpGrEq:
			wrapper = &genericConverter{
				Selector: imports.conversion,
				Method:   suffixSlice("LenNAtLeast", isSlice),
				Args:     []string{strconv.FormatInt(int64(lv.Len), 10)},
			}
		case structtag.LenOpLe:
			wrapper = &genericConverter{
				Selector: imports.conversion,
				Method:   suffixSlice("LenNAtMost", isSlice),
				Args:     []string{strconv.FormatInt(int64(lv.Len-1), 10)},
			}
		case structtag.LenOpLeEq:
			wrapper = &genericConverter{
				Selector: imports.conversion,
				Method:   suffixSlice("LenNAtMost", isSlice),
				Args:     []string{strconv.FormatInt(int64(lv.Len), 10)},
			}
		}
		c.wrappers = append(c.wrappers, wrapper)
	}
	if undOpt.Values.IsSome() {
		v := undOpt.Values.Value()
		var wrapper fieldConverter
		switch {
		case v.Nonnull:
			if undOpt.Len.IsSomeAnd(func(lv structtag.LenValidator) bool { return lv.Op == structtag.LenOpEqEq }) {
				wrapper = &templateConverter{
					t: mapUndNonNullFixedSize,
					p: newTemplateParams(imports, isSlice, "", typeParam, undOpt.Len.Value().Len),
				}
			} else {
				wrapper = &genericConverter{
					Selector: imports.conversion,
					Method:   suffixSlice("NonNull", isSlice),
				}
			}
			// no other cases at the moment. I don't expect it to expand tho.
		}
		c.wrappers = append(c.wrappers, wrapper)
	}

	// Then when len == 1, convert und.Und[[1]option.Option[T]] or und.Und[[1]T] to und.Und[option.Option[T]], und.Und[T] respectively
	if undOpt.Len.IsSomeAnd(func(lv structtag.LenValidator) bool { return lv.Op == structtag.LenOpEqEq && lv.Len == 1 }) {
		c.wrappers = append(
			c.wrappers,
			&genericConverter{
				Selector: imports.conversion,
				Method:   suffixSlice("UnwrapLen1", isSlice),
			})
	}

	// Finally unwrap value based on req,null,und
	if wrapper := undConverter(states, imports); wrapper != nil {
		c.wrappers = append(c.wrappers, wrapper)
	}

	return c
}

func (c *elasticConverter) Expr(
	field string,
) string {
	expr := c.g.Expr(field)
	for _, wrapper := range c.wrappers {
		expr = wrapper.Expr(expr)
		var ok bool
		expr, ok = strings.CutSuffix(expr, ")")
		if ok {
			s := strings.TrimLeftFunc(expr, unicode.IsSpace)
			switch s[len(s)-1] {
			case '(', '\n':
				expr += ")"
			default:
				expr += ",\n)"
			}
		}
	}
	return expr
}

// returns *struct{}
func startStructExpr() *ast.StarExpr {
	return &ast.StarExpr{
		X: &ast.StructType{
			Fields: &ast.FieldList{Opening: 1, Closing: 2},
		},
	}
}

func undExpr(imports UndImports) *ast.SelectorExpr {
	return &ast.SelectorExpr{
		X: &ast.Ident{
			Name: imports.und,
		},
		Sel: &ast.Ident{
			Name: "Und",
		},
	}
}

func sliceUndExpr(imports UndImports) *ast.SelectorExpr {
	return &ast.SelectorExpr{
		X: &ast.Ident{
			Name: imports.sliceUnd,
		},
		Sel: &ast.Ident{
			Name: "Und",
		},
	}
}

// option.Option
func optionExpr(imports UndImports) *ast.SelectorExpr {
	return &ast.SelectorExpr{
		X: &ast.Ident{
			Name: imports.option,
		},
		Sel: &ast.Ident{
			Name: "Option",
		},
	}
}

type templateConverter struct {
	t *template.Template
	p templateParams
}

func (c *templateConverter) Expr(
	field string,
) string {
	param := c.p
	param.Arg = field
	var buf bytes.Buffer
	err := c.t.Execute(&buf, param)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

type templateParams struct {
	OptionPkg  string
	UndPkg     string
	ElasticPkg string
	Arg        string
	TypeParam  string
	Size       string
}

func newTemplateParams(
	imports UndImports,
	isSlice bool,
	arg string,
	typeParam ast.Node,
	size int,
) templateParams {
	var buf bytes.Buffer
	fset := token.NewFileSet()
	err := printer.Fprint(&buf, fset, typeParam)
	if err != nil {
		panic(err)
	}
	var sizeStr string
	if size > 0 {
		sizeStr = strconv.FormatInt(int64(size), 10)
	}
	return templateParams{
		OptionPkg:  imports.option,
		UndPkg:     imports.Und(isSlice),
		ElasticPkg: imports.Elastic(isSlice),
		Arg:        arg,
		TypeParam:  buf.String(),
		Size:       sizeStr,
	}
}

func suffixSlice(s string, suffix bool) string {
	if suffix {
		return s + "Slice"
	}
	return s
}

var (
	undFixedSize = template.Must(template.New("").Parse(
		`{{.UndPkg}}.Map(
	{{.Arg}},
	func(s []{{.OptionPkg}}.Option[{{.TypeParam}}]) (r [{{.Size}}]{{.OptionPkg}}.Option[{{.TypeParam}}]) {
		copy(r[:], s)
		return
	},
)`))
	mapUndNonNullFixedSize = template.Must(template.New("").Parse(
		`{{.UndPkg}}.Map(
	{{.Arg}},
	func(s [{{.Size}}]{{.OptionPkg}}.Option[{{.TypeParam}}]) (r [{{.Size}}]{{.TypeParam}}) {
		for i := 0; i < {{.Size}}; i++ {
			r[i] = s[i].Value()
		}
		return
	},
)`,
	))
)
