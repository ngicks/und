package undgen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"

	"github.com/ngicks/und/internal/structtag"
)

func undRawFieldBackConverter(
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
				c := optionUndRawConverter(undOpt, imports, typeParam)
				if c != nil {
					r = c
				}
			},
			func(isSlice bool) {
				c := undUndRawConverter(undOpt.States.Value(), imports, typeParam, isSlice)
				if c != nil {
					r = c
				}
			},
			func(isSlice bool) {
				c := elasticUndRawConverter(undOpt, imports, isSlice, typeParam)
				if c != nil {
					r = c
				}
			},
		)
	return r, nil
}

func optionUndRawConverter(undOpt structtag.UndOpt, imports UndImports, typeParam ast.Node) *genericConverter {
	switch s := undOpt.States.Value(); {
	default:
		return nil
	case s.Def && (s.Null || s.Und):
		return nil
	case s.Def:
		return &genericConverter{
			Selector: imports.option,
			Method:   "Some",
		}
	case s.Null || s.Und:
		var buf bytes.Buffer
		fset := token.NewFileSet()
		err := printer.Fprint(&buf, fset, typeParam)
		if err != nil {
			panic(err)
		}
		return &genericConverter{
			Selector: imports.option,
			Method:   "None",
			TypePram: []string{buf.String()},
			OmitArg:  true,
		}
	}
}

func undUndRawConverter(states structtag.States, imports UndImports, typeParam ast.Node, isSlice bool) *genericConverter {
	var buf bytes.Buffer
	fset := token.NewFileSet()
	err := printer.Fprint(&buf, fset, typeParam)
	if err != nil {
		panic(err)
	}
	switch s := states; {
	default:
		return nil
	case s.Def && s.Null && s.Und:
		return nil
	case s.Def && (s.Null || s.Und):
		return &genericConverter{
			Selector: imports.conversion,
			Args:     []string{fmt.Sprintf("%t", s.Null)},
			Method:   suffixSlice("MapOptionToUnd", isSlice),
		}
	case s.Null && s.Und:
		return &genericConverter{
			Selector: imports.conversion,
			Method:   suffixSlice("UndNullishBack", isSlice),
			TypePram: []string{buf.String()},
		}
	case s.Def:
		return &genericConverter{
			Selector: imports.Und(isSlice),
			Method:   "Defined",
		}
	case s.Null || s.Und:
		return &genericConverter{
			Selector: imports.Und(isSlice),
			Method: func() string {
				if s.Null {
					return "Null"
				} else {
					return "Undefined"
				}
			}(),
			TypePram: []string{buf.String()},
			OmitArg:  true,
		}
	}
}

func elasticUndRawConverter(
	undOpt structtag.UndOpt,
	imports UndImports,
	isSlice bool,
	typeParam ast.Node,
) *nestedConverter {
	var buf bytes.Buffer
	fset := token.NewFileSet()
	err := printer.Fprint(&buf, fset, typeParam)
	if err != nil {
		panic(err)
	}
	typeParamStr := buf.String()

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
					Method:   suffixSlice("MapOptionOptionToElastic", isSlice),
					Args:     []string{fmt.Sprintf("%t", s.Null)},
				},
			}
		case s.Null && s.Und:
			return &nestedConverter{core: &genericConverter{
				Selector: imports.conversion,
				Method:   suffixSlice("UndNullishBackElastic", isSlice),
				TypePram: []string{typeParamStr},
			}}
		case s.Def:
			return &nestedConverter{
				core: &genericConverter{ // []option.Option[T]
					Selector: imports.Elastic(isSlice),
					Method:   "FromOptions",
				},
			}
		case s.Null || s.Und:
			return &nestedConverter{
				core: &genericConverter{
					Selector: imports.Elastic(isSlice),
					Method: func() string {
						if s.Null {
							return "Null"
						} else {
							return "Undefined"
						}
					}(),
					TypePram: []string{typeParamStr},
					OmitArg:  true,
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
			return &nestedConverter{core: &genericConverter{
				Selector: imports.conversion,
				Method:   suffixSlice("UndNullishBackElastic", isSlice),
				TypePram: []string{typeParamStr},
			}}
		case s.Null || s.Und:
			return &nestedConverter{
				core: &genericConverter{
					Selector: imports.Elastic(isSlice),
					Method: func() string {
						if s.Null {
							return "Null"
						} else {
							return "Undefined"
						}
					}(),
					TypePram: []string{typeParamStr},
					OmitArg:  true,
				},
			}
		}
	}
	// Below converts much like UndPlain but reversed order.
	// At last, Und[[]option.Option[T]] -> converts Elastic[T]
	c := &nestedConverter{
		core: noopFieldConverter{},
		wrappers: []fieldConverter{&genericConverter{
			Selector: imports.Elastic(isSlice),
			Method:   "FromUnd",
		}},
	}
	if undOpt.Len.IsSome() {
		// if len is EqEq, map Und[[n]option.Option[T]] -> Und[[]option.Option[T]]
		lv := undOpt.Len.Value()
		switch lv.Op {
		case structtag.LenOpEqEq:
			c.wrappers = append(
				[]fieldConverter{
					&templateConverter{
						t: mapUndFixedToSlice,
						p: newTemplateParams(imports, isSlice, "", typeParam, undOpt.Len.Value().Len),
					},
				},
				c.wrappers...,
			)
		}
	}
	if undOpt.Values.IsSome() {
		v := undOpt.Values.Value()
		var wrapper fieldConverter
		// Und[[n]T] -> Und[[n]option.Option[T]]
		switch {
		case v.Nonnull:
			if undOpt.Len.IsSomeAnd(func(lv structtag.LenValidator) bool { return lv.Op == structtag.LenOpEqEq }) {
				wrapper = &templateConverter{
					t: nullifyUndFixedSize,
					p: newTemplateParams(imports, isSlice, "", typeParam, undOpt.Len.Value().Len),
				}
			} else {
				wrapper = &genericConverter{
					Selector: imports.conversion,
					Method:   suffixSlice("Nullify", isSlice),
				}
			}
			// no other cases at the moment. I don't expect it to expand tho.
		}
		c.wrappers = append([]fieldConverter{wrapper}, c.wrappers...) // prepend
	}

	// When len == 1, convert und.Und[[1]option.Option[T]] or und.Und[[1]T] to und.Und[option.Option[T]], und.Und[T] respectively
	if undOpt.Len.IsSomeAnd(func(lv structtag.LenValidator) bool { return lv.Op == structtag.LenOpEqEq && lv.Len == 1 }) {
		c.wrappers = append(
			[]fieldConverter{&genericConverter{
				Selector: imports.conversion,
				Method:   suffixSlice("WrapLen1", isSlice),
			}},
			c.wrappers...,
		) // prepend.
	}

	// Finally wrap value based on req,null,und
	if wrapper := undUndRawConverter(states, imports, typeParam, isSlice); wrapper != nil {
		c.wrappers = append([]fieldConverter{wrapper}, c.wrappers...)
	}

	return c
}
