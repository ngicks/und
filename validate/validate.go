package validate

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
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
)

type undOpt struct {
	states optionLite[states]
	len    optionLite[lenValidator]
}

type states struct {
	filled bool
	def    bool
	null   bool
	und    bool
}

func (s states) String() string {
	if s.filled {
		if s.def {
			return "is " + UndTagValueRequired
		} else {
			return "is " + UndTagValueNullish
		}
	}
	var builder strings.Builder
	if s.def {
		builder.WriteString("defined")
	}
	if s.null {
		if builder.Len() > 0 {
			builder.WriteString(" or ")
		}
		builder.WriteString("null")
	}
	if s.und {
		if builder.Len() > 0 {
			builder.WriteString(" or ")
		}
		builder.WriteString("undefined")
	}
	return "must be " + builder.String()
}

func parseOption(s string) (undOpt, error) {
	org := s
	var (
		opt  string
		opts undOpt
	)
	for len(s) > 0 {
		opt, s, _ = strings.Cut(s, ",")
		if strings.HasPrefix(opt, UndTagValueLen) {
			if opts.len.IsSome() {
				return undOpt{}, ErrMultipleOption
			}
			lenV, err := parseLen(opt)
			if err != nil {
				return undOpt{}, fmt.Errorf("%w: %w", ErrMalformedLen, err)
			}
			opts.len = some(lenV)
			continue
		}

		switch opt {
		case UndTagValueRequired, UndTagValueNullish:
			if opts.states.IsSome() {
				return undOpt{}, fmt.Errorf("%w: und tag contains multiple mutually exclusive options, tag = %s", ErrMultipleOption, org)
			}
		case UndTagValueDef, UndTagValueNull, UndTagValueUnd:
			if opts.states.IsSomeAnd(func(s states) bool {
				return s.filled || opt == UndTagValueDef && s.def || opt == UndTagValueNull && s.null || opt == UndTagValueUnd && s.und
			}) {
				return undOpt{}, fmt.Errorf("%w: und tag contains multiple mutually exclusive options, tag = %s", ErrMultipleOption, org)
			}
		default:
			return undOpt{}, ErrUnknownOption
		}

		opts.states = opts.states.Or(some(states{})).Map(func(v states) states {
			switch opt {
			case UndTagValueRequired:
				v.filled = true
				v.def = true
			case UndTagValueNullish:
				v.filled = true
				v.null = true
				v.und = true
			case UndTagValueDef:
				v.def = true
			case UndTagValueNull:
				v.null = true
			case UndTagValueUnd:
				v.und = true
			}
			return v
		})
	}

	return opts, nil
}

func (o undOpt) String() string {
	if o.states.IsSome() {
		return o.states.Value().String()
	}
	return o.len.Value().String()
}

func (o undOpt) validOpt(opt OptionLike) bool {
	return o.states.IsSomeAnd(func(s states) bool {
		switch {
		case opt.IsSome():
			return s.def
		default: // opt.IsNone():
			return s.null || s.und
		}
	})
}

func (o undOpt) validUnd(u UndLike) bool {
	return o.states.IsSomeAnd(func(s states) bool {
		switch {
		case u.IsDefined():
			return s.def
		case u.IsNull():
			return s.null
		default: // case u.IsUndefined():
			return s.und
		}
	})
}

func (o undOpt) validElastic(e ElasticLike) bool {
	// states and len is mutually exclusive.
	return o.validUnd(e) || o.len.IsSomeAnd(func(lv lenValidator) bool {
		return lv.valid(e)
	})
}

type lenValidator struct {
	len int
	op  lenOp
}

func parseLen(s string) (lenValidator, error) {
	s, _ = strings.CutPrefix(s, UndTagValueLen)
	if len(s) < 2 { // <n, at least 2.
		return lenValidator{}, fmt.Errorf("unknown op: %s", s)
	}
	var v lenValidator
	switch {
	default:
		return lenValidator{}, fmt.Errorf("unknown op: %s", s)
	case s[:2] == "==":
		v.op = lenOpEqEq
	case s[:2] == ">=":
		v.op = lenOpGrEq
	case s[:2] == "<=":
		v.op = lenOpLeEq
	case s[0] == '<':
		v.op = lenOpLe
	case s[0] == '>':
		v.op = lenOpGr
	}

	s = s[v.op.len():]

	len, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return lenValidator{}, fmt.Errorf("unknown len: %w", err)
	}

	v.len = int(len)
	return v, nil
}

func (v lenValidator) String() string {
	return "length " + v.op.String() + " " + strconv.FormatInt(int64(v.len), 10)
}

func (v lenValidator) valid(e ElasticLike) bool {
	if v.op == 0 {
		return true
	}
	return v.op.compare(e.Len(), v.len)
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

func (o lenOp) compare(i, j int) bool {
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

type ElasticLike interface {
	UndLike
	Len() int
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

// ValidatorUnd wraps the ValidateUnd method.
//
// ValidateUnd method is implemented on data container types, und.Und[T] and option.Option[T], etc.
// It only validates its underlying T's compliance for constraints placed by `und` struct tag options.
type ValidatorUnd interface {
	ValidateUnd() error
}

// CheckerUnd wraps the CheckUnd method
// which is expected to be implemented on data container types like und.Und[T] and option.Option[T], etc.
//
// CheckUnd must checks if its internal data type conforms the constraint which ValidateUnd or CheckUnd would checks.
type CheckerUnd interface {
	CheckUnd() error
}

var (
	elasticLike  = reflect.TypeFor[ElasticLike]()
	undLikeTy    = reflect.TypeFor[UndLike]()
	optionLikeTy = reflect.TypeFor[OptionLike]()
	checkerOptTy = reflect.TypeFor[CheckerUnd]()
)

// ValidateUnd validates whether s is compliant to the constraint placed by `und` struct tag.
//
// ValidateUnd only accepts struct or pointer to struct.
//
// Only fields whose struct tag contains `und`, and whose type is implementor of OptionLike, UndLike or ElasticLike, are validated.
func ValidateUnd(s any) error {
	rv := reflect.ValueOf(s)
	return check(rv, rv.Type(), true)
}

// CheckUnd checks whether s is correctly configured with `und` struct tag option without validating it.
func CheckUnd(s any) error {
	rv := reflect.ValueOf(s)
	return check(rv, rv.Type(), false)
}

func check(rv reflect.Value, rt reflect.Type, doValidate bool) error {
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
		if doValidate {
			rv = rv.Elem()
		}
	}
	//TODO: warn or return error if rv is non addressable value? Validate method might be implemented on pointer receiver.
	if rt.Kind() != reflect.Struct {
		return fmt.Errorf("%w: input is expected to be a struct type but is %s", ErrNotStruct, rv.Type())
	}

	for i := 0; i < rt.NumField(); i++ {
		var fv reflect.Value
		if doValidate {
			fv = rv.Field(i)
		}
		ft := rt.Field(i)

		if !ft.IsExported() {
			continue
		}

		isElasticLike := ft.Type.Implements(elasticLike)
		isUndLike := ft.Type.Implements(undLikeTy)
		isOptLike := ft.Type.Implements(optionLikeTy)
		if !isElasticLike && !isUndLike && !isOptLike {
			switch ft.Type.Kind() {
			default:
				continue
			case reflect.Pointer:
				if ft.Type.Elem().Kind() != reflect.Struct {
					continue
				}
			case reflect.Struct:
			}
			err := check(fv, ft.Type, doValidate)
			if err != nil {
				return fmt.Errorf("%s.%w", ft.Name, err)
			}
			continue
		}

		if ft.Type.Kind() == reflect.Pointer {
			return fmt.Errorf("%s: pointer implementor field", ft.Name)
		}

		tag := ft.Tag.Get(UndTag)
		if tag == "" {
			continue
		}
		opt, err := parseOption(tag)
		if err != nil {
			return fmt.Errorf("%s: %w", ft.Name, err)
		}

		if !isElasticLike && opt.len.IsSome() {
			return fmt.Errorf("%s: len on non elastic", ft.Name)
		}

		if doValidate {
			switch {
			case isElasticLike:
				if !opt.validElastic(fv.Interface().(ElasticLike)) {
					return fmt.Errorf("%s: input %s", ft.Name, opt) //TODO
				}
			case isUndLike:
				if !opt.validUnd(fv.Interface().(UndLike)) {
					return fmt.Errorf("%s: input %s", ft.Name, opt) //TODO
				}
			case isOptLike:
				if !opt.validOpt(fv.Interface().(OptionLike)) {
					return fmt.Errorf("%s: input %s", ft.Name, opt) //TODO
				}
			}

			if v, ok := fv.Interface().(ValidatorUnd); ok {
				err := v.ValidateUnd()
				if err != nil {
					return fmt.Errorf("%s.%w", ft.Name, err)
				}
			}
		} else if ft.Type.Implements(checkerOptTy) {
			// keep it addressable. The type might implement it on pointer type.
			fv := reflect.New(ft.Type).Elem()
			err := fv.Interface().(CheckerUnd).CheckUnd()
			if err != nil {
				return fmt.Errorf("%s.%w", ft.Name, err)
			}
		}
	}

	return nil
}
