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
		app.writeJSONError(w, http.StatusBadRequest, errors.New("user already exists"))
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

	err = app.mailer.Send(input.Email, "user_welcome.tmpl", nil)

	if err != nil {
		app.writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, map[string]string{
		"created": "true",
	})

}

func (app *application) sendOtp(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	if email == "" {
		app.writeJSONError(w, http.StatusBadRequest, errors.New("email parameter missing"))
		return
	}

	v := validator.New()

	data.ValidateEmail(v, email)

	if !v.Valid() {
		app.writeJSON(w, http.StatusBadRequest, v.Errors)
		return
	}

	response, err := app.otpModel.SendOtp(email, "email")

	if err != nil {
		app.writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	app.writeJSON(w, http.StatusOK, response)
}

func (app *application) verifyOtp(w http.ResponseWriter, r *http.Request) {
	sessionId := r.URL.Query().Get("session")
	otp := r.URL.Query().Get("otp")

	if otp == "" || sessionId == "" {
		app.writeJSONError(w, http.StatusBadRequest, errors.New("missing sessionId and/or otp parameter"))
		return
	}

	response, err := app.otpModel.VerifyOtp(otp, sessionId)

	if err != nil {
		app.writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if !response {
		app.writeJSON(w, http.StatusOK, envelope{"message": "invalid otp"})
		return
	} else {
		app.writeJSON(w, http.StatusOK, envelope{"message": "success"})
	}
}

func (app *application) resetPassword(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	if email == "" {
		app.writeJSONError(w, http.StatusBadRequest, errors.New("email parameter missing"))
		return
	}

	v := validator.New()
	data.ValidateEmail(v, email)

	if !v.Valid() {
		app.writeJSON(w, http.StatusBadRequest, v.Errors)
		return
	}

	err := app.userModel.SendResetLink(email)

	if err != nil {
		app.writeJSONError(w, http.StatusInternalServerError, err)
	}

	app.writeJSON(w, http.StatusOK, envelope{"message": "reset link sent to email"})
}

func (app *application) changePassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ResetToken  string `json:"resetToken"`
		NewPassword string `json:"newPassword"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	v := validator.New()
	data.ValidateResetToken(v, input.ResetToken)
	data.ValidatePasswordPlainText(v, input.NewPassword)

	if !v.Valid() {
		app.writeJSON(w, http.StatusBadRequest, v.Errors)
		return
	}

	err = app.userModel.ChangePassword(input.ResetToken, input.NewPassword)

	if err != nil {
		app.writeJSONError(w, http.StatusNotAcceptable, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"message": "password changed successfully"})
}
