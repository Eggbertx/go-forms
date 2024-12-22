# go-forms
A package for simplifying parsing and basic validation of HTTP forms into Go structs using struct tags

## Usage
 
```Go
type loginForm struct {
	Username string `form:"username,required,notempty" method:"POST"`
	Password string `form:"password,required,notempty" method:"POST"`
	LoginBtn string `form:"dologin,required,notempty" method:"POST"`
	Query string `form:"q"`
}


// using GetStruct
formValues, err := GetStruct[loginForm](request)
if err != nil {
	// handle error
}

// using FillStructFromForm
err = FillStructFromForm(req, &formValues)
if err != nil {
	// handle error
}

```