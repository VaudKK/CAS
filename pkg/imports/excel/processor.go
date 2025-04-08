package excel

import (
	"bytes"
	"io"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/VaudKK/CAS/pkg/imports"
	"github.com/xuri/excelize/v2"
)

type ExcelImport struct {
}

func (exImport *ExcelImport) ProcessExcelFile(file multipart.File) ([]imports.ImportModel, error) {
	defer file.Close()

	data, err := io.ReadAll(file)

	if err != nil {
		return []imports.ImportModel{}, err
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))

	if err != nil {
		return []imports.ImportModel{}, err
	}

	defer f.Close()

	excelData := make([]imports.ImportModel, 0)

	for _, sheet := range f.GetSheetList() {
		// read the sheets with the data, the sheets with data are named as dates
		layout := "_2 Jan 2006"

		t, err := time.Parse(layout, exImport.cleanDate(sheet))
		if err != nil {
			break
		}

		rows, err := f.GetRows(sheet)

		if err != nil {
			return []imports.ImportModel{}, err
		}

		// read the categories from the first row
		categories := make([]string, 0)

		for i := 3; i < len(rows[0]); i++ {
			if rows[0][i] == "" {
				break
			}
			categories = append(categories, rows[0][i])
		}

		for i := 1; i < len(rows); i++ {
			if rows[i][0] == "" {
				break
			}

			total, err := strconv.ParseFloat(exImport.cleanNumericField(rows[i][2]), 32)

			if err != nil {
				return []imports.ImportModel{}, err
			}

			breakdown, err := exImport.readBreakDown(categories, rows[i])

			if err != nil {
				return []imports.ImportModel{}, err
			}

			excelData = append(excelData, imports.ImportModel{
				Name:      rows[i][0],
				ReceiptNo: rows[i][1],
				Total:     total,
				Date:      t.String(),
				BreakDown: breakdown})
		}

	}

	return excelData, nil
}

func (exImport *ExcelImport) cleanDate(date string) string {
	ordinals := []string{"st", "nd", "rd", "th"}
	for _, ordinal := range ordinals {
		date = strings.ReplaceAll(strings.ToLower(date), ordinal, "")
	}
	return date
}

func (exImport *ExcelImport) cleanNumericField(field string) string {
	replacer := strings.NewReplacer(",", "", " ", "")
	return replacer.Replace(field)
}

func (exImport *ExcelImport) readBreakDown(categories []string, row []string) (map[string]float64, error) {
	breakdown := make(map[string]float64)

	for i := 0; i < len(categories); i++ {
		// categories columns start from 3, so we need to add 3 to the index
		if row[i+3] == "" {
			break
		}

		value, err := strconv.ParseFloat(exImport.cleanNumericField(row[i+3]), 64)

		if err != nil {
			continue
		}

		breakdown[categories[i]] = value
	}

	return breakdown, nil
}
