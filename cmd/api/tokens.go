package main

import (
	"net/http"

	"github.com/VaudKK/CAS/pkg/data"
	"github.com/VaudKK/CAS/pkg/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request){
	var input struct  {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.writeJsonError(w,http.StatusBadRequest,err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v,input.Email)
	data.ValidatePasswordPlainText(v,input.Password)

	if (!v.Valid()){
		app.writeJson(w,http.StatusBadRequest,v.Errors)
		return
	}
	
}