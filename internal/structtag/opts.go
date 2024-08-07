package structtag

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ngicks/und/internal/option"
)

const (
	UndTag = "und"
	// The field must be required(Some or Defined).
	// mutually exclusive to nullish, def, null, und.
	// UndTagValueRequired can be combined with len (there's no point though).
	UndTagValueRequired = "required"
	// The field must be nullish(None, Null, Undefined).
	// mutually exclusive to required, def, null, und.
	// UndTagValueNullish can be combined with len.
	UndTagValueNullish = "nullish"
	// The field is allowed to be Some or Defined.
	// can be combined with null, und or len.
	UndTagValueDef = "def"
	// The field is allowed to be None or Null.
	// can be combined with def, und or len.
	UndTagValueNull = "null"
	// The field is allowed to be None or Undefined.
	// can be combined with def, null or len.
	UndTagValueUnd = "und"
	// Only for elastic types.
	//
	// The value must be formatted as len==n, len>n, len>=n, len<n or len<=n,
	// where n is unsigned integer.
	// The field's length will be evaluated as (length) (comparison operator) (n),
	// e.g. if tag is len>12, field.Len() > 12 must return true.
	//
	// can be combined with other options.
	UndTagValueLen = "len"
	// Only for elastic types.
	//
	// The value must be formatted as values:nonnull.
	//
	// nonnull value means its internal value must not have null.
	UndTagValueValues = "values"
)

var (
	// ErrNotStruct would be returned by ValidateUnd and CheckUnd
	// if input is not a struct nor a pointer to a struct.
	ErrNotStruct = errors.New("not struct")
	// ErrMultipleOption would be returned by ValidateUnd and CheckUnd
	// if input's `und` struct tags have multiple mutually exclusive options.
	ErrMultipleOption = errors.New("multiple option")
	// ErrUnknownOption is an error value which will be returned by ValidateUnd and CheckUnd
	// if an input has unknown options in `und` struct tag.
	ErrUnknownOption = errors.New("unknown option")
	// ErrMalformedLen is an error which will be returned by ValidateUnd and CheckUnd
	// if an input has malformed len option in `und` struct tag.
	ErrMalformedLen = errors.New("malformed len")
	// ErrMalformedLen is an error which will be returned by ValidateUnd and CheckUnd
	// if an input has malformed values option in `und` struct tag.
	ErrMalformedValues = errors.New("malformed values")
)

type UndOpt struct {
	States option.Option[States]
	Len    option.Option[LenValidator]
	Values option.Option[ValuesValidator]
}

type States struct {
	filled bool
	Def    bool
	Null   bool
	Und    bool
}

func (s States) Valid(u UndLike) bool {
	switch {
	case u.IsDefined():
		return s.Def
	case u.IsNull():
		return s.Null
	default: // case u.IsUndefined():
		return s.Und
	}
}

func (s States) String() string {
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

func ParseOption(s string) (UndOpt, error) {
	org := s
	var (
		opt  string
		opts UndOpt
	)
	for len(s) > 0 {
		opt, s, _ = strings.Cut(s, ",")
		if strings.HasPrefix(opt, UndTagValueLen) {
			if opts.Len.IsSome() {
				return UndOpt{}, fmt.Errorf("%w: %s", ErrMultipleOption, org)
			}
			lenV, err := ParseLen(opt)
			if err != nil {
				return UndOpt{}, fmt.Errorf("%w: %w", ErrMalformedLen, err)
			}
			opts.Len = option.Some(lenV)
			continue
		}

		if strings.HasPrefix(opt, UndTagValueValues) {
			if opts.Values.IsSome() {
				return UndOpt{}, fmt.Errorf("%w: %s", ErrMultipleOption, org)
			}
			valuesV, err := ParseValues(opt)
			if err != nil {
				return UndOpt{}, fmt.Errorf("%w: %w", ErrMalformedValues, err)
			}
			opts.Values = option.Some(valuesV)
			continue
		}

		switch opt {
		case UndTagValueRequired, UndTagValueNullish:
			if opts.States.IsSome() {
				return UndOpt{}, fmt.Errorf("%w: und tag contains multiple mutually exclusive options, tag = %s", ErrMultipleOption, org)
			}
		case UndTagValueDef, UndTagValueNull, UndTagValueUnd:
			if opts.States.IsSomeAnd(func(s States) bool {
				return s.filled || opt == UndTagValueDef && s.Def || opt == UndTagValueNull && s.Null || opt == UndTagValueUnd && s.Und
			}) {
				return UndOpt{}, fmt.Errorf("%w: und tag contains multiple mutually exclusive options, tag = %s", ErrMultipleOption, org)
			}
		default:
			return UndOpt{}, ErrUnknownOption
		}

		opts.States = opts.States.Or(option.Some(States{})).Map(func(v States) States {
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
	}

	return opts, nil
}

func (o UndOpt) String() string {
	var builder strings.Builder

	appendStr := func(s fmt.Stringer) {
		ss := s.String()
		if builder.Len() > 0 && len(ss) > 0 {
			_, _ = builder.WriteString(", and ")
		}
		_, _ = builder.WriteString(ss)
	}

	if o.States.IsSome() {
		appendStr(o.States.Value())
	}
	if o.Len.IsSome() {
		appendStr(o.Len.Value())
	}
	if o.Values.IsSome() {
		appendStr(o.Values.Value())
	}

	return builder.String()
}

func (o UndOpt) ValidOpt(opt OptionLike) bool {
	return o.States.IsSomeAnd(func(s States) bool {
		switch {
		case opt.IsSome():
			return s.Def
		default: // opt.IsNone():
			return s.Null || s.Und
		}
	})
}

func (o UndOpt) ValidUnd(u UndLike) bool {
	return o.States.IsSomeAnd(func(s States) bool {
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
	return option.Option[struct{}]{}
}

func (o UndOpt) ValidElastic(e ElasticLike) bool {
	return option.MapOption(o.States, func(s States) bool {
		return s.Valid(e)
	}).Or(option.Some(false)).Value() || option.MapOption(or(o.Len, o.Values), func(_ struct{}) bool {
		return option.MapOption(o.Len, func(s LenValidator) bool { return s.Valid(e) }).Or(option.Some(true)).Value() &&
			option.MapOption(o.Values, func(s ValuesValidator) bool { return s.Valid(e) }).Or(option.Some(true)).Value()
	}).Or(option.Some(false)).Value()
}

func (o UndOpt) Ty() {

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
		v.Op = lenOpEqEq
	case s[:2] == ">=":
		v.Op = lenOpGrEq
	case s[:2] == "<=":
		v.Op = lenOpLeEq
	case s[0] == '<':
		v.Op = lenOpLe
	case s[0] == '>':
		v.Op = lenOpGr
	}

	s = s[v.Op.len():]

	len, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return LenValidator{}, fmt.Errorf("unknown len: %w", err)
	}

	v.Len = int(len)
	return v, nil
}

func (v LenValidator) String() string {
	return "must have length of " + v.Op.String() + " " + strconv.FormatInt(int64(v.Len), 10)
}

func (v LenValidator) Valid(e ElasticLike) bool {
	if v.Op == 0 {
		return true
	}
	return v.Op.Compare(e.Len(), v.Len)
}

type lenOp int

const (
	lenOpEqEq = iota + 1 // ==
	lenOpGr              // >
	lenOpGrEq            // >=
	lenOpLe              // <
	lenOpLeEq            // <=
)

func (o lenOp) len() int {
	switch o {
	case lenOpLe, lenOpGr:
		return 1
	case lenOpEqEq, lenOpGrEq, lenOpLeEq:
		return 2
	}
	return 0
}

func (o lenOp) String() string {
	switch o {
	default: // case lenOpEqEq:
		return "=="
	case lenOpGr:
		return ">"
	case lenOpGrEq:
		return ">="
	case lenOpLe:
		return "<"
	case lenOpLeEq:
		return "<="
	}
}

func (o lenOp) Compare(i, j int) bool {
	switch o {
	default: // case lenOpEqEq:
		return i == j
	case lenOpGr:
		return i > j
	case lenOpGrEq:
		return i >= j
	case lenOpLe:
		return i < j
	case lenOpLeEq:
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

func (v ValuesValidator) String() string {
	switch {
	case v.Nonnull:
		return "must not contain null"
	}
	return ""
}
