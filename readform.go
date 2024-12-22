package formstructs

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strconv"
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
	ErrNeedStructType = errors.New("type parameter must be a struct")
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

func requestHasField(req *http.Request, fieldName string, method string) ([]string, bool) {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		req.PostFormValue("_") // parse the form with the default max memory
		return req.PostForm[fieldName], len(req.PostForm[fieldName]) > 0
	default:
		req.FormValue(fieldName) // parse the form with the default max memory
		return req.Form[fieldName], len(req.PostForm[fieldName]) > 0
	}
}

func setFieldVal(name string, val reflect.Value, data string) error {
	if !val.CanAddr() {
		return &fieldError{structField: name, val: data, msg: "unaddressable value", errorType: fieldUnaddressible}
	}
	switch val.Kind() {
	case reflect.Bool:
		val.SetBool(data == "1" || data == "on")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if data == "" {
			data = "0"
		}
		intVal, err := strconv.Atoi(data)
		if err != nil {
			return &fieldError{structField: name, val: data, msg: err.Error(), errorType: fieldParsingError}
		}
		val.SetInt(int64(intVal))
	case reflect.String:
		val.SetString(data)
	case reflect.Ptr, reflect.Interface:
		return setFieldVal(name, val.Elem(), data)
	default:
		return &fieldError{msg: fmt.Sprint("unsupported slice kind ", val.Kind().String()), errorType: fieldHasUnsupportedType}
	}
	return nil
}

func fillFieldValue(name string, structFieldVal reflect.Value, formFieldData []string) (err error) {
	switch structFieldVal.Kind() {
	case reflect.Slice:
		newSlice := reflect.MakeSlice(structFieldVal.Type(), len(formFieldData), len(formFieldData))
		for i, data := range formFieldData {
			if err = setFieldVal(name, newSlice.Index(i), data); err != nil {
				return err
			}
		}
		structFieldVal.Set(newSlice)
	case reflect.Bool:
		structFieldVal.SetBool(len(formFieldData) == 1 && (formFieldData[0] == "1" || formFieldData[0] == "on"))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.String:
		return setFieldVal(name, structFieldVal, formFieldData[0])
	case reflect.Ptr, reflect.Interface:
		return setFieldVal(name, structFieldVal.Elem(), formFieldData[0])
	default:
		return &fieldError{structField: name, msg: "unsupported field type", errorType: fieldHasUnsupportedType}
	}

	return nil
}

func FillStructFromForm(req *http.Request, dest any) error {
	return errors.New("not yet implemented")
}

func GetStruct[T any](req *http.Request) (*T, error) {
	destType := reflect.TypeFor[T]()
	destKind := destType.Kind()
	if destKind != reflect.Struct {
		return nil, ErrNeedStructType
	}

	var dest T
	destVal := reflect.ValueOf(&dest).Elem()

	err := req.ParseForm()
	if err != nil {
		return nil, err
	}

	fields := reflect.VisibleFields(destType)
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
		if fieldKind == reflect.Pointer || fieldKind == reflect.Interface || fieldKind == reflect.Struct {
			return nil, &fieldError{structField: field.Name, msg: "invalid field type, must be string, int, or bool", errorType: fieldHasUnsupportedType}
		}

		formNameOptionsParts := strings.Split(field.Tag.Get("form"), ",")
		formFieldName := formNameOptionsParts[0]
		if formFieldName == "" {
			formFieldName = field.Name
		} else if formFieldName == "-" {
			continue
		}
		formFieldVals, ok := requestHasField(req, formFieldName, method)
		var required bool
		var hasDefault bool
		var notEmpty bool
		for _, nameOptionPart := range formNameOptionsParts[1:] {
			// handle field options (required and default)
			if nameOptionPart == "required" {
				required = true
			} else if nameOptionPart == "notempty" && fieldKind == reflect.String {
				notEmpty = true
			} else if strings.HasPrefix(nameOptionPart, "default=") {
				hasDefault = true
				if !ok {
					formFieldVals = []string{nameOptionPart[8:]}
				}
			}
		}
		if required && hasDefault {
			return nil, &fieldError{structField: field.Name, msg: "field has mutually exclusive required and default options", errorType: fieldRequiredWithDefault}
		}
		if formFieldVals == nil {
			if required {
				return nil, &fieldError{structField: field.Name, msg: "required field not in form", errorType: fieldIsRequired}
			}
			formFieldVals = []string{""}
		} else if notEmpty && formFieldVals[0] == "" {
			return nil, &fieldError{structField: field.Name, msg: "form field must not be empty", errorType: fieldMustNotBeEmpty}
		}
		fieldVal := destVal.FieldByName(field.Name)
		if err = fillFieldValue(field.Name, fieldVal, formFieldVals); err != nil {
			return nil, err
		}
	}

	return &dest, nil
}
