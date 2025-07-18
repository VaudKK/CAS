package excel

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/VaudKK/CAS/pkg/imports"
	"github.com/xuri/excelize/v2"
)

type ExcelImport struct {
}

func (exImport *ExcelImport) ProcessExcelFile(data []byte) ([]imports.ImportModel, []string, error) {

	uniqueCategories := make(map[string]bool, 0)

	f, err := excelize.OpenReader(bytes.NewReader(data))

	if err != nil {
		return []imports.ImportModel{}, nil, err
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
			return []imports.ImportModel{}, nil, err
		}

		// read the categories from the first row
		categories := make([]string, 0)

		for i := 3; i < len(rows[0]); i++ {
			if rows[0][i] == "" {
				break
			}
			categories = append(categories, rows[0][i])
		}

		exImport.getUniqueCategories(categories, uniqueCategories)

		for i := 1; i < len(rows); i++ {
			if rows[i][0] == "" {
				break
			}

			total, err := strconv.ParseFloat(exImport.cleanNumericField(rows[i][2]), 32)

			if err != nil {
				return []imports.ImportModel{}, nil, err
			}

			breakdown, err := exImport.readBreakDown(categories, rows[i])

			if err != nil {
				return []imports.ImportModel{}, nil, err
			}

			excelData = append(excelData, imports.ImportModel{
				Name:      exImport.escapeSingleQuote(&rows[i][0]),
				ReceiptNo: rows[i][1],
				Total:     total,
				Date:      t,
				BreakDown: breakdown})
		}

	}

	return excelData, exImport.convertMapToStringArray(uniqueCategories), nil
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
			continue
		}

		value, err := strconv.ParseFloat(exImport.cleanNumericField(row[i+3]), 64)

		if err != nil {
			continue
		}

		breakdown[categories[i]] = value
	}

	return breakdown, nil
}

func (exImport *ExcelImport) escapeSingleQuote(s *string) string {
	if condition := strings.Contains(*s, "'"); condition {
		*s = strings.ReplaceAll(*s, "'", "''")
		return *s
	}
	return *s
}

func (exImport *ExcelImport) getUniqueCategories(categories []string, uniqueCategories map[string]bool) {
	for _, category := range categories {
		if _, ok := uniqueCategories[category]; !ok {
			uniqueCategories[category] = true
		}
	}
}

func (exImport *ExcelImport) convertMapToStringArray(uniqueCategories map[string]bool) []string {
	categories := make([]string, 0)

	for category := range uniqueCategories {
		categories = append(categories, category)
	}

	return categories
}

func HashFile(data []byte) string {
	hashObject := sha256.New()
	hashObject.Write(data)
	return fmt.Sprintf("%x", hashObject.Sum(nil))
}
