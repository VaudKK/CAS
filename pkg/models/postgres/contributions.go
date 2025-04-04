package postgres

import (
	"database/sql"
	"fmt"

	"github.com/VaudKK/CAS/pkg/models"
	"github.com/VaudKK/CAS/utils"
)

type ContributionModel struct {
	DB *sql.DB
}


func  (m *ContributionModel) Insert(contributions []models.Contribution) (int, error){
	stmt := `INSERT INTO CONTRIBUTIONS(category,amount,organization_id,date,contributor) VALUES`

	for i , contribution := range contributions {
		s := fmt.Sprintf("(%s,%.2f,%d,%s,%s)",contribution.Category,contribution.Amount,contribution.OrganizationId,contribution.Date,
						contribution.Contributor)
		
		if i != len(contributions) -1 {
			s += ","
		}
	}

	result, err := m.DB.Exec(stmt)

	if err != nil {
		return 0,err
	}

	rowAffected, err := result.RowsAffected()

	if err != nil {
		return 0,err
	}

	return int(rowAffected),nil

}

func (m *ContributionModel) Get(organizationId int,pageable utils.Pageable) ([]models.Contribution,error){
	return nil,nil
}

