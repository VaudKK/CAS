package main

import (
	"errors"
	"math"
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
		Date        string             `json:"date"`
		Total       float64            `json:"total"`
		BreakDown   map[string]float64 `json:"breakDown"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	t, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		app.writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	contributions := []models.Fund{{
		Contributor:    input.Contributor,
		Date:           t.Format("2006-01-02"),
		Total:          input.Total,
		BreakDown:      input.BreakDown,
		OrganizationId: 1,
	}}

	if !app.fundsModel.ValidateTotalAndBreakDown(input.Total, input.BreakDown) {
		app.writeJSONError(w, http.StatusBadRequest, errors.New("total and break down items dont tally"))
		return
	}

	_, err = app.fundsModel.SaveContributions(app.contextGetUser(r), contributions)

	if err != nil {
		app.writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, envelope{"message": "successfully added contribution"})

}

func (app *application) getContributions(w http.ResponseWriter, r *http.Request) {
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

func (app *application) updateContribution(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contributionId := vars["id"]

	if contributionId == "" {
		app.writeJSONError(w, http.StatusBadRequest, errors.New("missing contributor id in path variable"))
		return
	}

	input := models.UpdateFund{}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	if !app.fundsModel.ValidateTotalAndBreakDown(input.Total, input.BreakDown) {
		app.writeJSONError(w, http.StatusBadRequest, errors.New("total and break down items dont tally"))
		return
	}

	id, err := strconv.Atoi(contributionId)

	if err != nil {
		app.writeJSONError(w, http.StatusBadRequest, errors.New("path parameter must be a positive INTEGER"))
		return
	}

	_, err = app.fundsModel.UpdateContribution(id, &input)

	if err != nil {
		app.writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"message": "updated successfully"})

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

func (app *application) getStatisticalVariance(w http.ResponseWriter, _ *http.Request) {

	stats, err := app.fundsModel.GetMonthlyVariance([]string{"LCB", "COMB. OFFERING", "BUILDING CHURCH FUNDS"})

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
	dateFrom, hasFrom := app.readDateParam(qs, "from")
	dateTo, hasTo := app.readDateParam(qs, "to")
	generateExcel := qs.Get("generateExcel")
	generatePdf := qs.Get("generatePdf")
	exact := qs.Get("exact")

	pageable := utils.Pageable{
		Page:   page,
		Size:   size,
		OffSet: page * size,
	}

	if generateExcel == "true" || generatePdf == "true" {
		pageable.Size = math.MaxInt
		pageable.Page = 0
		pageable.OffSet = 0
	}

	searchTerm := qs.Get("terms")

	var contributions []*models.Fund
	var pageInfo utils.PageInfo
	var err error

	if searchTerm != "" {
		contributions, pageInfo, err = app.fundsModel.FullTextSearch(searchTerm, exact == "true", dateFrom, dateTo, pageable)
	} else if hasFrom && hasTo {
		contributions, pageInfo, err = app.fundsModel.SearchByDateRange(dateFrom, dateTo, pageable)
	} else if hasFrom && !hasTo {
		contributions, pageInfo, err = app.fundsModel.SearchByDateRange(dateFrom, time.Time{}, pageable)
	} else {
		app.writeJSONError(w, http.StatusBadRequest, errors.New("missing query params, specify search term or both from and to dates"))
		return
	}

	if err != nil {
		app.writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if generatePdf == "true" {
		file, err := app.fundsModel.GeneratePdfFile(contributions, dateFrom, dateTo)
		if err != nil {
			app.writeJSONError(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=contributions.pdf")
		w.Write(file)
		return
	}

	if generateExcel == "true" {
		file, err := app.fundsModel.GenerateExcelFile(contributions)
		if err != nil {
			app.writeJSONError(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=contributions.xlsx")
		w.Write(file)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"data": contributions, "pageInfo": pageInfo})
}

func (app *application) getSummary(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()

	startDate, hasFrom := app.readDateParam(qs, "from")
	endDate, _ := app.readDateParam(qs, "to")

	if !hasFrom {
		app.writeJSONError(w, http.StatusBadRequest, errors.New("missing from request param"))
		return
	}

	file, err := app.fundsModel.GetSummary(startDate, endDate, 1)
	if err != nil {
		app.writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=contributions-summary.xlsx")
	w.Write(file)
}

func (app *application) upload(w http.ResponseWriter, r *http.Request) {
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

	data, err := app.fundsModel.ValidateFile(file, fileName)

	if err != nil {
		app.writeJSONError(w, http.StatusBadRequest, errors.New("file has already been uploaded or could not be saved"))
		return
	}

	go app.fundsModel.ProcessExcelFile(app.contextGetUser(r), data, fileName)

	app.writeJSON(w, http.StatusOK, map[string]string{"message": "file uploaded successfully"})

}

func (app *application) getCategories(w http.ResponseWriter, r *http.Request) {
	data := app.fundsModel.GetCategories()

	if data == nil {
		app.writeJSON(w, http.StatusOK, make([]string, 0))
		return
	}

	app.writeJSON(w, http.StatusOK, data)
}
