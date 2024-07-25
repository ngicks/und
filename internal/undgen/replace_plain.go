package undgen

import (
	"go/ast"
	"go/token"
	"reflect"
	"strconv"

	"github.com/ngicks/und/internal/structtag"
	"golang.org/x/tools/go/ast/astutil"
)

func checkAndModifyUndField(
	f *ast.Field,
	imports UndImports,
) (modified bool, err error) {
	fieldTy, x, left, right, undOpt, ok, err := isUndField(f, imports)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	modified = modifyUndField(f, imports, fieldTy, x, left, right, undOpt)
	return modified, nil
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
