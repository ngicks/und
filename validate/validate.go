package validate

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/ngicks/und/undtag"
)

type fieldSelectorType int

const (
	fieldSelectorTypeDot = fieldSelectorType(iota)
	fieldSelectorTypeIndex
)

type fieldSelector struct {
	ty       fieldSelectorType
	selector string
}

type ValidationError struct {
	fieldChain []fieldSelector
	err        error
}

func ReportState(v any) string {
	if i, ok := v.(ElasticLike); ok {
		switch {
		case i.IsUndefined():
			return "undefined"
		case i.IsNull():
			return "null"
		default:
			return fmt.Sprintf("defined, len=%d, has null=%t", i.Len(), i.HasNull())
		}
	}
	if i, ok := v.(UndLike); ok {
		switch {
		case i.IsUndefined():
			return "undefined"
		case i.IsNull():
			return "null"
		default:
			return "defined"
		}
	}
	if i, ok := v.(OptionLike); ok {
		if i.IsSome() {
			return "some"
		} else {
			return "none"
		}
	}

	return ""
}

func NewValidationError(err error) *ValidationError {
	return &ValidationError{err: err}
}

func AppendValidationErrorDot(err error, selector string) error {
	vErr, ok := err.(*ValidationError)
	if !ok {
		return &ValidationError{err: err, fieldChain: []fieldSelector{{fieldSelectorTypeDot, selector}}}
	}
	vErr.fieldChain = append(vErr.fieldChain, fieldSelector{fieldSelectorTypeDot, selector})
	return vErr
}

func AppendValidationErrorIndex(err error, selector string) error {
	vErr, ok := err.(*ValidationError)
	if !ok {
		return &ValidationError{err: err, fieldChain: []fieldSelector{{fieldSelectorTypeIndex, selector}}}
	}
	vErr.fieldChain = append(vErr.fieldChain, fieldSelector{fieldSelectorTypeIndex, selector})
	return vErr
}

func (e *ValidationError) Unwrap() error {
	return e.err
}

func (e *ValidationError) Error() string {
	var builder strings.Builder
	builder.WriteString("validation failed at ")
	for _, f := range slices.Backward(e.fieldChain) {
		switch f.ty {
		case fieldSelectorTypeDot:
			builder.WriteByte('.')
			builder.WriteString(f.selector)
		case fieldSelectorTypeIndex:
			builder.WriteByte('[')
			builder.WriteString(f.selector)
			builder.WriteByte(']')
		}
	}
	builder.WriteString(": ")
	builder.WriteString(e.err.Error())
	return builder.String()
}

// Pointer returns rfc6901 compliant json pointer
func (e *ValidationError) Pointer() string {
	var builder strings.Builder
	for _, f := range slices.Backward(e.fieldChain) {
		builder.WriteByte('/')
		sel := f.selector
		sel = strings.ReplaceAll(sel, "~", "~0")
		sel = strings.ReplaceAll(sel, "/", "~1")
		builder.WriteString(sel)
	}
	return builder.String()
}

var (
	// ErrNotStruct would be returned by UndValidate and UndCheck
	// if input is not a struct nor a pointer to a struct.
	ErrNotStruct = errors.New("not struct")
)

var (
	// ErrMultipleOption would be returned by UndValidate and UndCheck
	// if input's `und` struct tags have multiple mutually exclusive options.
	ErrMultipleOption = undtag.ErrMultipleOption
	// ErrUnknownOption is an error value which will be returned by UndValidate and UndCheck
	// if an input has unknown options in `und` struct tag.
	ErrUnknownOption = undtag.ErrUnknownOption
	// ErrMalformedLen is an error which will be returned by UndValidate and UndCheck
	// if an input has malformed len option in `und` struct tag.
	ErrMalformedLen = undtag.ErrMalformedLen
	// ErrMalformedLen is an error which will be returned by UndValidate and UndCheck
	// if an input has malformed values option in `und` struct tag.
	ErrMalformedValues = undtag.ErrMalformedValues
)

// UndValidator wraps the UndValidate method.
//
// UndValidate method is implemented on data container types, und.Und[T] and option.Option[T], etc.
// It only validates its underlying T's compliance for constraints placed by `und` struct tag options.
type UndValidator interface {
	UndValidate() error
}

// UndChecker wraps the UndCheck method
// which is expected to be implemented on data container types like und.Und[T] and option.Option[T], etc.
//
// UndCheck must checks if its internal data type conforms the constraint which UndValidate or UndCheck would checks.
type UndChecker interface {
	UndCheck() error
}

type (
	ElasticLike = undtag.ElasticLike
	UndLike     = undtag.UndLike
	OptionLike  = undtag.OptionLike
)

var (
	elasticLike    = reflect.TypeFor[undtag.ElasticLike]()
	undLikeTy      = reflect.TypeFor[undtag.UndLike]()
	optionLikeTy   = reflect.TypeFor[undtag.OptionLike]()
	validatorUndTy = reflect.TypeFor[UndValidator]()
	checkerUndTy   = reflect.TypeFor[UndChecker]()
)

// UndValidate validates whether s is compliant to the constraint placed by `und` struct tag.
//
// UndValidate only accepts struct or pointer to struct.
//
// Only fields whose struct tag contains `und`, and whose type is implementor of OptionLike, UndLike, ElasticLike,
// or array, slice, map whose value type are one of implementor,
// are validated.
func UndValidate(s any) error {
	rv := reflect.ValueOf(s)
	return cacheValidator(rv.Type()).validate(rv)
}

// UndCheck checks whether s is correctly configured with `und` struct tag option without validating it.
func UndCheck(s any) error {
	return cacheValidator(reflect.TypeOf(s)).check()
}

var validatorCache sync.Map

type cachedValidator struct {
	rt  reflect.Type
	err error
	v   []fieldValidator
}

func (v cachedValidator) validate(rv reflect.Value) error {
	if v.err != nil {
		return v.err
	}
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			// no further stepping
			return nil
		}
		rv = rv.Elem()
	}
	for _, f := range v.v {
		if err := f.validate(rv.Field(f.i)); err != nil {
			return err
		}
	}
	return nil
}

func (v cachedValidator) check() error {
	if v.err != nil {
		return v.err
	}
	return nil
}

type fieldValidator struct {
	i        int
	rt       reflect.Type
	validate func(fv reflect.Value) error
}

func cacheValidator(rt reflect.Type) cachedValidator {
	v, ok := validatorCache.Load(rt)
	if !ok {
		v, _ = validatorCache.LoadOrStore(rt, makeValidator(rt, nil))
	}
	return v.(cachedValidator)
}

func makeValidator(rt reflect.Type, visited map[reflect.Type]*cachedValidator) cachedValidator {
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	//TODO: warn or return error if rv is non addressable value? Validate method might be implemented on pointer receiver.
	if rt.Kind() != reflect.Struct {
		return cachedValidator{rt: rt, err: fmt.Errorf("%w: input is expected to be a struct type but is %s", ErrNotStruct, rt.Kind())}
	}

	if visited == nil {
		visited = make(map[reflect.Type]*cachedValidator)
	}
	mainValidator := &cachedValidator{}
	visited[rt] = mainValidator

	var fieldValidators []fieldValidator
	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)

		if !ft.IsExported() {
			continue
		}

		isElasticLike := ft.Type.Implements(elasticLike)
		isUndLike := ft.Type.Implements(undLikeTy)
		isOptLike := ft.Type.Implements(optionLikeTy)
		if !isElasticLike && !isUndLike && !isOptLike {
			ftDeref := ft.Type

			if ftDeref.Kind() == reflect.Pointer {
				ftDeref = ftDeref.Elem()
				if ftDeref.Kind() != reflect.Struct {
					// maybe we'll loosen this restriction
					// then we could allow *map[K]V or *[]T fields to be validated.
					continue
				}
			}

			subFieldValidator, has := visited[ftDeref]
			var validateField func(fv reflect.Value) error
			if !has {
				switch ftDeref.Kind() {
				default:
					continue
				case reflect.Struct:
					v := makeValidator(ft.Type, visited)
					if v.err != nil {
						return cachedValidator{
							rt: rt,
							v: []fieldValidator{{
								i:        i,
								validate: v.validate,
							}},
						}
					}
					subFieldValidator = &v
				case reflect.Array, reflect.Slice, reflect.Map:
					elem := ftDeref.Elem()
					isElasticLike := elem.Implements(elasticLike)
					isUndLike := elem.Implements(undLikeTy)
					isOptLike := elem.Implements(optionLikeTy)
					hasTag, validator, err := makeFieldValidator(ft, isOptLike, isUndLike, isElasticLike)
					if !hasTag {
						continue
					}
					if err != nil {
						return cachedValidator{rt: rt, err: err}
					}
					validateField = func(fv reflect.Value) error {
						for k, v := range fv.Seq2() {
							if err := validator(v); err != nil {
								return AppendValidationErrorIndex(err, fmt.Sprintf("%v", k.Interface()))
							}
						}
						return nil
					}
				}
			}

			if validateField == nil {
				validateField = func(fv reflect.Value) error {
					err := subFieldValidator.validate(fv)
					if err != nil {
						return AppendValidationErrorDot(err, ft.Name)
					}
					return nil
				}
			}
			fieldValidators = append(fieldValidators, fieldValidator{
				i:        i,
				validate: validateField,
			})

			continue
		}
		hasTag, validator, err := makeFieldValidator(ft, isOptLike, isUndLike, isElasticLike)
		if !hasTag {
			continue
		}
		if err != nil {
			return cachedValidator{rt: rt, err: err}
		}
		fieldValidators = append(
			fieldValidators,
			fieldValidator{
				i:        i,
				rt:       rt,
				validate: validator,
			},
		)
	}

	*mainValidator = cachedValidator{rt: rt, v: fieldValidators}
	return *mainValidator
}

func makeFieldValidator(ft reflect.StructField, isOptLike, isUndLike, isElasticLike bool) (hasTag bool, validator func(fv reflect.Value) error, err error) {
	if ft.Type.Kind() == reflect.Pointer {
		// When field is nil, what should we do? It it considered none or undefined?
		// I don't have any idea on this. Just return an error.
		return false, nil, AppendValidationErrorDot(fmt.Errorf("pointer implementor field"), ft.Name)
	}

	tag := ft.Tag.Get(undtag.TagName)
	if tag == "" {
		return false, nil, nil
	}
	opt, err := undtag.ParseOption(tag)
	if err != nil {
		return true, nil, AppendValidationErrorDot(err, ft.Name)
	}

	if !isElasticLike {
		if opt.Len().IsSome() {
			return true, nil, AppendValidationErrorDot(fmt.Errorf("len on non elastic"), ft.Name)
		}
		if opt.Values().IsSome() {
			return true, nil, AppendValidationErrorDot(fmt.Errorf("values on non elastic"), ft.Name)
		}
	}

	var validateOpt func(fv reflect.Value) error
	switch {
	case isElasticLike:
		validateOpt = func(fv reflect.Value) error {
			if !opt.ValidElastic(fv.Interface().(ElasticLike)) {
				return AppendValidationErrorDot(fmt.Errorf("input %s", opt.Describe()), ft.Name)
			}
			return nil
		}
	case isUndLike:
		validateOpt = func(fv reflect.Value) error {
			if !opt.ValidUnd(fv.Interface().(UndLike)) {
				return AppendValidationErrorDot(fmt.Errorf("input %s", opt.Describe()), ft.Name)
			}
			return nil
		}
	case isOptLike:
		validateOpt = func(fv reflect.Value) error {
			if !opt.ValidOpt(fv.Interface().(OptionLike)) {
				return AppendValidationErrorDot(fmt.Errorf("input %s", opt.Describe()), ft.Name)
			}
			return nil
		}
	}

	validate := validateOpt
	if ft.Type.Implements(validatorUndTy) {
		validate = func(fv reflect.Value) error {
			err := validateOpt(fv)
			if err != nil {
				return err
			}
			return fv.Interface().(UndValidator).UndValidate()
		}
	}

	if ft.Type.Implements(checkerUndTy) {
		// keep it addressable. The type might implement it on pointer type.
		fv := reflect.New(ft.Type).Elem()
		err := fv.Interface().(UndChecker).UndCheck()
		if err != nil {
			return true, nil, AppendValidationErrorDot(err, ft.Name)
		}
	}
	return true, validate, nil
}
