package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"slices"

	exporter "github.com/VaudKK/CAS/pkg/exports/excel"
	pdf_exporter "github.com/VaudKK/CAS/pkg/exports/pdf"
	"github.com/VaudKK/CAS/pkg/imports/excel"
	"github.com/VaudKK/CAS/pkg/models"
	"github.com/VaudKK/CAS/utils"
)

type FundsModel struct {
	DB            *sql.DB
	ExcelExporter *exporter.ExcelExport
	PdfExporter   *pdf_exporter.PdfExport
	Logger        *utils.CLogger
}

func (m *FundsModel) ValidateFile(file multipart.File, fileName string) ([]byte, error) {
	fileData, err := io.ReadAll(file)

	if err != nil {
		utils.GetLoggerInstance().ErrorLog.Println(err)
		return nil, err
	}

	defer file.Close()

	hash := utils.HashFile(fileData)

	err = m.SaveFileHash(hash, fileName, 1)

	if err != nil {
		utils.GetLoggerInstance().ErrorLog.Println(err)
		return nil, err
	}

	return fileData, nil
}

func (m *FundsModel) ProcessExcelFile(currentUser *models.User, fileData []byte, fileName string) error {
	ctx := context.Background()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		m.Logger.ErrorLog.Printf("Error while starting db transaction: %s", err.Error())
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			m.Logger.ErrorLog.Fatalf("Error while processing excel data: %v", err)
		}

		tx.Rollback()
	}()

	excelModel := excel.ExcelImport{}

	data, categories, err := excelModel.ProcessExcelFile(fileData)

	if err != nil {
		m.Logger.ErrorLog.Printf("Error while processing excel data: %v", err)
		return err
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

	insertedCategories, err := m.SaveCategories(tx, ctx, categories)

	if err != nil {
		m.Logger.ErrorLog.Printf("Error while saving categories: %v", err)
		return err
	}

	utils.GetLoggerInstance().InfoLog.Printf("Will insert %d categories", insertedCategories)

	inserted, err := m.insert(tx, ctx, currentUser, funds)

	if err != nil {
		m.Logger.ErrorLog.Printf("Error while saving contributions: %v", err)
		return err
	}

	utils.GetLoggerInstance().InfoLog.Printf("Will insert %d funds", inserted)

	tx.Commit()
	return nil
}

func (m *FundsModel) SaveContributions(user *models.User, contributions []models.Fund) (int, error) {
	ctx := context.Background()
	tx, err := m.DB.BeginTx(ctx, nil)

	if err != nil {
		m.Logger.ErrorLog.Fatalf("Error while starting transaction: %v", err)
		return -1, err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			m.Logger.ErrorLog.Fatalf("Error while saving contribution(s): %v", err)
		}

		tx.Rollback()
	}()

	response, err := m.insert(tx, ctx, user, contributions)
	tx.Commit()

	return response, err
}

func (m *FundsModel) insert(tx *sql.Tx, ctx context.Context, currentUser *models.User, contributions []models.Fund) (int, error) {

	stmt := `INSERT INTO funds(break_down,total,organization_id,contribution_date,contributor,receipt_no,created_by) VALUES`

	for i, contribution := range contributions {
		breakDown, err := json.Marshal(contribution.BreakDown)

		if err != nil {
			utils.GetLoggerInstance().ErrorLog.Println(err)
			continue
		}

		if contribution.ReceiptNo == "" {
			contribution.ReceiptNo = "KCS-" + strconv.Itoa(int(time.Now().UnixMicro()))
		}

		s := fmt.Sprintf("('%s',%.2f,%d,'%s','%s','%s',%s)", string(breakDown), contribution.Total, contribution.OrganizationId, contribution.Date,
			strings.ToUpper(contribution.Contributor), contribution.ReceiptNo, strconv.Itoa(currentUser.ID))

		if i != len(contributions)-1 {
			s += ","
		}

		stmt += s
	}

	stmt += ";"

	result, err := tx.ExecContext(ctx, stmt)

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

	contributions, pageInfo := mapSqlRowsToModel(rows, pageable)

	if err = rows.Err(); err != nil {
		return nil, utils.PageInfo{}, err
	}

	return contributions, pageInfo, nil
}

func (m *FundsModel) UpdateContribution(id int, updateFund *models.UpdateFund) (int, error) {
	stmt := `UPDATE funds SET total = $1,contribution_date = $2,contributor = $3,break_down = $4 WHERE
				id = $5;`

	breakDown, err := json.Marshal(updateFund.BreakDown)

	if err != nil {
		return 0, err
	}

	result, err := m.DB.Exec(stmt, updateFund.Total, updateFund.Date, strings.ToUpper(updateFund.Contributor), string(breakDown), id)

	if err != nil {
		return 0, err
	}

	rowAffected, err := result.RowsAffected()

	if err != nil {
		return 0, err
	}

	m.Logger.InfoLog.Printf("Updated contribution with ID %d, rows affected: %d", id, rowAffected)

	return int(rowAffected), nil
}

func (m *FundsModel) FullTextSearch(searchString string, exact bool, startDate, endDate time.Time, pageable utils.Pageable) ([]*models.Fund, utils.PageInfo, error) {

	var query string

	match := "(to_tsvector(contributor || ' ' || receipt_no) @@ to_tsquery('%s'))"

	if exact {
		match = fmt.Sprintf("(contributor ILIKE '%%%s%%' OR receipt_no ILIKE '%%%s%%')", searchString, searchString)
	}

	if startDate.IsZero() && endDate.IsZero() {
		query = `SELECT count(*) OVER(), id,receipt_no,total,organization_id,contribution_date,
						contributor,break_down,created_at,modified_at 
				FROM funds
				where organization_id = $1 AND ` + match + `
				ORDER BY created_at DESC LIMIT $2 OFFSET $3;`
	} else if !startDate.IsZero() && !endDate.IsZero() {
		query = `SELECT count(*) OVER(), id,receipt_no,total,organization_id,contribution_date,
						contributor,break_down,created_at,modified_at 
				FROM funds
				where organization_id = $1 AND ` + match + `
				AND contribution_date BETWEEN $2 AND $3
				ORDER BY created_at DESC LIMIT $4 OFFSET $5;`
	} else if !startDate.IsZero() {
		query = `SELECT count(*) OVER(), id,receipt_no,total,organization_id,contribution_date,
						contributor,break_down,created_at,modified_at 
				FROM funds
				where organization_id = $1 AND ` + match + `
				AND contribution_date = $2
				ORDER BY created_at DESC LIMIT $3 OFFSET $4;`
	}

	if !exact {
		tokens := strings.Split(searchString, " ")
		searchTerms := ""

		for i, token := range tokens {
			if token != "" {
				if i == len(tokens)-1 {
					searchTerms += token + ":*"
				} else {
					searchTerms += token + ":* | "
				}
			}
		}

		query = fmt.Sprintf(query, searchTerms)
	}

	var rows *sql.Rows
	var err error

	if startDate.IsZero() && endDate.IsZero() {
		rows, err = m.DB.Query(query, 1, pageable.Size, pageable.OffSet)
	} else if !startDate.IsZero() && !endDate.IsZero() {
		rows, err = m.DB.Query(query, 1, startDate, endDate, pageable.Size, pageable.OffSet)
	} else if !startDate.IsZero() {
		rows, err = m.DB.Query(query, 1, startDate, pageable.Size, pageable.OffSet)
	}

	if err != nil {
		return nil, utils.PageInfo{}, err
	}

	defer rows.Close()

	contributions, pageInfo := mapSqlRowsToModel(rows, pageable)

	if err = rows.Err(); err != nil {
		return nil, utils.PageInfo{}, err
	}

	m.Logger.InfoLog.Printf("Found %d contributions for search: %s", len(contributions), searchString)

	return contributions, pageInfo, nil
}

func (m *FundsModel) SearchByDateRange(startDate, endDate time.Time, pageable utils.Pageable) ([]*models.Fund, utils.PageInfo, error) {
	query := ""

	if endDate.IsZero() {
		query = `SELECT count(*) OVER(), id,receipt_no,total,organization_id,contribution_date,
						contributor,break_down,created_at,modified_at 
				FROM funds
				where organization_id = $1 AND contribution_date = $2
				ORDER BY created_at DESC LIMIT $3 OFFSET $4;`
	} else {
		query = `SELECT count(*) OVER(), id,receipt_no,total,organization_id,contribution_date,
						contributor,break_down,created_at,modified_at 
				FROM funds
				where organization_id = $1 AND contribution_date BETWEEN $2 AND $3
				ORDER BY created_at DESC LIMIT $4 OFFSET $5;`
	}

	var rows *sql.Rows
	var err error

	if endDate.IsZero() {
		rows, err = m.DB.Query(query, 1, startDate, pageable.Size, pageable.OffSet)
	} else {
		rows, err = m.DB.Query(query, 1, startDate, endDate, pageable.Size, pageable.OffSet)
	}

	if err != nil {
		return nil, utils.PageInfo{}, err
	}

	defer rows.Close()

	contributions, pageInfo := mapSqlRowsToModel(rows, pageable)

	if err = rows.Err(); err != nil {
		return nil, utils.PageInfo{}, err
	}

	return contributions, pageInfo, nil
}

func (m *FundsModel) GetSummary(startDate, endDate time.Time, organizationId int) ([]byte, error) {
	var stmt string

	if !startDate.IsZero() && endDate.IsZero() {
		stmt = `SELECT key as name, sum(value::jsonb::text::numeric) as value,contribution_date
			from funds, jsonb_each(funds.break_down)
			where organization_id = $1 AND contribution_date = $2 
			group by contribution_date,key
			order by contribution_date;`
	} else if !startDate.IsZero() && !endDate.IsZero() {
		stmt = `SELECT key as name, sum(value::jsonb::text::numeric) as value,contribution_date
			from funds, jsonb_each(funds.break_down)
			where organization_id = $1 AND contribution_date BETWEEN $2 AND $3
			group by contribution_date,key
			order by contribution_date;`
	} else {
		return nil, fmt.Errorf("start date and end date cannot be both empty")
	}

	var rows *sql.Rows
	var err error

	if endDate.IsZero() {
		rows, err = m.DB.Query(stmt, organizationId, startDate)
	} else {
		rows, err = m.DB.Query(stmt, organizationId, startDate, endDate)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	summary := make(map[string][]models.MonthlySummations, 0)
	data := []*models.MonthlySummations{}

	for rows.Next() {
		row := &models.MonthlySummations{}

		err := rows.Scan(&row.Category, &row.Total, &row.Date)

		if err != nil {
			return nil, err
		}

		data = append(data, row)
	}

	for _, row := range data {
		if value, ok := summary[row.Date]; ok {
			value = append(value, models.MonthlySummations{
				Category: row.Category,
				Date:     row.Date,
				Total:    row.Total,
			})
			summary[row.Date] = value
		} else {
			list := []models.MonthlySummations{}
			list = append(list, models.MonthlySummations{
				Category: row.Category,
				Date:     row.Date,
				Total:    row.Total,
			})

			summary[row.Date] = list
		}
	}

	file, err := m.GenerateExcelSummaryFile(summary)

	if err != nil {
		return nil, err
	}

	return file, nil
}

func (m *FundsModel) GetMonthlyStatistics(year, month, organizationId int) ([]*models.MonthlyStats, error) {
	stmt := `SELECT key as name, sum(value::jsonb::text::numeric) as value
				from 
				funds, jsonb_each(funds.break_down)
				where extract(year from contribution_date) = $1 and extract(month from contribution_date) = $2 and organization_id = $3
				group by key;`

	rows, err := m.DB.Query(stmt, year, month, 1)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	stats := []*models.MonthlyStats{}

	for rows.Next() {
		row := &models.MonthlyStats{}

		err := rows.Scan(&row.Name, &row.Value)

		if err != nil {
			return nil, err
		}

		stats = append(stats, row)
	}

	return stats, nil
}

func (m *FundsModel) GetMonthlyVariance(targetCategories []string) ([]*models.Variance, error) {
	stmt := `WITH previous AS (SELECT key as prev_category,sum(value::jsonb::text::numeric) as prev_total
				FROM funds, jsonb_each(funds.break_down) 
				WHERE extract(month from contribution_date) =
				extract(month from date_trunc('month', now() - interval '1' month))
				group by prev_category),
				current_val AS (SELECT key as category,sum(value::jsonb::text::numeric) as total
				FROM funds, jsonb_each(funds.break_down) 
				WHERE extract(month from funds.contribution_date) = extract(month from now())
				group by category)

				SELECT current_val.category,current_val.total,coalesce(previous.prev_total,0) prev_total,
				(current_val.total - coalesce(previous.prev_total,0)) as difference FROM
				current_val LEFT JOIN previous ON current_val.category = previous.prev_category;`

	rows, err := m.DB.Query(stmt)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	stats := []*models.StatisticalVariance{}

	for rows.Next() {
		row := &models.StatisticalVariance{}

		err := rows.Scan(&row.Category, &row.Total, &row.PreviousTotal, &row.Difference)

		if err != nil {
			return nil, err
		}

		stats = append(stats, row)
	}

	statistics := []*models.Variance{}

	for _, category := range targetCategories {
		variance := getTargetCategory(category, stats)

		if variance != nil {

			var percentage float64

			if variance.PreviousTotal == 0 {
				percentage = 100.0
			} else {
				percentage = ((variance.Total - variance.PreviousTotal) / math.Abs(variance.PreviousTotal)) * 100.0
			}

			direction := 0

			if percentage > 0 {
				direction = 1
			} else if percentage < 0 {
				direction = -1
			}

			vars := models.Variance{
				Category:     variance.Category,
				CurrentValue: variance.Total,
				Percentage:   float32(percentage),
				Direction:    int8(direction),
			}

			statistics = append(statistics, &vars)
		} else {
			vars := models.Variance{
				Category:     category,
				CurrentValue: 0,
				Percentage:   0.00,
				Direction:    0,
			}
			statistics = append(statistics, &vars)
		}
	}

	return statistics, nil
}

func (m *FundsModel) SaveCategories(tx *sql.Tx, ctx context.Context, categories []string) (int, error) {

	distinctCategories := makeCategoriesUnique(&categories, m.GetCategories())

	if len(distinctCategories) == 0 {
		return 0, nil
	}

	stmt := `INSERT INTO fund_categories(name) VALUES`

	for i, category := range categories {
		s := fmt.Sprintf("('%s')", category)

		if i != len(categories)-1 {
			s += ","
		}

		stmt += s
	}

	stmt += ";"

	result, err := tx.ExecContext(ctx, stmt)

	if err != nil {
		return 0, err
	}

	rowAffected, err := result.RowsAffected()

	if err != nil {
		return 0, err
	}

	return int(rowAffected), nil
}

func (m *FundsModel) GetCategories() []string {
	stmt := `SELECT DISTINCT(name) FROM public.fund_categories;`

	rows, err := m.DB.Query(stmt)

	if err != nil {
		utils.GetLoggerInstance().ErrorLog.Println("Error while fetching categories ", err)
		return nil
	}

	defer rows.Close()

	categories := make([]string, 0)

	for rows.Next() {
		value := ""
		err := rows.Scan(&value)
		if err != nil {
			utils.GetLoggerInstance().ErrorLog.Println("Error reading categories ", err)
			return nil
		}

		categories = append(categories, value)
	}

	return categories
}

func (m *FundsModel) SaveFileHash(hash, filename string, organizationId int) error {
	stmt := `INSERT INTO imports(hash,filename,organization_id) VALUES ($1,$2,$3);`

	result, err := m.DB.Exec(stmt, hash, filename, organizationId)

	if err != nil {
		return err
	}

	_, err = result.RowsAffected()

	if err != nil {
		return err
	}

	return nil
}

func makeCategoriesUnique(newCategories *[]string, existingCategories []string) []string {
	categoriesToSave := make([]string, 0)

	for _, category := range *newCategories {
		if !slices.Contains(existingCategories, category) {
			categoriesToSave = append(categoriesToSave, category)
		}
	}

	return categoriesToSave
}

func mapSqlRowsToModel(rows *sql.Rows, pageable utils.Pageable) ([]*models.Fund, utils.PageInfo) {
	contributions := []*models.Fund{}

	jsonb := make([]byte, 0)
	totalRecords := 0

	for rows.Next() {
		row := &models.Fund{}
		err := rows.Scan(&totalRecords, &row.ID, &row.ReceiptNo, &row.Total, &row.OrganizationId, &row.Date, &row.Contributor,
			&jsonb, &row.Audit.CreatedAt, &row.Audit.ModifiedAt)

		if err != nil {
			return nil, utils.PageInfo{}
		}

		err = json.Unmarshal(jsonb, &row.BreakDown)

		if err != nil {
			utils.GetLoggerInstance().ErrorLog.Println(err)
		}

		contributions = append(contributions, row)
	}

	pageInfo := utils.PageInfo{
		CurrentPage: pageable.Page,
		Size:        pageable.Size,
		TotalItems:  totalRecords,
		FirstPage:   0,
		LastPage:    int(math.Floor(float64(totalRecords) / float64(pageable.Size))),
	}

	return contributions, pageInfo
}

func getTargetCategory(target string, data []*models.StatisticalVariance) *models.StatisticalVariance {
	for _, stat := range data {
		if strings.EqualFold(target, stat.Category) {
			return stat
		}
	}

	return nil
}

func (app *FundsModel) ValidateTotalAndBreakDown(total float64, breakDown map[string]float64) bool {
	var sum float64

	for _, value := range breakDown {
		sum += value
	}

	return sum == total
}

func (m *FundsModel) GenerateExcelFile(contributions []*models.Fund) ([]byte, error) {

	categories := m.GetCategories()

	excelFile, err := m.ExcelExporter.GenerateExcelFile(contributions, categories)

	if err != nil {
		return nil, err
	}

	return excelFile, nil
}

func (m *FundsModel) GeneratePdfFile(contributions []*models.Fund, startDate, endDate time.Time) ([]byte, error) {
	categories := m.GetCategories()

	pdfFile, err := m.PdfExporter.GeneratePdfFile(contributions, categories, startDate, endDate)

	if err != nil {
		return nil, err
	}

	return pdfFile, nil
}

func (m *FundsModel) GenerateExcelSummaryFile(data map[string][]models.MonthlySummations) ([]byte, error) {
	categories := m.GetCategories()

	excelFile, err := m.ExcelExporter.GenerateExcelSummary(
		data,
		categories)

	if err != nil {
		return nil, err
	}

	return excelFile, nil
}
