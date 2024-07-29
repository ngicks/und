package undgen

import (
	"fmt"

	"github.com/dave/dst"
	"github.com/ngicks/und/internal/structtag"
)

func undRawFieldBackConverter(
	f *dst.Field,
	imports UndImports,
	fieldInfo undFieldInfo,
) (fieldConverter, error) {
	_, _, left, right, undOpt, ok, err := isUndField(f, imports)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	var r fieldConverter
	imports.
		Matcher(left.Name, right.Name).
		Match(
			func() {
				c := optionUndRawConverter(undOpt, imports, fieldInfo.TypeParm)
				if c != nil {
					r = c
				}
			},
			func(isSlice bool) {
				c := undUndRawConverter(undOpt.States.Value(), imports, fieldInfo.TypeParm, isSlice)
				if c != nil {
					r = c
				}
			},
			func(isSlice bool) {
				c := elasticUndRawConverter(undOpt, imports, isSlice, fieldInfo.TypeParm)
				if c != nil {
					r = c
				}
			},
		)
	return r, nil
}

func optionUndRawConverter(undOpt structtag.UndOpt, imports UndImports, typeParam string) *genericConverter {
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
		return &genericConverter{
			Selector: imports.option,
			Method:   "None",
			TypePram: []string{typeParam},
			OmitArg:  true,
		}
	}
}

func undUndRawConverter(states structtag.States, imports UndImports, typeParam string, isSlice bool) *genericConverter {
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
			TypePram: []string{typeParam},
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
			TypePram: []string{typeParam},
			OmitArg:  true,
		}
	}
}

func elasticUndRawConverter(
	undOpt structtag.UndOpt,
	imports UndImports,
	isSlice bool,
	typeParam string,
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
					Method:   suffixSlice("MapOptionOptionToElastic", isSlice),
					Args:     []string{fmt.Sprintf("%t", s.Null)},
				},
			}
		case s.Null && s.Und:
			return &nestedConverter{core: &genericConverter{
				Selector: imports.conversion,
				Method:   suffixSlice("UndNullishBackElastic", isSlice),
				TypePram: []string{typeParam},
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
					TypePram: []string{typeParam},
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
				TypePram: []string{typeParam},
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
					TypePram: []string{typeParam},
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
