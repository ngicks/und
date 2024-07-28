package undgen

import (
	"go/ast"
	"strconv"

	"github.com/ngicks/und/internal/structtag"
)

func undPlainFieldConverter(
	f *ast.Field,
	imports UndImports,
) (fieldConverter, error) {
	fieldTy, _, left, right, undOpt, ok, err := isUndField(f, imports)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	typeParam := fieldTy.Index

	var r fieldConverter
	imports.
		Matcher(left.Name, right.Name).
		Match(
			func() {
				c := optionUndPlainConverter(undOpt)
				if c != nil {
					r = c
				}
			},
			func(_ bool) {
				c := undUndPlainConverter(undOpt.States.Value(), imports)
				if c != nil {
					r = c
				}
			},
			func(isSlice bool) {
				c := elasticUndPlainConverter(undOpt, imports, isSlice, typeParam)
				if c != nil {
					r = c
				}
			},
		)
	return r, nil
}

func optionUndPlainConverter(undOpt structtag.UndOpt) fieldConverter {
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
		return nilSimpleExpr()
	}
}

func undUndPlainConverter(states structtag.States, imports UndImports) fieldConverter {
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
		return nilSimpleExpr()
	}
}

func elasticUndPlainConverter(
	undOpt structtag.UndOpt,
	imports UndImports,
	isSlice bool,
	typeParam ast.Node,
) *nestedConverter {
	// very really simple case.
	if undOpt.States.IsSome() && undOpt.Len.IsNone() && undOpt.Values.IsNone() {
		switch s := undOpt.States.Value(); {
		default:
			return nil
		case s.Def && s.Null && s.Und:
			return nil
		case s.Def && (s.Null || s.Und):
			return &nestedConverter{
				core: &genericConverter{
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
			return &nestedConverter{core: nullishConverter(imports)}
		case s.Def:
			return &nestedConverter{
				core: &genericConverter{ // []option.Option[T]
					Method: "Unwrap().Value",
				},
			}
		case s.Null || s.Und:
			return &nestedConverter{
				core: nilSimpleExpr(),
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
			return &nestedConverter{
				core: nullishConverter(imports),
			}
		case s.Null || s.Und:
			return &nestedConverter{
				core: nilSimpleExpr(),
			}
		}
	}
	// fist, converts Elastic[T] -> Und[[]option.Option[T]]
	c := &nestedConverter{
		core: &genericConverter{
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
	if wrapper := undUndPlainConverter(states, imports); wrapper != nil {
		c.wrappers = append(c.wrappers, wrapper)
	}

	return c
}
