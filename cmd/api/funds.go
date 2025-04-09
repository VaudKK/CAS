package main

import (
	"errors"
	"net/http"
	"path/filepath"
)

func (app *application) upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		app.writeJsonError(w, http.StatusMethodNotAllowed, errors.New(http.StatusText(http.StatusMethodNotAllowed)))
		return
	}

	err := r.ParseMultipartForm(10 << 20) // limit upload to 10MB

	if err != nil {
		app.writeJsonError(w, http.StatusBadRequest, err)
		return
	}

	file, handler, err := r.FormFile("document")

	if err != nil {
		app.writeJsonError(w, http.StatusBadRequest, errors.New("missing file or field name 'document'"))
		return
	}

	fileName := handler.Filename

	if filepath.Ext(fileName) != ".xlsx" {
		app.writeJsonError(w, http.StatusBadRequest, errors.New("invalid file type, expected xlsx excel file"))
		return
	}

	go app.fundsModel.ProcessExcelFile(file)

	app.writeJson(w, http.StatusOK, map[string]string{"message": "file uploaded successfully"})

}
