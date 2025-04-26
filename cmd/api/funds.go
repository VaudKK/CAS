package main

import (
	"errors"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/VaudKK/CAS/pkg/models"
	"github.com/VaudKK/CAS/utils"
	"github.com/gorilla/mux"
)

func (app *application) addContribution(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Contributor string             `json:"contributor"`
		Date        time.Time          `json:"date"`
		Total       float64            `json:"total"`
		BreakDown   map[string]float64 `json:"breakDown"`
		ReceiptNo   string             `json:"receiptNo"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	contributions := []models.Fund{{
		Contributor:    input.Contributor,
		Date:           input.Date.Format("2006-01-02"),
		Total:          input.Total,
		BreakDown:      input.BreakDown,
		ReceiptNo:      input.ReceiptNo,
		OrganizationId: 1,
	}}

	_, err = app.fundsModel.Insert(contributions)

	if err != nil {
		app.writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, envelope{"message": "successfully added contribution"})

}

func (app *application) getContributions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		app.writeJSONError(w, http.StatusMethodNotAllowed, errors.New(http.StatusText(http.StatusMethodNotAllowed)))
		return
	}

	qs := r.URL.Query()
	page := app.readIntParam(qs, "page", 1)
	size := app.readIntParam(qs, "size", 10)

	pageable := utils.Pageable{
		Page:   page,
		Size:   size,
		OffSet: page * size,
	}

	contributions, pageInfo, err := app.fundsModel.GetContributions(1, pageable)

	if err != nil {
		app.writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"data": contributions, "pageInfo": pageInfo})
}

func (app *application) updateContribution(w http.ResponseWriter,r *http.Request){
	vars := mux.Vars(r)
	contributionId := vars["id"]

	if contributionId == "" {
		app.writeJSONError(w,http.StatusBadRequest,errors.New("missing contributor id in path variable"))
		return
	}

	input := models.UpdateFund{}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	id, err := strconv.Atoi(contributionId)

	if err != nil {
		app.writeJSONError(w,http.StatusBadRequest,err)
		return
	}

	_,err = app.fundsModel.UpdateContribution(id,&input)

	if err != nil {
		app.writeJSONError(w,http.StatusInternalServerError,err)
		return
	}


	app.writeJSON(w,http.StatusOK,envelope{"message": "updated successfully"})

}

func (app *application) getMonthlyStats(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()

	year := app.readIntParam(qs, "year", int(time.Now().Year()))
	month := app.readIntParam(qs, "month", int(time.Now().Month()))

	stats, err := app.fundsModel.GetMonthlyStatistics(year, month, 1)

	if err != nil {
		app.writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	app.writeJSON(w, http.StatusOK, stats)
}

func (app *application) search(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		app.writeJSONError(w, http.StatusMethodNotAllowed, errors.New(http.StatusText(http.StatusMethodNotAllowed)))
		return
	}

	qs := r.URL.Query()

	page := app.readIntParam(qs, "page", 1)
	size := app.readIntParam(qs, "size", 10)

	pageable := utils.Pageable{
		Page:   page,
		Size:   size,
		OffSet: page * size,
	}

	searchTerm := qs.Get("terms")

	contributions, pageInfo, err := app.fundsModel.FullTextSearch(searchTerm, pageable)

	if err != nil {
		app.writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"data": contributions, "pageInfo": pageInfo})
}

func (app *application) upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		app.writeJSONError(w, http.StatusMethodNotAllowed, errors.New(http.StatusText(http.StatusMethodNotAllowed)))
		return
	}

	err := r.ParseMultipartForm(10 << 20) // limit upload to 10MB

	if err != nil {
		app.writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	file, handler, err := r.FormFile("document")

	if err != nil {
		app.writeJSONError(w, http.StatusBadRequest, errors.New("missing file or field name 'document'"))
		return
	}

	fileName := handler.Filename

	if filepath.Ext(fileName) != ".xlsx" {
		app.writeJSONError(w, http.StatusBadRequest, errors.New("invalid file type, expected xlsx excel file"))
		return
	}

	go app.fundsModel.ProcessExcelFile(file)

	app.writeJSON(w, http.StatusOK, map[string]string{"message": "file uploaded successfully"})

}
