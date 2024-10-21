package validate

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/ngicks/und/undtag"
)

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
						for i, v := range fv.Seq2() {
							if err := validator(v); err != nil {
								return fmt.Errorf("[%v].%w", i.Interface(), err)
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
						return fmt.Errorf("%s.%w", ft.Name, err)
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
		return false, nil, fmt.Errorf("%s: pointer implementor field", ft.Name)
	}

	tag := ft.Tag.Get(undtag.TagName)
	if tag == "" {
		return false, nil, nil
	}
	opt, err := undtag.ParseOption(tag)
	if err != nil {
		return true, nil, fmt.Errorf("%s: %w", ft.Name, err)
	}

	if !isElasticLike {
		if opt.Len().IsSome() {
			return true, nil, fmt.Errorf("%s: len on non elastic", ft.Name)
		}
		if opt.Values().IsSome() {
			return true, nil, fmt.Errorf("%s: values on non elastic", ft.Name)
		}
	}

	var validateOpt func(fv reflect.Value) error
	switch {
	case isElasticLike:
		validateOpt = func(fv reflect.Value) error {
			if !opt.ValidElastic(fv.Interface().(ElasticLike)) {
				return fmt.Errorf("%s: input %s", ft.Name, opt)
			}
			return nil
		}
	case isUndLike:
		validateOpt = func(fv reflect.Value) error {
			if !opt.ValidUnd(fv.Interface().(UndLike)) {
				return fmt.Errorf("%s: input %s", ft.Name, opt)
			}
			return nil
		}
	case isOptLike:
		validateOpt = func(fv reflect.Value) error {
			if !opt.ValidOpt(fv.Interface().(OptionLike)) {
				return fmt.Errorf("%s: input %s", ft.Name, opt)
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
			return true, nil, fmt.Errorf("%s.%w", ft.Name, err)
		}
	}
	return true, validate, nil
}
