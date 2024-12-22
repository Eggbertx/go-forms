# go-forms
A package for simplifying parsing and basic validation of HTTP forms into Go structs using struct tags

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