package postgres

import (
	"database/sql"

	"github.com/VaudKK/CAS/pkg/data"
	"github.com/VaudKK/CAS/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) CreateUser(user *models.User) (int, error) {

	stmt := `INSERT INTO users (username, email, organization_id, password,active,verified) VALUES ($1, $2, $3, $4, $5,$6)`

	hashPassword, err := hashPassword(user.Password)

	if err != nil {
		return 0, err
	}

	result, err := m.DB.Exec(stmt, user.UserName, user.Email, user.OrganizationId, hashPassword, true, false)

	if err != nil {
		return 0, err
	}

	rowAffected, err := result.RowsAffected()

	if err != nil {
		return 0, err
	}

	return int(rowAffected), nil
}

func (m *UserModel) ValidateUser(username, password string) (bool, error) {
	stmt := `SELECT email,password FROM users WHERE email = $1 AND active = true`

	row, err := m.DB.Query(stmt, username)

	if err != nil {
		return false, err
	}

	defer row.Close()

	if row.Next() {
		var email, passwordHash string
		err = row.Scan(&email, &passwordHash)

		if err != nil {
			return false, err
		}

		if checkPassword(password, passwordHash) {
			return true, nil
		}
	}

	return false, nil
}

func (m *UserModel) VerifyUser(email string) error {
	stmt := `UPDATE users SET verified = true WHERE email = $1;`

	_, err := m.DB.Exec(stmt, email)

	if err != nil {
		return err
	}

	return nil
}

func (m *UserModel) GetUserByEmail(email string) (*models.User, error) {
	stmt := `SELECT id, username FROM users WHERE email = $1`

	row, err := m.DB.Query(stmt, email)

	if err != nil {
		return nil, err
	}

	defer row.Close()

	var user models.User

	if row.Next() {
		err = row.Scan(&user.ID, &user.UserName)
		if err != nil {
			return nil, err
		}
		return &user, nil
	} else {
		return nil, data.ErrorNoRecords
	}
}

func (m *UserModel) GetUserID(id int) (*models.User, error) {
	stmt := `SELECT id, username, email, verified, active FROM users WHERE id = $1`

	row, err := m.DB.Query(stmt, id)

	if err != nil {
		return nil, err
	}

	defer row.Close()

	var user models.User

	if row.Next() {
		err = row.Scan(&user.ID, &user.UserName, &user.Email, &user.Verified, &user.Active)
		if err != nil {
			return nil, err
		}
		return &user, nil
	} else {
		return nil, data.ErrorNoRecords
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
