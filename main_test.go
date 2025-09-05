package forms

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testCasesGetFormAsStruct = []testCase{
		{
			desc:   "basline test",
			method: http.MethodPost,
			form: url.Values{
				"username":     []string{"lorem"},
				"password":     []string{"ipsum"},
				"multipleInts": []string{"3", "4", "5"},
				// "multipleStrings": []string{"lorem", "ipsum"},
			},
			expect: testStruct{
				Username:        "lorem",
				Password:        "ipsum",
				Blah:            "blah blah blah",
				MultipleInts:    []int{3, 4, 5},
				MultipleStrings: []string{"blah blah"},
			},
			getData: func(req *http.Request) (*any, error) {
				return boxPtr(GetStruct[testStruct](req))
			},
		},
		{
			desc:           "test required",
			method:         http.MethodGet,
			form:           url.Values{"username": []string{"username"}},
			expect:         3,
			expectError:    true,
			fieldErrorType: fieldIsRequired,
			getData: func(req *http.Request) (*any, error) {
				return boxPtr(GetStruct[testStruct](req))
			},
		},
		{
			desc:           "test required",
			method:         http.MethodGet,
			form:           url.Values{"username": []string{"username"}},
			expect:         3,
			expectError:    true,
			fieldErrorType: fieldIsRequired,
			getData: func(req *http.Request) (*any, error) {
				return boxPtr(GetStruct[testStruct](req))
			},
		},
		{
			desc:   "test notempty",
			method: http.MethodPost,
			form: url.Values{
				"username":     []string{""},
				"password":     []string{""},
				"multipleInts": []string{"3", "4", "5"},
			},
			expect:         3,
			expectError:    true,
			fieldErrorType: fieldMustNotBeEmpty,
			getData: func(req *http.Request) (*any, error) {
				return boxPtr(GetStruct[testStruct](req))
			},
		},
		{
			desc:   "test default",
			method: http.MethodPost,
			form: url.Values{
				"username":     []string{"lorem"},
				"password":     []string{"ipsum"},
				"multipleInts": []string{"3", "4", "5"},
				// "multipleStrings": []string{"lorem", "ipsum"},
			},
			expect: testStruct{
				Username:        "lorem",
				Password:        "ipsum",
				MultipleInts:    []int{3, 4, 5},
				MultipleStrings: []string{"blah blah"},
				Blah:            "blah blah blah",
			},
			getData: func(req *http.Request) (*any, error) {
				return boxPtr(GetStruct[testStruct](req))
			},
		},
		{
			desc:   "test mutually exclusive required and default",
			method: http.MethodPost,
			form: url.Values{
				"field": []string{"d"},
			},
			expect:         3,
			expectError:    true,
			fieldErrorType: fieldRequiredWithDefault,
			getData: func(req *http.Request) (*any, error) {
				return boxPtr(GetStruct[testRequiredAndDefault](req))
			},
		},
		{
			desc:   "test GET",
			method: http.MethodGet,
			form: url.Values{
				"fieldget":  []string{"get"},
				"fieldpost": []string{"this should be ignored"},
				"FieldAll":  []string{"blah"},
			},
			expect: testMethodsStruct{
				FieldGET:  "get",
				FieldPOST: "",
				FieldAll:  "blah",
			},
			getData: func(req *http.Request) (*any, error) {
				return boxPtr(GetStruct[testMethodsStruct](req))
			},
		},
		{
			desc:   "test POST",
			method: http.MethodPost,
			form: url.Values{
				"fieldget":  []string{"this should be ignored"},
				"fieldpost": []string{"post"},
				"FieldAll":  []string{"blah"},
			},
			expect: testMethodsStruct{
				FieldGET:  "",
				FieldPOST: "post",
				FieldAll:  "blah",
			},
			getData: func(req *http.Request) (*any, error) {
				return boxPtr(GetStruct[testMethodsStruct](req))
			},
		},
		{
			desc:   "test pointer fields",
			method: http.MethodPost,
			form: url.Values{
				"id": []string{"3"},
			},
			expect: testPtrFields{
				ID: getPtr(3),
			},
			getData: func(req *http.Request) (*any, error) {
				return boxPtr(GetStruct[testPtrFields](req))
			},
			expectError:    true,
			fieldErrorType: fieldUnaddressible,
		},
	}
	testCasesFillStructFromForm = []testCase{
		{
			desc:   "baseline test",
			method: http.MethodPost,
			form: url.Values{
				"username":     []string{"lorem"},
				"password":     []string{"ipsum"},
				"multipleInts": []string{"3", "4", "5"},
				// "multipleStrings": []string{"lorem", "ipsum"},
			},
			expect: testStruct{
				Username:        "lorem",
				Password:        "ipsum",
				Blah:            "blah blah blah",
				MultipleInts:    []int{3, 4, 5},
				MultipleStrings: []string{"blah blah"},
			},
			getData: func(req *http.Request) (*any, error) {
				var ts testStruct
				err := FillStructFromForm(req, &ts)
				return boxPtr(&ts, err)
			},
		},
		{
			desc:   "test numeric types",
			method: http.MethodPost,
			form: url.Values{
				"int8":    []string{"-1"},
				"int16":   []string{"-2"},
				"int32":   []string{"-3"},
				"int64":   []string{"-4"},
				"uint8":   []string{"1"},
				"uint16":  []string{"2"},
				"uint32":  []string{"3"},
				"uint64":  []string{"4"},
				"float32": []string{"3.14"},
				"float64": []string{"-46"},
			},
			expect: testNumericTypes{
				Int8:    -1,
				Int16:   -2,
				Int32:   -3,
				Int64:   -4,
				Uint8:   1,
				Uint16:  2,
				Uint32:  3,
				Uint64:  4,
				Float32: 3.14,
				Float64: -46,
			},
			getData: func(req *http.Request) (*any, error) {
				var tnt testNumericTypes
				err := FillStructFromForm(req, &tnt)
				return boxPtr(&tnt, err)
			},
		},
	}
)

type testStruct struct {
	Username        string   `form:"username,required,notempty"`
	Password        string   `form:"password,required,notempty"`
	Blah            string   `form:"blah,default=blah blah blah"`
	MultipleInts    []int    `form:"multipleInts,required"`
	MultipleStrings []string `form:"multipleStrings,default=blah blah"`
	SkipMe          string   `form:"-" method:"POST"`
}

type testRequiredAndDefault struct {
	Field string `form:"field,required,default=blah"`
}

type testMethodsStruct struct {
	FieldGET  string `form:"fieldget" method:"GET"`
	FieldPOST string `form:"fieldpost" method:"POST"`
	FieldAll  string
}

type testNumericTypes struct {
	Int8    int8    `form:"int8,required"`
	Int16   int16   `form:"int16,required"`
	Int32   int32   `form:"int32,required"`
	Int64   int64   `form:"int64,required"`
	Uint8   uint8   `form:"uint8,required"`
	Uint16  uint16  `form:"uint16,required"`
	Uint32  uint32  `form:"uint32,required"`
	Uint64  uint64  `form:"uint64,required"`
	Float32 float32 `form:"float32,required"`
	Float64 float64 `form:"float64,required"`
}

type testPtrFields struct {
	ID *int `form:"id,required"`
}

type testCase struct {
	desc           string
	form           url.Values
	expect         any
	method         string
	expectError    bool
	fieldErrorType fieldErrorType
	getData        func(req *http.Request) (*any, error)
}

func (tc *testCase) doTest(t *testing.T) {
	req := makeRequest(t, tc.form, tc.method)
	dest, err := tc.getData(req)
	if tc.expectError && assert.Error(t, err) {
		assert.Equal(t, tc.fieldErrorType, err.(*fieldError).errorType)
	} else {
		if !assert.NoError(t, err) {
			return
		}
		if !assert.NotNil(t, dest) {
			return
		}
		if !assert.IsType(t, tc.expect, *dest) {
			return
		}
		assert.Equal(t, tc.expect, *dest)
	}
}

func makeRequest(t *testing.T, form url.Values, method string) *http.Request {
	u := "http://localhost/"
	if method == http.MethodGet {
		u += "?" + form.Encode()
	}
	req, err := http.NewRequest(method, u, strings.NewReader(form.Encode()))
	if !assert.NoError(t, err) {
		return nil
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func boxPtr[T any](ptr *T, err error) (*any, error) {
	if err != nil {
		return nil, err
	}
	var data any = *ptr
	return &data, err
}

func getPtr[T any](a T) *T {
	return &a
}

func TestRequirePointerToStruct(t *testing.T) {
	req := makeRequest(t, url.Values{
		"username": []string{"lorem"},
	}, http.MethodPost)
	if !assert.NotNil(t, req) {
		return
	}
	var tc struct {
		Username string `form:"username,required" method:"POST"`
	}
	err := FillStructFromForm(req, tc)
	assert.ErrorIs(t, err, ErrNeedPointerToStruct)
	err = FillStructFromForm(req, 3)
	assert.ErrorIs(t, err, ErrNeedPointerToStruct)
	ptr := &tc
	ptr2ptr := &ptr
	assert.Panics(t, func() {
		err = FillStructFromForm(req, ptr2ptr)
	})
	_, err = GetStruct[int](req)
	assert.ErrorIs(t, err, ErrNeedStructType)
}

func TestGetFormAsStruct(t *testing.T) {
	for _, tC := range testCasesGetFormAsStruct {
		t.Run(tC.desc, tC.doTest)
	}
}

func TestFillStructFromForm(t *testing.T) {
	for _, tC := range testCasesFillStructFromForm {
		t.Run(tC.desc, tC.doTest)
	}
}

func TestIgnoreRequirementsIfNotSameMethod(t *testing.T) {
	req := makeRequest(t, url.Values{
		"q": []string{"blah"},
	}, http.MethodGet)
	if !assert.NotNil(t, req) {
		return
	}
	type loginForm struct {
		Username string `form:"username,required,notempty" method:"POST"`
		Password string `form:"password,required,notempty" method:"POST"`
		LoginBtn string `form:"dologin,required,notempty" method:"POST"`
	}
	var form loginForm
	err := FillStructFromForm(req, &form)
	if !assert.NoError(t, err, "Expected required/notempty fields to be ignored since the method doesn't match") {
		return
	}

}

type Base struct {
	A string `form:"a,required"`
}

type baseUnexported struct {
	A string `form:"a,required"`
}

type composed struct {
	Base
	B int `form:"b,required"`
}

type composedPtr struct {
	*Base
	B int `form:"b,required"`
}

type composedUnexported struct {
	baseUnexported
	B int `form:"b,required"`
}
type composedUnexportedPtr struct {
	*baseUnexported
	B int `form:"b,required"`
}

func TestStructsWithComposition(t *testing.T) {
	req := makeRequest(t, url.Values{
		"a": []string{"aa"},
		"b": []string{"42"},
	}, http.MethodPost)
	if !assert.NotNil(t, req) {
		return
	}
	t.Run("composed struct", func(t *testing.T) {
		var form composed
		err := FillStructFromForm(req, &form)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "aa", form.A)
		assert.Equal(t, 42, form.B)
	})
	t.Run("composed pointer to struct", func(t *testing.T) {
		var formPtr composedPtr
		err := FillStructFromForm(req, &formPtr)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "aa", formPtr.A)
		assert.Equal(t, 42, formPtr.B)
	})

	t.Run("unexported embedded struct fields", func(t *testing.T) {
		var formUnexported composedUnexported
		assert.PanicsWithValue(t, "reflect.Value.Interface: cannot return value obtained from unexported field or method", func() {
			FillStructFromForm(req, &formUnexported)
		})
	})

	t.Run("unexported embedded struct pointer fields", func(t *testing.T) {
		var formUnexportedPtr composedUnexportedPtr
		assert.PanicsWithValue(t, "reflect: reflect.Value.Set using value obtained using unexported field", func() {
			FillStructFromForm(req, &formUnexportedPtr)
		})
	})
}
