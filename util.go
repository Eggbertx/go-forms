package forms

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

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
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if data == "" {
			data = "0"
		}
		intVal, err := strconv.Atoi(data)
		if err != nil {
			return &fieldError{structField: name, val: data, msg: err.Error(), errorType: fieldParsingError}
		}
		val.SetInt(int64(intVal))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if data == "" {
			data = "0"
		}
		uintVal, err := strconv.ParseUint(data, 10, 64)
		if err != nil {
			return &fieldError{structField: name, val: data, msg: err.Error(), errorType: fieldParsingError}
		}
		val.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		if data == "" {
			data = "0"
		}
		floatVal, err := strconv.ParseFloat(data, 64)
		if err != nil {
			return &fieldError{structField: name, val: data, msg: err.Error(), errorType: fieldParsingError}
		}
		val.SetFloat(floatVal)
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
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.String:
		return setFieldVal(name, structFieldVal, formFieldData[0])
	case reflect.Ptr, reflect.Interface:
		return setFieldVal(name, structFieldVal.Elem(), formFieldData[0])
	default:
		return &fieldError{structField: name, msg: "unsupported field type", errorType: fieldHasUnsupportedType}
	}

	return nil
}

type structFieldFormOptions struct {
	formName   string
	required   bool
	notEmpty   bool
	hasDefault bool
	formVals   []string
}

func (sffo *structFieldFormOptions) skip() bool {
	return sffo.formName == ""
}

func parseFieldOptions(req *http.Request, field reflect.StructField) (*structFieldFormOptions, error) {
	var options structFieldFormOptions
	formNameOptionsParts := strings.Split(field.Tag.Get("form"), ",")
	options.formName = formNameOptionsParts[0]
	if options.formName == "" {
		options.formName = field.Name
	} else if options.formName == "-" {
		// skip this field
		return &options, nil
	}
	var ok bool
	options.formVals, ok = requestHasField(req, options.formName, req.Method)
	for _, nameOptionPart := range formNameOptionsParts[1:] {
		// handle field options (required and default)
		if nameOptionPart == "required" {
			options.required = true
		} else if nameOptionPart == "notempty" && field.Type.Kind() == reflect.String {
			options.notEmpty = true
		} else if strings.HasPrefix(nameOptionPart, "default=") {
			options.hasDefault = true
			if !ok {
				// use default value in struct tag
				options.formVals = []string{nameOptionPart[8:]}
			}
		}
	}
	return &options, nil
}
