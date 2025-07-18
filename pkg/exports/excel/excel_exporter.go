package exports

import (
	"fmt"
	"strings"

	"github.com/VaudKK/CAS/pkg/models"
	"github.com/xuri/excelize/v2"
)

type ExcelExport struct {
}

func (exExport *ExcelExport) GenerateExcelFile(data []*models.Fund, categories []string) ([]byte, error) {
	f := excelize.NewFile()

	// Create a new sheet
	index, err := f.NewSheet("Contributions")

	if err != nil {
		return nil, err
	}

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Family: "Arial"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#D3D3D3"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	if err != nil {
		return nil, err
	}

	boarderStyle, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		}})

	if err != nil {
		return nil, err
	}

	f.SetColWidth("Contributions", "A", "AZ", 21)

	categoryIndex := make(map[string]int)

	// Set the headers
	headers := []string{"NAME", "RECEIPT NO", "TOTAL", "DATE"}

	for i, category := range categories {
		headers = append(headers, category)
		categoryIndex[category] = 4 + i
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellStyle("Contributions", cell, cell, headerStyle)
		f.SetColWidth("Contributions", cell, cell, 17)
		f.SetCellValue("Contributions", cell, header)
	}

	// Set the data
	for i, contribution := range data {
		cellContributor, _ := excelize.CoordinatesToCellName(1, i+2)
		cellReceiptNo, _ := excelize.CoordinatesToCellName(2, i+2)
		cellTotal, _ := excelize.CoordinatesToCellName(3, i+2)
		cellDate, _ := excelize.CoordinatesToCellName(4, i+2)

		f.SetCellValue("Contributions", cellDate, strings.Split(contribution.Date, "T")[0])
		f.SetCellValue("Contributions", cellContributor, contribution.Contributor)
		f.SetCellValue("Contributions", cellTotal, contribution.Total)
		f.SetCellValue("Contributions", cellReceiptNo, contribution.ReceiptNo)

		f.SetCellStyle("Contributions", cellContributor, cellDate, boarderStyle)
		f.SetCellStyle("Contributions", cellTotal, cellTotal, boarderStyle)
		f.SetCellStyle("Contributions", cellReceiptNo, cellReceiptNo, boarderStyle)
		f.SetCellStyle("Contributions", cellDate, cellDate, boarderStyle)

		// set border on all data cells
		for k := 3; k < len(headers); k++ {
			cellCategory, _ := excelize.CoordinatesToCellName(k+1, i+2)
			f.SetCellStyle("Contributions", cellCategory, cellCategory, boarderStyle)
		}

		for category, amount := range contribution.BreakDown {
			if j, ok := categoryIndex[category]; ok {
				cellCategory, _ := excelize.CoordinatesToCellName(j+1, i+2)
				f.SetCellValue("Contributions", cellCategory, amount)
			}
		}
	}

	// summation
	row := len(data) + 1

	totalStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Family: "Arial"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D3D3D3"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		}})

	if err != nil {
		return nil, err
	}

	cellContributor, _ := excelize.CoordinatesToCellName(1, row+2)
	cellReceiptNo, _ := excelize.CoordinatesToCellName(2, row+2)
	cellTotal, _ := excelize.CoordinatesToCellName(3, row+2)
	cellDate, _ := excelize.CoordinatesToCellName(4, row+2)

	f.SetCellStyle("Contributions", cellContributor, cellDate, totalStyle)
	f.SetCellStyle("Contributions", cellTotal, cellTotal, totalStyle)
	f.SetCellStyle("Contributions", cellReceiptNo, cellReceiptNo, totalStyle)
	f.SetCellStyle("Contributions", cellDate, cellDate, totalStyle)

	cellTotalStart, _ := excelize.CoordinatesToCellName(3, 2)
	cellTotalEnd, _ := excelize.CoordinatesToCellName(3, row)

	f.SetCellFormula("Contributions", cellTotal, fmt.Sprintf("SUM(%s:%s)", cellTotalStart, cellTotalEnd))

	for i := 3; i < len(categoryIndex)+3; i++ {
		cellStart, _ := excelize.CoordinatesToCellName(i+2, 2)
		cellEnd, _ := excelize.CoordinatesToCellName(i+2, row)

		cellCategory, _ := excelize.CoordinatesToCellName(i+2, row+2)
		f.SetCellFormula("Contributions", cellCategory, fmt.Sprintf("SUM(%s:%s)", cellStart, cellEnd))
		f.SetCellStyle("Contributions", cellCategory, cellCategory, totalStyle)
	}

	// Set the active sheet
	f.SetActiveSheet(index)

	buff, err := f.WriteToBuffer()

	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (exExport *ExcelExport) GenerateExcelSummary(data map[string][]models.MonthlySummations,
	categories []string) ([]byte, error) {

	f := excelize.NewFile()

	index, err := f.NewSheet("ContributionsSummary")

	if err != nil {
		return nil, err
	}

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Family: "Arial"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#D3D3D3"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	if err != nil {
		return nil, err
	}

	boarderStyle, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		}})

	if err != nil {
		return nil, err
	}

	f.SetColWidth("ContributionsSummary", "A", "AZ", 17)

	f.SetCellValue("ContributionsSummary", "A1", "CHURCH TREASURER'S CASH STATEMENT")
	f.SetCellValue("ContributionsSummary", "A2", "KITENGELA CENTRAL SDA CHURCH")
	f.SetCellValue("ContributionsSummary", "A3", "CONSOLIDATED REPORT")

	f.SetCellStyle("ContributionsSummary", "A1", "J1", headerStyle)
	f.SetCellStyle("ContributionsSummary", "A2", "J2", headerStyle)
	f.SetCellStyle("ContributionsSummary", "A3", "J3", headerStyle)

	if err := f.MergeCell("ContributionsSummary", "A1", "J1"); err != nil {
		return nil, err
	}

	if err := f.MergeCell("ContributionsSummary", "A2", "J2"); err != nil {
		return nil, err
	}

	if err := f.MergeCell("ContributionsSummary", "A3", "J3"); err != nil {
		return nil, err
	}

	categoryIndex := make(map[string]int)

	// Set the headers
	headers := []string{"SABBATH"}

	for i, category := range categories {
		headers = append(headers, category)
		categoryIndex[category] = i
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		f.SetCellStyle("ContributionsSummary", cell, cell, headerStyle)
		f.SetCellValue("ContributionsSummary", cell, header)
	}

	// Set the data
	i := 1
	for key, value := range data {
		cellCategory, _ := excelize.CoordinatesToCellName(1, i+4)
		f.SetCellValue("ContributionsSummary", cellCategory, strings.Split(key, "T")[0])

		for _, row := range value {
			if j, ok := categoryIndex[row.Category]; ok {
				j += 1
				cellCategory, _ := excelize.CoordinatesToCellName(j+1, i+4)
				f.SetCellValue("ContributionsSummary", cellCategory, row.Total)
			}
		}

		// set border on all data cells
		for k := 0; k < len(headers); k++ {
			cellCategory, _ := excelize.CoordinatesToCellName(k+1, i+4)
			f.SetCellStyle("ContributionsSummary", cellCategory, cellCategory, boarderStyle)
		}
		i += 1
	}

	// summation
	totalsRow := len(data) + 6

	totalStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Family: "Arial"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D3D3D3"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		}})

	if err != nil {
		return nil, err
	}

	text, _ := excelize.CoordinatesToCellName(1, totalsRow)
	f.SetCellStyle("ContributionsSummary", text, text, totalStyle)
	f.SetCellValue("ContributionsSummary", text, "Total")

	for i := 1; i < len(categoryIndex)+1; i++ {
		cellStart, _ := excelize.CoordinatesToCellName(i+1, 5)
		cellEnd, _ := excelize.CoordinatesToCellName(i+1, totalsRow-2)

		cellCategory, _ := excelize.CoordinatesToCellName(i+1, totalsRow)
		f.SetCellFormula("ContributionsSummary", cellCategory, fmt.Sprintf("SUM(%s:%s)", cellStart, cellEnd))
		f.SetCellStyle("ContributionsSummary", cellCategory, cellCategory, totalStyle)
	}

	f.SetActiveSheet(index)

	buff, err := f.WriteToBuffer()

	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil

}
