# go-forms
A package for simplifying parsing and basic validation of HTTP forms into Go structs using struct tags. It supports strings, boolean values (set to true if the field value is "on" or "1", and false otherwise), numeric types, and slices of the same types.

## Usage
 
```Go
import (
	"github.com/Eggbertx/go-forms"
)

type loginForm struct {
	Username string `form:"username,required,notempty" method:"POST"`
	Password string `form:"password,required,notempty" method:"POST"`
	LoginBtn string `form:"dologin,required,notempty" method:"POST"`
	Query string `form:"q"`
	Pages []int `form:"pages" method="POST,PUT"` // multiple fields with the same name attribute
	SomeCheckbox bool `form:"chk" method="POST"`
}


// using GetStruct
formValues, err := forms.GetStruct[loginForm](request)
if err != nil {
	// handle error
}

// using FillStructFromForm
err = forms.FillStructFromForm(req, &formValues)
if err != nil {
	// handle error
}

```

## Field options

The `form` struct tag can be used to customize the expected field name in the request and do some basic validation. It is comma-separated, similarly to the `encoding/json` package. The first part should be the form key to be checked. If you want to use the struct's field name as the form name while still using field options, you can do so with `form:",..."` (replace "..." with the options used, if any)

The valid field form options are:
- `required`: if set, the field is required for the given HTTP method specified by the `method` tag (or for any HTTP method if the method tag is not used). If it is not included in the request, an error will be returned.
- `notempty`: if set, the field is required to not be empty
- `default=[value]`: if included, default will set the default value if the field is not in the request.

The `required` and `default` options are mutually exclusive, and will return an error if both are used.

The `method` struct tag can be used to limit the field to certain HTTP methods (GET, POST, etc). If the request HTTP method does not match the field's set method, it will not be set in the struct, even if it is set as required. It is comma separated, so the field can be used for multiple methods, ex: `method:"POST,PUT"`