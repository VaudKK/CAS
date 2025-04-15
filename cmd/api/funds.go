package main

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/VaudKK/CAS/utils"
)

func (app *application) getContributions(w http.ResponseWriter, r *http.Request){
	if r.Method != "GET" {
		app.writeJsonError(w, http.StatusMethodNotAllowed, errors.New(http.StatusText(http.StatusMethodNotAllowed)))
		return
	}

	qs := r.URL.Query()
	page := app.readIntParam(qs, "page", 1)
	size := app.readIntParam(qs, "size", 10)

	pageable := utils.Pageable{
		Page:    page,
		Size:    size,
		OffSet:  page * size,
	}

	contributions, pageInfo, err := app.fundsModel.GetContributions(1,pageable)

	if err != nil {
		app.writeJsonError(w, http.StatusInternalServerError, err)
		return
	}

	app.writeJson(w, http.StatusOK, envelope{"data": contributions, "pageInfo": pageInfo})
}

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
