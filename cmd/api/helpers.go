package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"

	"github.com/VaudKK/CAS/utils"
)

type envelope map[string]interface{}

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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(httpStatus)
	w.Write(js)

	return nil
}

func (app *application) writeJson(w http.ResponseWriter, httpStatus int, body any) error {
	w.Header().Set("Content-Type", "appication/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(httpStatus)

	js,err := json.Marshal(body)
	
	if err != nil{
		return err
	}

	w.Write(js)

	return nil
}
// TODO handle large integer values
func (app *application) readIntParam(values url.Values,key string,defaultValue int) int {
	s := values.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		utils.GetLoggerInstance().InfoLog.Printf("Error converting %s to int", s)
		return defaultValue
	}
	return i
}
