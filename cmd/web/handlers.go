package main

import (
	"errors"
	"net/http"
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

	file, _, err := r.FormFile("document")

	if err != nil {
		app.writeJsonError(w, http.StatusBadRequest, errors.New("missing file or field name 'document'"))
		return
	}

	//fileName := handler.Filename

	//app.infoLog.Printf("Filename :%s File Extension :%s\n", fileName, filepath.Ext(fileName))

	// if filepath.Ext(fileName) != ".xls" || filepath.Ext(fileName) != ".xlsx" {
	// 	app.writeJsonError(w, http.StatusBadRequest, errors.New("invalid file type, expected excel file"))
	// 	return
	// }

	data, err := app.importModel.ProcessExcelFile(file)

	if err != nil {
		app.writeJsonError(w, http.StatusInternalServerError, err)
		return
	}

	app.writeJson(w, http.StatusOK, data)

}
