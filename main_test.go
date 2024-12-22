package formstructs

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
			fieldErrorType: fieldHasUnsupportedType,
		},
	}
)

type testStruct struct {
	Username        string   `form:"username,required,notempty"`
	Password        string   `form:"password,required,notempty"`
	Blah            string   `form:"blah,default=blah blah blah"`
	MultipleInts    []int    `form:"multipleInts,required"`
	MultipleStrings []string `form:"multipleStrings,default=blah blah"`
}

type testRequiredAndDefault struct {
	Field string `form:"field,required,default=blah"`
}

type testMethodsStruct struct {
	FieldGET  string `form:"fieldget" method:"GET"`
	FieldPOST string `form:"fieldpost" method:"POST"`
	FieldAll  string
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
	u := "http://localhost/"
	if tc.method == http.MethodGet {
		u += "?" + tc.form.Encode()
	}
	req, err := http.NewRequest(tc.method, u, strings.NewReader(tc.form.Encode()))
	if !assert.NoError(t, err) {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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

func TestGetFormAsStruct(t *testing.T) {
	for _, tC := range testCasesGetFormAsStruct {
		t.Run(tC.desc, tC.doTest)
	}
}
