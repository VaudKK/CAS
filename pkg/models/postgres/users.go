package postgres

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/VaudKK/CAS/pkg/data"
	"github.com/VaudKK/CAS/pkg/mailer"
	"github.com/VaudKK/CAS/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	DB     *sql.DB
	Mailer *mailer.Mailer
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

func (m *UserModel) SendResetLink(email string) error {
	_, err := m.GetUserByEmail(strings.TrimSpace(email))

	if err != nil {
		return err
	}

	stmt := `INSERT INTO reset_links (email,expiry,reset_token) VALUES ($1,$2,$3)`

	resetToken, _ := randomString(50)

	_, err = m.DB.Exec(stmt, email, time.Now().Add(time.Minute*30), resetToken)
	if err != nil {
		return err
	}

	var resetData struct {
		ResetLink string
	}

	resetData.ResetLink = "http://localhost:3000/auth/reset?token=" + resetToken

	//send mail
	_ = m.Mailer.Send(strings.TrimSpace(email), "user_reset_link.tmpl", resetData)

	return nil
}

func (m *UserModel) ChangePassword(resetToken, newPassword string) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	stmt := `SELECT email,expiry FROM reset_links WHERE reset_token = $1 AND is_used = false`

	row, err := m.DB.Query(stmt, resetToken)

	if err != nil {
		return err
	}

	defer row.Close()

	var response struct {
		email  string
		expiry time.Time
	}

	if row.Next() {
		err = row.Scan(&response.email, &response.expiry)
		if err != nil {
			return err
		}

		if response.expiry.Before(time.Now()) {
			return fmt.Errorf("reset link expired")
		}
	} else {
		return fmt.Errorf("invalid reset token")
	}

	hashPassword, err := hashPassword(newPassword)

	if err != nil {
		return err
	}

	stmt = `UPDATE users SET password = $1 WHERE email = $2`

	_, err = m.DB.Exec(stmt, hashPassword, response.email)

	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	stmt = `UPDATE reset_links SET is_used = true WHERE reset_token = $1`
	_, err = m.DB.Exec(stmt, resetToken)

	if err != nil {
		return fmt.Errorf("failed to mark reset link as used: %w", err)
	}

	//send mail
	_ = m.Mailer.Send(strings.TrimSpace(response.email), "password_change.tmpl", nil)

	return tx.Commit()
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func randomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[randomIndex.Int64()]
	}
	return string(b), nil
}
