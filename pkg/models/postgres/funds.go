package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"

	"github.com/VaudKK/CAS/pkg/imports/excel"
	"github.com/VaudKK/CAS/pkg/models"
	"github.com/VaudKK/CAS/utils"
)

type FundsModel struct {
	DB *sql.DB
}

func (m *FundsModel) ProcessExcelFile(file multipart.File) {
	excelModel := excel.ExcelImport{}
	data, categories, err := excelModel.ProcessExcelFile(file)

	if err != nil {
		utils.GetLoggerInstance().ErrorLog.Println(err)
		return
	}

	funds := make([]models.Fund, 0)

	for _, row := range data {
		fund := models.Fund{
			BreakDown:      row.BreakDown,
			Total:          row.Total,
			OrganizationId: 1,
			Date:           row.Date.Format("2006-01-02"),
			Contributor:    row.Name,
		}
		funds = append(funds, fund)
	}

	insertedCategories, err := m.SaveCategories(categories)

	if err != nil {
		utils.GetLoggerInstance().ErrorLog.Println(err)
		return
	}

	utils.GetLoggerInstance().InfoLog.Printf("Inserted %d categories", insertedCategories)

	inserted, err := m.Insert(funds)

	if err != nil {
		utils.GetLoggerInstance().ErrorLog.Println(err)
		return
	}

	utils.GetLoggerInstance().InfoLog.Printf("Inserted %d funds", inserted)
}

func (m *FundsModel) Insert(contributions []models.Fund) (int, error) {
	stmt := `INSERT INTO funds(break_down,total,organization_id,contribution_date,contributor) VALUES`

	for i, contribution := range contributions {
		breakDown, err := json.Marshal(contribution.BreakDown)

		if err != nil {
			utils.GetLoggerInstance().ErrorLog.Println(err)
			continue
		}

		s := fmt.Sprintf("('%s',%.2f,%d,'%s','%s')", string(breakDown), contribution.Total, contribution.OrganizationId, contribution.Date,
			contribution.Contributor)

		if i != len(contributions)-1 {
			s += ","
		}

		stmt += s
	}

	stmt += ";"

	result, err := m.DB.Exec(stmt)

	if err != nil {
		return 0, err
	}

	rowAffected, err := result.RowsAffected()

	if err != nil {
		return 0, err
	}

	return int(rowAffected), nil

}

func (m *FundsModel) Get(organizationId int, pageable utils.Pageable) ([]models.Fund, error) {
	return nil, nil
}

func (m *FundsModel) SaveCategories(categories []string) (int, error) {
	stmt := `INSERT INTO fund_categories(name) VALUES`

	for i, category := range categories {
		s := fmt.Sprintf("('%s')", category)

		if i != len(categories)-1 {
			s += ","
		}

		stmt += s
	}

	stmt += ";"

	result, err := m.DB.Exec(stmt)

	if err != nil {
		return 0, err
	}

	rowAffected, err := result.RowsAffected()

	if err != nil {
		return 0, err
	}

	return int(rowAffected), nil
}
