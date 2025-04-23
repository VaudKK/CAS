package main

import (
	"errors"
	"net/http"

	"github.com/VaudKK/CAS/pkg/data"
	"github.com/VaudKK/CAS/pkg/models"
	"github.com/VaudKK/CAS/pkg/validator"
)

func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		app.writeJSONError(w, http.StatusMethodNotAllowed, errors.New(http.StatusText(http.StatusMethodNotAllowed)))
		return
	}

	var input struct {
		Username string `json:"username"`
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
	data.ValidateUserName(v, input.Username)

	if !v.Valid() {
		app.writeJSON(w, http.StatusBadRequest, v.Errors)
		return
	}

	usr, err := app.userModel.GetUserByEmail(input.Email)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrorNoRecords):
			break
		default:
			app.writeJSONError(w, http.StatusInternalServerError, err)
			return
		}
	}

	if usr != nil {
		app.writeJSONError(w,http.StatusBadRequest,errors.New("user already exists"))
		return
	}

	newUser := &models.User{
		UserName:       input.Username,
		Email:          input.Email,
		Password:       input.Password,
		OrganizationId: 1,
	}

	_, err = app.userModel.CreateUser(newUser)

	if err != nil {
		app.writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	err = app.mailer.Send(input.Email, "user_welcome.tmpl",nil )

	if err != nil {
		app.writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, map[string]string{
		"created": "true",
	})

}
