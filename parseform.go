package forms

import (
	"errors"
	"net/http"
	"reflect"
	"slices"
	"strings"
)

const (
	fieldIsValid fieldErrorType = iota
	fieldUnaddressible
	fieldHasUnsupportedType
	fieldIsRequired
	fieldMustNotBeEmpty
	fieldRequiredWithDefault
	fieldParsingError
)

var (
	ErrNeedStructType      = errors.New("type parameter must be a struct")
	ErrNeedPointerToStruct = errors.New("dest must be a non-nil pointer to a struct")
)

type fieldErrorType int

type fieldError struct {
	structField string
	val         any
	msg         string
	errorType   fieldErrorType
}

func (fe *fieldError) Error() string {
	errStr := "field "
	if fe.structField != "" {
		errStr += fe.structField + " "
	}
	errStr += "error: " + fe.msg
	return errStr
}

func FillStructFromForm(req *http.Request, dest any) error {
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Pointer || dest == nil {
		return ErrNeedPointerToStruct
	}
	destVal = destVal.Elem()

	err := req.ParseForm()
	if err != nil {
		return err
	}

	fields := reflect.VisibleFields(destVal.Type())
	method := strings.ToUpper(req.Method)
	for _, field := range fields {
		if tagMethodVal, ok := field.Tag.Lookup("method"); ok {
			methods := strings.Split(strings.ToUpper(tagMethodVal), ",")
			if !slices.Contains(methods, method) {
				// ex: `method:"POST"` not filled for HTTP GET requests
				continue
			}
		}
		fieldKind := field.Type.Kind()
		if fieldKind == reflect.Struct {
			// nested struct, possibly composite
			fieldVal := destVal.FieldByName(field.Name)
			if !fieldVal.CanAddr() {
				return &fieldError{structField: field.Name, msg: "unaddressible field", errorType: fieldUnaddressible}
			}
			if err = FillStructFromForm(req, fieldVal.Addr().Interface()); err != nil {
				return err
			}
			continue
		}
		if fieldKind == reflect.Pointer {
			// nested pointer to struct, possibly composite
			if field.Type.Elem().Kind() == reflect.Struct {
				fieldVal := destVal.FieldByName(field.Name)
				if !fieldVal.CanAddr() {
					return &fieldError{structField: field.Name, msg: "unaddressible field", errorType: fieldUnaddressible}
				}
				if fieldVal.IsNil() {
					fieldVal.Set(reflect.New(field.Type.Elem()))
				}
				if err = FillStructFromForm(req, fieldVal.Interface()); err != nil {
					return err
				}
				continue
			}
		}
		if fieldKind == reflect.Interface {
			return &fieldError{structField: field.Name, msg: "invalid field type, must be string, int, bool, a struct, or pointer to a struct with these fields", errorType: fieldHasUnsupportedType}
		}

		fieldOptions, _ := parseFieldOptions(req, field)
		if fieldOptions.skip() {
			continue
		}
		if fieldOptions.required && fieldOptions.hasDefault {
			return &fieldError{structField: field.Name, msg: "field has mutually exclusive required and default options", errorType: fieldRequiredWithDefault}
		}
		if fieldOptions.formVals == nil {
			if fieldOptions.required {
				return &fieldError{structField: field.Name, msg: "required field not in form", errorType: fieldIsRequired}
			}
			fieldOptions.formVals = []string{""}
		} else if fieldOptions.notEmpty && fieldOptions.formVals[0] == "" {
			return &fieldError{structField: field.Name, msg: "form field must not be empty", errorType: fieldMustNotBeEmpty}
		}
		fieldVal := destVal.FieldByName(field.Name)
		if err = fillFieldValue(field.Name, fieldVal, fieldOptions.formVals); err != nil {
			return err
		}
	}
	return nil
}

func GetStruct[T any](req *http.Request) (*T, error) {
	if reflect.TypeFor[T]().Kind() != reflect.Struct {
		return nil, ErrNeedStructType
	}
	var dest T
	err := FillStructFromForm(req, &dest)
	if err != nil {
		return nil, err
	}
	return &dest, nil
}
