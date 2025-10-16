package ginbinding

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

var (
	ginCtxTy   = reflect.TypeOf(gin.Context{})
	errTy      = reflect.TypeOf((*error)(nil)).Elem()
	strTy      = reflect.TypeOf("")
	timeTy     = reflect.TypeOf(time.Time{})
	durationTy = reflect.TypeOf(time.Duration(0))
)

// BasicFormBindingGinHandlerBuilder is the basic implementation of FormBindingGinHandlerBuilder
// that supports validation and customizable response handling.
type BasicFormBindingGinHandlerBuilder struct {
	validator       binding.StructValidator
	responseHandler ResponseHandler
}

// NewBasicFormBindingGinHandlerBuilder creates a new builder with optional validator and response handler
func NewBasicFormBindingGinHandlerBuilder(
	validator binding.StructValidator,
	responseHandler ResponseHandler,
) *BasicFormBindingGinHandlerBuilder {
	if responseHandler == nil {
		responseHandler = NewDefaultResponseHandler()
	}
	return &BasicFormBindingGinHandlerBuilder{
		validator:       validator,
		responseHandler: responseHandler,
	}
}

// FormBindingGinHandlerFunc converts a function to a gin.HandlerFunc
// Supported function signatures:
//  1. func(*gin.Context, any struct) error
//  2. func(*gin.Context, any struct) (any, error)
//  3. func(*gin.Context) (any, error)
func (builder *BasicFormBindingGinHandlerBuilder) FormBindingGinHandlerFunc(
	i any,
) (gin.HandlerFunc, error) {
	ity := reflect.TypeOf(i)

	if ity.Kind() != reflect.Func {
		return nil, errors.New("input must be a function")
	}

	// Check parameter and return value counts
	inNum := ity.NumIn()
	outNum := ity.NumOut()

	if inNum == 0 {
		return nil, errors.New("function must have at least one parameter")
	}

	if inNum > 2 {
		return nil, errors.New("function can have at most 2 parameters")
	}

	if outNum == 0 {
		return nil, errors.New("function must have at least one return value")
	}

	if outNum > 2 {
		return nil, errors.New("function can have at most 2 return values")
	}

	// Check first parameter is *gin.Context
	in0Ty := ity.In(0)
	if in0Ty.Kind() != reflect.Pointer || in0Ty.Elem() != ginCtxTy {
		return nil, errors.New("first parameter must be *gin.Context")
	}

	// If function has second parameter, it must be a struct or pointer to struct
	if inNum == 2 {
		in1Ty := ity.In(1)
		if in1Ty.Kind() != reflect.Struct &&
			(in1Ty.Kind() != reflect.Pointer || in1Ty.Elem().Kind() != reflect.Struct) {
			return nil, errors.New("second parameter must be a struct or pointer to struct")
		}
	}

	// Check return value types
	if outNum == 1 {
		out0Ty := ity.Out(0)
		if !out0Ty.Implements(errTy) {
			return nil, errors.New("single return value must be error")
		}
	}

	if outNum == 2 {
		out1Ty := ity.Out(1)
		if !out1Ty.Implements(errTy) {
			return nil, errors.New("second return value must be error")
		}
	}

	funcVal := reflect.ValueOf(i)

	return func(ctx *gin.Context) {
		in := make([]reflect.Value, 0, 2)
		in = append(in, reflect.ValueOf(ctx))

		if inNum == 2 {
			form, err := bindingFormValue(ctx, ity.In(1))
			if err != nil {
				builder.responseHandler.HandleError(ctx, &BindingError{Err: err})
				return
			}

			if builder.validator != nil {
				if err := builder.validator.ValidateStruct(form.Interface()); err != nil {
					builder.responseHandler.HandleError(ctx, err)
					return
				}
			}

			in = append(in, form)
		}

		out := funcVal.Call(in)

		if outNum == 1 {
			err := out[0].Interface()
			if err != nil {
				builder.responseHandler.HandleError(ctx, err.(error))
				return
			}
			builder.responseHandler.HandleSuccess(ctx, nil)
			return
		}

		err := out[1].Interface()
		if err != nil {
			builder.responseHandler.HandleError(ctx, err.(error))
			return
		}

		builder.responseHandler.HandleSuccess(ctx, out[0].Interface())
	}, nil
}

func bindingFormValue(ctx *gin.Context, ty reflect.Type) (reflect.Value, error) {
	if ty.Kind() == reflect.Pointer {
		val, err := bindingFormValue(ctx, ty.Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		ret := reflect.New(ty.Elem())
		ret.Elem().Set(val)
		return ret, nil
	}

	val := reflect.New(ty)

	headerTagsNum := 0
	formTagsNum := 0

	for i := 0; i < ty.NumField(); i++ {
		sf := ty.Field(i)

		if !sf.IsExported() {
			continue
		}

		if pathKey, ok := sf.Tag.Lookup("path"); ok {
			sfv, err := stringToVal(ctx.Param(pathKey), sf.Type)
			if err != nil {
				return val.Elem(), fmt.Errorf("failed to parse path parameter %q: %w", pathKey, err)
			}
			val.Elem().Field(i).Set(sfv)
		}

		if _, ok := sf.Tag.Lookup("header"); ok {
			headerTagsNum += 1
		}

		if _, ok := sf.Tag.Lookup("form"); ok {
			formTagsNum += 1
		}
	}

	if formTagsNum > 0 {
		if err := ctx.BindQuery(val.Interface()); err != nil {
			return val.Elem(), err
		}
	}

	if headerTagsNum > 0 {
		if err := ctx.ShouldBindHeader(val.Interface()); err != nil {
			return val.Elem(), err
		}
	}

	err := ctx.ShouldBind(val.Interface())

	// Apply default values for zero-valued fields
	if err == nil {
		if defaultErr := applyDefaultValues(val.Elem()); defaultErr != nil {
			return val.Elem(), defaultErr
		}
	}

	return val.Elem(), err
}

func stringToVal(s string, ty reflect.Type) (reflect.Value, error) {
	if s == "" {
		return reflect.Zero(ty), nil
	}

	if strTy.ConvertibleTo(ty) {
		return reflect.ValueOf(s).Convert(ty), nil
	}

	ret := reflect.New(ty)

	switch ty.Kind() {
	case reflect.String:
		ret.Elem().Set(reflect.ValueOf(s))
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		// Handle time.Duration specially
		if ty == durationTy {
			d, err := time.ParseDuration(s)
			if err != nil {
				return reflect.Zero(ty), fmt.Errorf("invalid duration %q: %w", s, err)
			}
			ret.Elem().Set(reflect.ValueOf(d))
		} else {
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return reflect.Zero(ty), err
			}
			ret.Elem().SetInt(i)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return reflect.Zero(ty), err
		}
		ret.Elem().SetUint(i)
	case reflect.Bool:
		b, err := parseBool(s)
		if err != nil {
			return reflect.Zero(ty), err
		}
		ret.Elem().SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return reflect.Zero(ty), err
		}
		ret.Elem().SetFloat(f)
	default:
		// Handle time.Time types
		if ty == timeTy {
			// Try multiple time formats
			timeFormats := []string{
				time.RFC3339,
				time.RFC3339Nano,
				"2006-01-02 15:04:05",
				"2006-01-02T15:04:05Z",
				"2006-01-02T15:04:05.000Z",
				"2006-01-02",
				"15:04:05",
			}

			var parsedTime time.Time
			var parseErr error

			for _, format := range timeFormats {
				parsedTime, parseErr = time.Parse(format, s)
				if parseErr == nil {
					break
				}
			}

			if parseErr != nil {
				return reflect.Zero(ty), fmt.Errorf("invalid time format %q: %w", s, parseErr)
			}

			ret.Elem().Set(reflect.ValueOf(parsedTime))
		} else {
			return reflect.Zero(ty), fmt.Errorf("unsupported type conversion from %q to %s", s, ty)
		}
	}

	return ret.Elem(), nil
}

// applyDefaultValues applies default values to zero-valued fields that have a "default" tag
func applyDefaultValues(val reflect.Value) error {
	ty := val.Type()

	for i := 0; i < ty.NumField(); i++ {
		sf := ty.Field(i)

		if !sf.IsExported() {
			continue
		}

		fieldVal := val.Field(i)

		// Handle embedded structs (anonymous fields)
		if sf.Anonymous {
			// Handle pointer-type embedded structs
			if fieldVal.Kind() == reflect.Ptr {
				if fieldVal.IsNil() {
					// Pointer is nil, skip processing
					continue
				}
				// Dereference pointer
				fieldVal = fieldVal.Elem()
			}

			// Recursively process embedded struct fields
			if err := applyDefaultValues(fieldVal); err != nil {
				return fmt.Errorf("embedded struct %s: %w", sf.Name, err)
			}
			continue
		}

		// Handle default values for regular fields
		defaultValue, hasDefault := sf.Tag.Lookup("default")
		if !hasDefault {
			continue
		}

		// Check if field is zero value
		// We only apply defaults to truly zero values
		if !fieldVal.IsZero() {
			continue
		}

		// Convert and set default value based on field type
		if err := setDefaultValue(fieldVal, defaultValue, sf.Name); err != nil {
			return fmt.Errorf("field %s: %w", sf.Name, err)
		}
	}

	return nil
}

// setDefaultValue converts a string default value to the appropriate type and sets it
func setDefaultValue(fieldVal reflect.Value, defaultValue string, fieldName string) error {
	// Handle pointer types
	if fieldVal.Kind() == reflect.Ptr {
		if fieldVal.IsNil() {
			// Create new instance of the underlying type
			elemType := fieldVal.Type().Elem()
			newVal := reflect.New(elemType)

			// Set the default value on the new instance
			if err := setDefaultValue(newVal.Elem(), defaultValue, fieldName); err != nil {
				return err
			}

			fieldVal.Set(newVal)
		}
		return nil
	}

	// Use stringToVal to convert the default value to the field type
	convertedVal, err := stringToVal(defaultValue, fieldVal.Type())
	if err != nil {
		return fmt.Errorf("failed to convert default value %q for field %s: %w", defaultValue, fieldName, err)
	}

	fieldVal.Set(convertedVal)
	return nil
}

// parseBool parses a string to boolean value
func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", s)
	}
}
