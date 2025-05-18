package exports

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/VaudKK/CAS/pkg/models"
	"github.com/signintech/gopdf"
)


type PdfExport struct {
}

func (pdfExport *PdfExport) GeneratePdfFile(data []*models.Fund, categories []string) ([]byte, error) {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4,})

	pdf.AddPage()

	err := pdf.AddTTFFont("Roboto", "./pkg/exports/pdf/fonts/roboto/Roboto-Regular.ttf")
    if err != nil {
        return nil,err
    }

    err = pdf.SetFont("Roboto", "", 12)
    if err != nil {
        return nil, err
    }

    // Title
	pdf.SetX(50)
	pdf.SetY(40)
	pdf.Cell(nil, "KCSDA Contribution")
	
    // Set the starting Y position for the table
	tableStartY := 10.0

	// Create a new table layout
	table := pdf.NewTableLayout(10.0, tableStartY, 25, 5)

	// Add columns to the table
	// Header row
	headers := []string{"Name", "Receipt", "Total", "Date", "Break Down"}
	colWidths := []float64{150, 100, 100, 100,100}
	
	y := 80.0
	for i, h := range headers {
		pdf.SetX(50 + sum(colWidths[:i]))
		pdf.SetY(y)
		pdf.CellWithOption(&gopdf.Rect{W: colWidths[i], H: 20}, h, gopdf.CellOption{Align: gopdf.Left})
	}


	// Add rows to the table

	y += 25
	for _, fund := range data {

		breakDown := ""

		for category, amount := range fund.BreakDown {
			breakDown += fmt.Sprintf("%s: %.2f\n", category, amount)
		}

		row := []string{fund.Contributor, fund.ReceiptNo, 
			fmt.Sprintf("%.2f",fund.Total), strings.Split(fund.Date, "T")[0], breakDown}

		for j, cell := range row {
			pdf.SetX(50 + sum(colWidths[:j]))
			pdf.SetY(y)
			pdf.CellWithOption(&gopdf.Rect{W: colWidths[j], H: 20}, cell, gopdf.CellOption{Align: gopdf.Left})
		}
		y += 25
	}

	table.DrawTable()

	// Write to buffer
	var buf bytes.Buffer
	_,err = pdf.WriteTo(&buf)
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