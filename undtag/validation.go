package undtag

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ngicks/und/internal/option"
)

const (
	// "und" struct tag tells tools like ../validate or github.com/ngicks/go-codegen how fields should be treated.
	//
	// example:
	// type Sample struct {
	// 	Foo string `und:"def,und"`
	// }
	TagName = "und"
	// The field must be required(Some or Defined).
	// mutually exclusive to nullish, def, null, und.
	// UndTagValueRequired can be combined with len (there's no point though).
	//
	// example:
	// type Sample struct {
	// 	Foo string `und:"required"`
	// }
	UndTagValueRequired = "required"
	// The field must be nullish(None, Null, Undefined).
	// mutually exclusive to required, def, null, und.
	// UndTagValueNullish can be combined with len.
	//
	// example:
	// type Sample struct {
	// 	Foo string `und:"nullish"`
	// }
	UndTagValueNullish = "nullish"
	// The field is allowed to be Some or Defined.
	// can be combined with null, und or len.
	//
	// example:
	// type Sample struct {
	// 	Foo string `und:"def"`
	// }
	UndTagValueDef = "def"
	// The field is allowed to be None or Null.
	// can be combined with def, und or len.
	//
	// example:
	// type Sample struct {
	// 	Foo string `und:"null"`
	// }
	UndTagValueNull = "null"
	// The field is allowed to be None or Undefined.
	// can be combined with def, null or len.
	//
	// example:
	// type Sample struct {
	// 	Foo string `und:"und"`
	// }
	UndTagValueUnd = "und"
	// Only for elastic types.
	//
	// The value must be formatted as len==n, len>n, len>=n, len<n or len<=n,
	// where n is unsigned integer.
	// The field's length will be evaluated as (length) (comparison operator) (n),
	// e.g. if tag is len>12, field.Len() > 12 must return true.
	//
	// can be combined with other options.
	//
	// example:
	// type Sample struct {
	// 	Foo string `und:"len==3"`
	// }
	UndTagValueLen = "len"
	// Only for elastic types.
	//
	// The value must be formatted as values:nonnull.
	//
	// nonnull value means its internal value must not have null.
	//
	// example:
	// type Sample struct {
	// 	Foo string `und:"values:nonnull"`
	// }
	UndTagValueValues = "values"
)

var (
	// ErrEmptyOption would be returned by [ParseOption]
	// if input's `und` struct tag have no option specified.
	ErrEmptyOption = errors.New("empty option")
	// ErrMultipleOption would be returned by UndValidate and UndCheck
	// if input's `und` struct tags have multiple mutually exclusive options.
	ErrMultipleOption = errors.New("multiple option")
	// ErrUnknownOption is an error value which will be returned by UndValidate and UndCheck
	// if an input has unknown options in `und` struct tag.
	ErrUnknownOption = errors.New("unknown option")
	// ErrMalformedLen is an error which will be returned by UndValidate and UndCheck
	// if an input has malformed len option in `und` struct tag.
	ErrMalformedLen = errors.New("malformed len")
	// ErrMalformedLen is an error which will be returned by UndValidate and UndCheck
	// if an input has malformed values option in `und` struct tag.
	ErrMalformedValues = errors.New("malformed values")
)

type ElasticLike interface {
	UndLike
	Len() int
	HasNull() bool
}

type UndLike interface {
	IsDefined() bool
	IsNull() bool
	IsUndefined() bool
}

type OptionLike interface {
	IsNone() bool
	IsSome() bool
}

type UndOptExport struct {
	States *StateValidator
	Len    *LenValidator
	Values *ValuesValidator
}

func (o UndOptExport) Into() UndOpt {
	// the outer code can not initialize UndOpt itself since it uses internal package.
	// Export type can not rely on Option like types.
	return UndOpt{
		states: option.FromPointer(o.States),
		len:    option.FromPointer(o.Len),
		values: option.FromPointer(o.Values),
	}
}

type UndOpt struct {
	// TODO: warn user about use of internal package?
	// I suspect they don't realize these are actually vendored internal option package.
	states option.Option[StateValidator]
	len    option.Option[LenValidator]
	values option.Option[ValuesValidator]
}

func ParseOption(s string) (UndOpt, error) {
	if s == "" {
		return UndOpt{}, ErrEmptyOption
	}
	org := s
	var (
		opt         string
		sawStateOpt bool
		opts        UndOpt
	)
	for len(s) > 0 {
		opt, s, _ = strings.Cut(s, ",")
		if strings.HasPrefix(opt, UndTagValueLen) {
			if opts.len.IsSome() {
				return UndOpt{}, fmt.Errorf("%w: %s", ErrMultipleOption, org)
			}
			lenV, err := ParseLen(opt)
			if err != nil {
				return UndOpt{}, fmt.Errorf("%w: %w", ErrMalformedLen, err)
			}
			opts.states = opts.states.
				Or(option.Some(StateValidator{})).
				Map(func(v StateValidator) StateValidator { v.Def = true; return v })
			opts.len = option.Some(lenV)
			continue
		}

		if strings.HasPrefix(opt, UndTagValueValues) {
			if opts.values.IsSome() {
				return UndOpt{}, fmt.Errorf("%w: %s", ErrMultipleOption, org)
			}
			valuesV, err := ParseValues(opt)
			if err != nil {
				return UndOpt{}, fmt.Errorf("%w: %w", ErrMalformedValues, err)
			}
			opts.values = option.Some(valuesV)
			continue
		}

		switch opt {
		case UndTagValueRequired, UndTagValueNullish:
			if sawStateOpt {
				return UndOpt{}, fmt.Errorf("%w: und tag contains multiple mutually exclusive options, tag = %s", ErrMultipleOption, org)
			}
		case UndTagValueDef, UndTagValueNull, UndTagValueUnd:
			if opts.states.IsSomeAnd(func(s StateValidator) bool {
				return s.filled || opt == UndTagValueDef && s.Def || opt == UndTagValueNull && s.Null || opt == UndTagValueUnd && s.Und
			}) {
				return UndOpt{}, fmt.Errorf("%w: und tag contains multiple mutually exclusive options, tag = %s", ErrMultipleOption, org)
			}
		default:
			return UndOpt{}, ErrUnknownOption
		}

		opts.states = opts.states.Or(option.Some(StateValidator{})).Map(func(v StateValidator) StateValidator {
			switch opt {
			case UndTagValueRequired:
				v.filled = true
				v.Def = true
			case UndTagValueNullish:
				v.filled = true
				v.Null = true
				v.Und = true
			case UndTagValueDef:
				v.Def = true
			case UndTagValueNull:
				v.Null = true
			case UndTagValueUnd:
				v.Und = true
			}
			return v
		})

		sawStateOpt = true
	}

	return opts, nil
}

func (u UndOpt) States() option.Option[StateValidator] {
	return u.states
}

func (u UndOpt) Len() option.Option[LenValidator] {
	return u.len
}

func (u UndOpt) Values() option.Option[ValuesValidator] {
	return u.values
}

func (o UndOpt) Describe() string {
	var builder strings.Builder

	appendStr := func(s interface{ Describe() string }) {
		ss := s.Describe()
		if builder.Len() > 0 && len(ss) > 0 {
			_, _ = builder.WriteString(", and ")
		}
		_, _ = builder.WriteString(ss)
	}

	if o.states.IsSome() {
		appendStr(o.states.Value())
	}
	if o.len.IsSome() {
		appendStr(o.len.Value())
	}
	if o.values.IsSome() {
		appendStr(o.values.Value())
	}

	return builder.String()
}

func (o UndOpt) ValidOpt(opt OptionLike) bool {
	return o.states.IsSomeAnd(func(s StateValidator) bool {
		switch {
		case opt.IsSome():
			return s.Def
		default: // opt.IsNone():
			return s.Null || s.Und
		}
	})
}

func (o UndOpt) ValidUnd(u UndLike) bool {
	return o.states.IsSomeAnd(func(s StateValidator) bool {
		switch {
		case u.IsDefined():
			return s.Def
		case u.IsNull():
			return s.Null
		default: // case u.IsUndefined():
			return s.Und
		}
	})
}

func or[T, U any](t option.Option[T], u option.Option[U]) option.Option[struct{}] {
	if t.IsSome() || u.IsSome() {
		return option.Some(struct{}{})
	}
	return option.None[struct{}]()
}

func (o UndOpt) ValidElastic(e ElasticLike) bool {
	// there's possible many intersection
	validState := option.Map(
		o.states,
		func(s StateValidator) bool {
			return s.Valid(e)
		},
	).
		Or(option.Some(true))
	if !validState.Value() {
		return false
	}
	if !e.IsDefined() {
		return true
	}
	return option.Map(
		or(o.len, o.values),
		func(_ struct{}) bool {
			validLen := option.Map(
				o.len,
				func(s LenValidator) bool { return s.Valid(e) },
			).
				Or(option.Some(true))
			if !validLen.Value() {
				return false
			}

			validValue := option.Map(
				o.values,
				func(s ValuesValidator) bool {
					return s.Valid(e)
				},
			).
				Or(option.Some(true))

			return validValue.Value()
		},
	).Or(option.Some(true)).Value()
}

type StateValidator struct {
	filled bool
	Def    bool
	Null   bool
	Und    bool
}

func (s StateValidator) Valid(u UndLike) bool {
	switch {
	case u.IsDefined():
		return s.Def
	case u.IsNull():
		return s.Null
	default: // case u.IsUndefined():
		return s.Und
	}
}

func (s StateValidator) Describe() string {
	if s.filled {
		if s.Def {
			return "is " + UndTagValueRequired
		} else {
			return "is " + UndTagValueNullish
		}
	}
	var builder strings.Builder
	if s.Def {
		builder.WriteString("defined")
	}
	if s.Null {
		if builder.Len() > 0 {
			builder.WriteString(" or ")
		}
		builder.WriteString("null")
	}
	if s.Und {
		if builder.Len() > 0 {
			builder.WriteString(" or ")
		}
		builder.WriteString("undefined")
	}
	return "must be " + builder.String()
}

type LenValidator struct {
	Len int
	Op  lenOp
}

func ParseLen(s string) (LenValidator, error) {
	org := s
	s, _ = strings.CutPrefix(s, UndTagValueLen)
	if len(s) < 2 { // <n, at least 2.
		return LenValidator{}, fmt.Errorf("unknown op: %s", org)
	}
	var v LenValidator
	switch {
	default:
		return LenValidator{}, fmt.Errorf("unknown op: %s", org)
	case s[:2] == "==":
		v.Op = LenOpEqEq
	case s[:2] == ">=":
		v.Op = LenOpGrEq
	case s[:2] == "<=":
		v.Op = LenOpLeEq
	case s[0] == '<':
		v.Op = LenOpLe
	case s[0] == '>':
		v.Op = LenOpGr
	}

	s = s[v.Op.len():]

	len, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return LenValidator{}, fmt.Errorf("unknown len: %w", err)
	}

	v.Len = int(len)
	return v, nil
}

func (v LenValidator) Describe() string {
	return "defined or must have length of " + v.Op.String() + " " + strconv.FormatInt(int64(v.Len), 10)
}

func (v LenValidator) Valid(e ElasticLike) bool {
	if v.Op == 0 {
		return true
	}
	return v.Op.Compare(e.Len(), v.Len)
}

type lenOp int

const (
	LenOpEqEq = lenOp(iota + 1) // ==
	LenOpGr                     // >
	LenOpGrEq                   // >=
	LenOpLe                     // <
	LenOpLeEq                   // <=
)

func (o lenOp) len() int {
	switch o {
	case LenOpLe, LenOpGr:
		return 1
	case LenOpEqEq, LenOpGrEq, LenOpLeEq:
		return 2
	}
	return 0
}

func (o lenOp) String() string {
	switch o {
	default: // case lenOpEqEq:
		return "=="
	case LenOpGr:
		return ">"
	case LenOpGrEq:
		return ">="
	case LenOpLe:
		return "<"
	case LenOpLeEq:
		return "<="
	}
}

func (o lenOp) Compare(i, j int) bool {
	switch o {
	default: // case lenOpEqEq:
		return i == j
	case LenOpGr:
		return i > j
	case LenOpGrEq:
		return i >= j
	case LenOpLe:
		return i < j
	case LenOpLeEq:
		return i <= j
	}
}

type ValuesValidator struct {
	Nonnull bool
}

func ParseValues(s string) (ValuesValidator, error) {
	org := s
	s, _ = strings.CutPrefix(s, UndTagValueValues)
	if len(s) < 2 || s[0] != ':' { // :nonull
		return ValuesValidator{}, fmt.Errorf("unknown op: %s", org)
	}

	s = s[1:] // removes ':'

	switch s {
	case "nonnull":
		return ValuesValidator{Nonnull: true}, nil
	}

	return ValuesValidator{}, fmt.Errorf("unknown op: %s", org)
}

func (v ValuesValidator) Valid(e ElasticLike) bool {
	switch {
	case v.Nonnull:
		return !e.HasNull()
	}
	return true
}

func (v ValuesValidator) Describe() string {
	switch {
	case v.Nonnull:
		return "must not contain null"
	}
	return ""
}
