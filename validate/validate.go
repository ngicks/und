package validate

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/ngicks/und/internal/structtag"
)

var (
	// ErrNotStruct would be returned by ValidateUnd and CheckUnd
	// if input is not a struct nor a pointer to a struct.
	ErrNotStruct = structtag.ErrNotStruct
	// ErrMultipleOption would be returned by ValidateUnd and CheckUnd
	// if input's `und` struct tags have multiple mutually exclusive options.
	ErrMultipleOption = structtag.ErrMultipleOption
	// ErrUnknownOption is an error value which will be returned by ValidateUnd and CheckUnd
	// if an input has unknown options in `und` struct tag.
	ErrUnknownOption = structtag.ErrUnknownOption
	// ErrMalformedLen is an error which will be returned by ValidateUnd and CheckUnd
	// if an input has malformed len option in `und` struct tag.
	ErrMalformedLen = structtag.ErrMalformedLen
	// ErrMalformedLen is an error which will be returned by ValidateUnd and CheckUnd
	// if an input has malformed values option in `und` struct tag.
	ErrMalformedValues = structtag.ErrMalformedValues
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

type (
	ElasticLike = structtag.ElasticLike
	UndLike     = structtag.UndLike
	OptionLike  = structtag.OptionLike
)

var (
	elasticLike    = reflect.TypeFor[structtag.ElasticLike]()
	undLikeTy      = reflect.TypeFor[structtag.UndLike]()
	optionLikeTy   = reflect.TypeFor[structtag.OptionLike]()
	validatorUndTy = reflect.TypeFor[ValidatorUnd]()
	checkerUndTy   = reflect.TypeFor[CheckerUnd]()
)

// ValidateUnd validates whether s is compliant to the constraint placed by `und` struct tag.
//
// ValidateUnd only accepts struct or pointer to struct.
//
// Only fields whose struct tag contains `und`, and whose type is implementor of OptionLike, UndLike or ElasticLike, are validated.
func ValidateUnd(s any) error {
	rv := reflect.ValueOf(s)
	return cacheValidator(rv.Type()).validate(rv)
}

// CheckUnd checks whether s is correctly configured with `und` struct tag option without validating it.
func CheckUnd(s any) error {
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
			switch ftDeref.Kind() {
			default:
				continue
			case reflect.Pointer:
				ftDeref = ftDeref.Elem()
				if ftDeref.Kind() != reflect.Struct {
					continue
				}
			case reflect.Struct:
			}
			var (
				subFieldValidator *cachedValidator
				has               bool
			)
			if subFieldValidator, has = visited[ftDeref]; !has {
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
			}

			fieldValidators = append(fieldValidators, fieldValidator{
				i: i,
				validate: func(rv reflect.Value) error {
					err := subFieldValidator.validate(rv)
					if err != nil {
						return fmt.Errorf("%s.%w", ft.Name, err)
					}
					return nil
				},
			})

			continue
		}

		if ft.Type.Kind() == reflect.Pointer {
			return cachedValidator{rt: rt, err: fmt.Errorf("%s: pointer implementor field", ft.Name)}
		}

		tag := ft.Tag.Get(UndTag)
		if tag == "" {
			continue
		}
		opt, err := structtag.ParseOption(tag)
		if err != nil {
			return cachedValidator{rt: rt, err: fmt.Errorf("%s: %w", ft.Name, err)}
		}

		if !isElasticLike {
			if opt.Len.IsSome() {
				return cachedValidator{rt: rt, err: fmt.Errorf("%s: len on non elastic", ft.Name)}
			}
			if opt.Values.IsSome() {
				return cachedValidator{rt: rt, err: fmt.Errorf("%s: values on non elastic", ft.Name)}
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
				return fv.Interface().(ValidatorUnd).ValidateUnd()
			}
		}

		if ft.Type.Implements(checkerUndTy) {
			// keep it addressable. The type might implement it on pointer type.
			fv := reflect.New(ft.Type).Elem()
			err := fv.Interface().(CheckerUnd).CheckUnd()
			if err != nil {
				return cachedValidator{rt: rt, err: fmt.Errorf("%s.%w", ft.Name, err)}
			}
		}

		fieldValidators = append(fieldValidators, fieldValidator{
			i:        i,
			rt:       ft.Type,
			validate: validate,
		})
	}

	*mainValidator = cachedValidator{rt: rt, v: fieldValidators}
	return *mainValidator
}
