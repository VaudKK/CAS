package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
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
			ReceiptNo:      row.ReceiptNo,
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
	stmt := `INSERT INTO funds(break_down,total,organization_id,contribution_date,contributor,receipt_no) VALUES`

	for i, contribution := range contributions {
		breakDown, err := json.Marshal(contribution.BreakDown)

		if err != nil {
			utils.GetLoggerInstance().ErrorLog.Println(err)
			continue
		}

		s := fmt.Sprintf("('%s',%.2f,%d,'%s','%s','%s')", string(breakDown), contribution.Total, contribution.OrganizationId, contribution.Date,
			contribution.Contributor, contribution.ReceiptNo)

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

func (m *FundsModel) GetContributions(organizationId int, pageable utils.Pageable) ([]*models.Fund, utils.PageInfo, error) {
	stmt := `SELECT count(*) OVER(), id,receipt_no,total,organization_id,contribution_date,
	contributor,break_down,created_at,modified_at FROM funds WHERE organization_id = $1 ORDER BY contribution_date DESC, id DESC LIMIT $2 OFFSET $3;`

	rows, err := m.DB.Query(stmt, organizationId, pageable.Size, pageable.OffSet)

	if err != nil {
		return nil, utils.PageInfo{}, err
	}

	defer rows.Close()

	contributions := []*models.Fund{}

	jsonb := make([]byte, 0)
	totalRecords := 0

	for rows.Next() {
		row := &models.Fund{}
		err := rows.Scan(&totalRecords, &row.ID, &row.ReceiptNo, &row.Total, &row.OrganizationId, &row.Date, &row.Contributor,
			&jsonb, &row.Audit.CreatedAt, &row.Audit.ModifiedAt)

		if err != nil {
			return nil, utils.PageInfo{}, err
		}

		err = json.Unmarshal(jsonb, &row.BreakDown)

		if err != nil {
			utils.GetLoggerInstance().ErrorLog.Println(err)
		}

		contributions = append(contributions, row)
	}

	pageInfo := utils.PageInfo{
		CurrentPage:  pageable.Page,
		Size:         pageable.Size,
		TotalRecords: totalRecords,
		FirstPage:    0,
		LastPage:     int(math.Floor(float64(totalRecords) / float64(pageable.Size))),
	}

	if err = rows.Err(); err != nil {
		return nil, utils.PageInfo{}, err
	}

	return contributions, pageInfo, nil
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
