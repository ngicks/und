package undgen

import (
	"go/token"
	"reflect"
	"slices"
	"strconv"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
	"github.com/ngicks/und/internal/undtag"
)

func checkAndModifyUndField(
	f *dst.Field,
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
func isUndField(field *dst.Field, imports UndImports) (
	fieldTy *dst.IndexExpr,
	x *dst.SelectorExpr,
	left, right *dst.Ident,
	undOpt undtag.UndOpt,
	ok bool,
	err error,
) {
	fieldTy, ok = field.Type.(*dst.IndexExpr)
	if !ok {
		// no traversal for now.
		return
	}
	// It's not possible to placing und types without selection since it's outer type.
	// Do not alias.
	x, ok = fieldTy.X.(*dst.SelectorExpr)
	if !ok {
		return
	}
	left, ok = x.X.(*dst.Ident)
	if !ok || left == nil {
		return
	}
	right = x.Sel
	if !imports.Has(left.Name, right.Name) {
		return
	}

	tag := ""
	if field.Tag != nil {
		tag = field.Tag.Value
	}
	if tag == "" {
		return
	}

	tag = reflect.StructTag(tag[1 : len(tag)-1]).Get(undtag.UndTag)
	if tag == "" {
		return
	}

	undOpt, err = undtag.ParseOption(tag)
	if err != nil {
		return
	}

	ok = true
	return
}

func modifyUndField(
	field *dst.Field,
	imports UndImports,
	fieldTy *dst.IndexExpr,
	x *dst.SelectorExpr,
	left *dst.Ident, right *dst.Ident,
	undOpt undtag.UndOpt,
) (modified bool) {
	dstutil.Apply(
		field,
		func(c *dstutil.Cursor) bool {
			f, ok := c.Node().(*dst.Field)
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
					if (undOpt.States.IsSomeAnd(func(s undtag.States) bool {
						return s.Def && s.Null && s.Und
					})) && (undOpt.Len.IsNone() || undOpt.Len.IsSomeAnd(func(lv undtag.LenValidator) bool {
						return lv.Op != undtag.LenOpEqEq
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
					fieldTy.Index = &dst.ArrayType{
						Elt: &dst.IndexExpr{
							X:     optionExpr(imports),
							Index: fieldTy.Index,
						},
					}

					if undOpt.Len.IsSome() {
						lv := undOpt.Len.Value()
						if lv.Op == undtag.LenOpEqEq {
							if lv.Len == 1 {
								// und.Und[[]option.Option[T]] -> und.Und[option.Option[T]]
								fieldTy.Index = fieldTy.Index.(*dst.ArrayType).Elt
							} else {
								// und.Und[[]option.Option[T]] -> und.Und[[n]option.Option[T]]
								fieldTy.Index.(*dst.ArrayType).Len = &dst.BasicLit{
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
							case *dst.ArrayType:
								// und.Und[[n]option.Option[T]] -> und.Und[[n]T]
								x.Elt = x.Elt.(*dst.IndexExpr).Index
							case *dst.IndexExpr:
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
func startStructExpr() *dst.StarExpr {
	return &dst.StarExpr{
		X: &dst.StructType{
			Fields: &dst.FieldList{Opening: true, Closing: true},
		},
	}
}

func undExpr(imports UndImports) *dst.SelectorExpr {
	return &dst.SelectorExpr{
		X: &dst.Ident{
			Name: imports.und,
		},
		Sel: &dst.Ident{
			Name: "Und",
		},
	}
}

func sliceUndExpr(imports UndImports) *dst.SelectorExpr {
	return &dst.SelectorExpr{
		X: &dst.Ident{
			Name: imports.sliceUnd,
		},
		Sel: &dst.Ident{
			Name: "Und",
		},
	}
}

// option.Option
func optionExpr(imports UndImports) *dst.SelectorExpr {
	return &dst.SelectorExpr{
		X: &dst.Ident{
			Name: imports.option,
		},
		Sel: &dst.Ident{
			Name: "Option",
		},
	}
}
