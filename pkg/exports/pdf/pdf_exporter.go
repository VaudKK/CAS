package exports

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/VaudKK/CAS/pkg/models"
	"github.com/VaudKK/CAS/utils"
	"github.com/signintech/gopdf"
)

type PdfExport struct {
	Logger *utils.CLogger
}

const (
	pageHeight   = 842.0 // A4 height in points
	topMargin    = 80.0
	bottomMargin = 60.0
	rowHeight    = 25.0
	colCount     = 5
	tableWidth   = 500.0
	footerHeight = 30.0
)

func (pdfExport *PdfExport) GeneratePdfFile(data []*models.Fund, categories []string, startDate, endDate time.Time) ([]byte, error) {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	// Load font
	err := pdf.AddTTFFont("Roboto", "./pkg/exports/pdf/fonts/roboto/Roboto-Regular.ttf")
	if err != nil {
		return nil, err
	}

	err = pdf.AddTTFFont("Roboto-Bold", "./pkg/exports/pdf/fonts/roboto/Roboto-Bold.ttf")
	if err != nil {
		return nil, err
	}

	// Calculate how many rows per page
	availableHeight := pageHeight - topMargin - bottomMargin - rowHeight - footerHeight
	rowsPerPage := int(availableHeight / (rowHeight * 2))

	colWidth := 120.0
	totalRows := len(data)
	pages := ((totalRows * 2) + rowsPerPage - 1) / rowsPerPage

	rowIndex := 0

	// prevents having a zero paged document
	if pages == 1 {
		pages = 2
	}

	//summary page
	pdf.AddPage()

	// Draw Header
	err = pdf.SetFont("Roboto-Bold", "", 20)
	if err != nil {
		return nil, err
	}
	pdf.SetX(40)
	pdf.SetY(20)
	pdf.Cell(nil, "KCSDA Contributions Report")

	err = pdf.SetFont("Roboto", "", 15)
	if err != nil {
		return nil, err
	}

	pdf.SetX(40)
	pdf.SetY(170)

	if !startDate.IsZero() && !endDate.IsZero() {
		pdf.Cell(nil, fmt.Sprintf("Consolidated summary from: %s to: %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")))
	} else if !startDate.IsZero() && endDate.IsZero() {
		pdf.Cell(nil, fmt.Sprintf("Consolidated summary for: %s", startDate.Format("2006-01-02")))
	}else{
		pdf.Cell(nil, "Consolidated summary")
	}

	y := 200.0

	summation := make(map[string]float64)

	for _, row := range data {
		for key, value := range row.BreakDown {
			if val, ok := summation[key]; ok {
				summation[key] = val + value
			} else {
				summation[key] = value
			}
		}
	}

	headers := []string{"N", "Fund Category", "Amount"}
	drawRow(pdf, headers, 40, y, 200, rowHeight, true, false)
	y += rowHeight

	k := 0
	for key, value := range summation {
		row := []string{fmt.Sprintf("%d", k+1), key, fmt.Sprintf("%.2f", value)}
		drawRow(pdf, row, 40, y, 200, rowHeight, true, false)
		y += rowHeight
		k += 1
	}

	pdf.SetX(40)
	pdf.SetY(y + 50)
	pdf.Cell(nil, "Turn over page for detailed contributions")

	for page := 0; page < pages/2; page++ {
		pdf.AddPage()

		y := topMargin

		err = pdf.SetFont("Roboto-Bold", "", 20)
		if err != nil {
			return nil, err
		}
		pdf.SetX(40)
		pdf.SetY(20)
		pdf.Cell(nil, "Details of Contributions")

		// Draw header row
		headers := []string{"N", "Name", "Receipt", "Total", "Date"}
		drawRow(pdf, headers, 40, y, colWidth, rowHeight, true, false)
		y += rowHeight

		// Draw data rows for this page
		for i := 0; i < rowsPerPage && rowIndex < totalRows; i++ {
			rowData := data[rowIndex]

			breakDown := ""

			for category, amount := range rowData.BreakDown {
				breakDown += fmt.Sprintf("%s: %.2f  ", category, amount)
			}

			row := []string{fmt.Sprintf("%d", (rowIndex + 1)), rowData.Contributor, rowData.ReceiptNo,
				fmt.Sprintf("%.2f", rowData.Total), strings.Split(rowData.Date, "T")[0]}

			drawRow(pdf, row, 40, y, colWidth, rowHeight, false, false)
			y += rowHeight

			drawRow(pdf, []string{breakDown}, 40, y, 500, rowHeight, false, true)
			y += rowHeight
			rowIndex++
		}

		// Draw footer
		pdf.SetX(40)
		pdf.SetY(pageHeight - bottomMargin + 10)
		pdf.Cell(nil, fmt.Sprintf("Page %d of %d %50v", page+1, pages/2, "Generated on " + time.Now().Format("2006-01-02 15:04:05")))
	}

	// Write to buffer
	var buf bytes.Buffer
	_, err = pdf.WriteTo(&buf)
	if err != nil {
		panic(err)
	}

	return buf.Bytes(), nil
}

func sum(vals []float64) float64 {
	var total float64
	for _, v := range vals {
		total += v
	}
	return total
}

func drawRow(pdf *gopdf.GoPdf, cells []string, x, y, colWidth, rowHeight float64, isHeader bool, isSingleCellRow bool) {
	pdf.SetX(x)
	pdf.SetY(y)

	if isSingleCellRow {
		pdf.RectFromUpperLeftWithStyle(x, y, 500, rowHeight, "D")
		pdf.SetX(x + 5)
		pdf.SetY(y + 7)
		pdf.SetFont("Roboto", "", 10)
		pdf.Cell(nil, cells[0])
		return
	}

	for i, cell := range cells {
		if i == 0 {
			pdf.RectFromUpperLeftWithStyle(x, y, 20, rowHeight, "D")
		} else {
			pdf.RectFromUpperLeftWithStyle(x, y, colWidth, rowHeight, "D")
		}

		pdf.SetX(x + 5)
		pdf.SetY(y + 7)

		if isHeader {
			pdf.SetFont("Roboto-Bold", "", 12)
		} else {
			pdf.SetFont("Roboto", "", 10)
		}
		pdf.Cell(nil, cell)

		if i == 0 {
			x += 20
		} else {
			x += colWidth
		}
	}
}
