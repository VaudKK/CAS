package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
)

type ErrorResponse struct {
	ErrorMessage string
}

func (app *application) writeJsonError(w http.ResponseWriter, httpStatus int, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())

	app.errorLog.Output(2, trace)

	errResponse := ErrorResponse{
		ErrorMessage: err.Error(),
	}

	w.Header().Set("Content-Type", "appication/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(errResponse)
}

func (app *application) writeJson(w http.ResponseWriter, httpStatus int, body any) {
	w.Header().Set("Content-Type", "appication/json")
	w.WriteHeader(httpStatus)

	json.NewEncoder(w).Encode(body)
}
