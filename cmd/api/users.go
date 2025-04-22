package main

import (
	"errors"
	"net/http"
)

func (app *application) createUser(w http.ResponseWriter,r *http.Request){
	if r.Method != "POST" {
		app.writeJsonError(w, http.StatusMethodNotAllowed, errors.New(http.StatusText(http.StatusMethodNotAllowed)))
		return
	}

	
}