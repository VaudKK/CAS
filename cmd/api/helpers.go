package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/VaudKK/CAS/utils"
)

type ErrorResponse struct {
	ErrorMessage string
}

func (app *application) writeJsonError(w http.ResponseWriter, httpStatus int, err error) error {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())

	utils.GetLoggerInstance().ErrorLog.Output(2, trace)

	errResponse := ErrorResponse{
		ErrorMessage: err.Error(),
	}

	js,err := json.Marshal(errResponse)
	
	if err != nil{
		return err
	}

	w.Header().Set("Content-Type", "appication/json")
	w.WriteHeader(httpStatus)
	w.Write(js)

	return nil
}

func (app *application) writeJson(w http.ResponseWriter, httpStatus int, body any) error {
	w.Header().Set("Content-Type", "appication/json")
	w.WriteHeader(httpStatus)

	js,err := json.Marshal(body)
	
	if err != nil{
		return err
	}

	w.Write(js)

	return nil
}
