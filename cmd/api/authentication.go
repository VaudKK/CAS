package main

import (
	"errors"
	"net/http"

	"github.com/VaudKK/CAS/pkg/data"
	"github.com/VaudKK/CAS/pkg/validator"
)

func (app *application) issueToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlainText(v, input.Password)

	if !v.Valid() {
		app.writeJSON(w, http.StatusBadRequest, v.Errors)
		return
	}

	usr, err := app.userModel.GetUserByEmail(input.Email)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrorNoRecords):
			app.writeUnauthorizedJSON(w, r)

		default:
			app.writeJSONError(w, http.StatusInternalServerError, err)
		}

		return
	}

	valid, err := app.userModel.VerifyUser(input.Email, input.Password)

	if err != nil {
		app.writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if !valid {
		app.writeUnauthorizedJSON(w, r)
		return
	}


	token,err := data.CreateToken(usr.ID)

	if err != nil {
		app.writeUnauthorizedJSON(w,r)
		return
	}

	app.writeJSON(w,http.StatusOK,token)
}
