package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/VaudKK/CAS/utils"
)

type envelope map[string]interface{}

type ErrorResponse struct {
	ErrorMessage interface{} `json:"errorMessage"`
}

func (app *application) writeJSONError(w http.ResponseWriter, httpStatus int, err error) error {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())

	utils.GetLoggerInstance().ErrorLog.Output(2, trace)

	errResponse := ErrorResponse{
		ErrorMessage: err.Error(),
	}

	js, err := json.Marshal(errResponse)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "appication/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(httpStatus)
	w.Write(js)

	return nil
}

func (app *application) writeJSON(w http.ResponseWriter, httpStatus int, body any) error {
	w.Header().Set("Content-Type", "appication/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(httpStatus)

	js, err := json.Marshal(body)

	if err != nil {
		return err
	}

	w.Write(js)

	return nil
}

func (app *application) writeUnauthorizedJSON(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusUnauthorized, ErrorResponse{
		ErrorMessage: "Unauthorized",
	})
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	err := json.NewDecoder(r.Body).Decode(dst)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUmarshalError *json.InvalidUnmarshalError

		switch {

		// Use the errors.As() function to check whether the error has the type
		// *json.SyntaxError. If it does, then return a plain-english error message
		// which includes the location of the problem.
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		// In some circumstances Decode() may also return an io.ErrUnexpectedEOF error
		// for syntax errors in the JSON. So we check for this using errors.Is() and
		// return a generic error message. There is an open issue regarding this at
		// https://github.com/golang/go/issues/25956.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("body contains badly-formed JSON")

		// Likewise, catch any *json.UnmarshalTypeError errors. These occur when the
		// JSON value is the wrong type for the target destination. If the error relates
		// to a specific field, then we include that in our error message to make it
		// easier for the client to debug.
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		// An io.EOF error will be returned by Decode() if the request body is empty. We
		// check for this with errors.Is() and return a plain-english error message
		// instead.
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// A json.InvalidUnmarshalError error will be returned if we pass a non-nil
		// pointer to Decode(). We catch this and panic, rather than returning an error
		// to our handler.
		case errors.As(err, &invalidUmarshalError):
			panic(err)

		default:
			return err
		}
	}

	return nil
}

// TODO handle large integer values
func (app *application) readIntParam(values url.Values, key string, defaultValue int) int {
	s := values.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		utils.GetLoggerInstance().ErrorLog.Printf("Error converting %s to int", s)
		return defaultValue
	}
	return i
}

func (app *application) readDateParam(values url.Values,key string) (time.Time,bool) {
	s := values.Get(key)
	
	if s == "" {
		return time.Now(),false
	}

	t,err := time.Parse("2006-01-02",s)
	if err != nil {
		utils.GetLoggerInstance().ErrorLog.Printf("Error converting %s to date",s)
		return time.Now(),false
	}

	return t,true
}
