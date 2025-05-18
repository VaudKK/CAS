package postgres

import (
	"crypto/rand"
	"database/sql"
	"math/big"
	"strings"
	"time"

	"github.com/VaudKK/CAS/pkg/mailer"
	"github.com/VaudKK/CAS/pkg/models"
	"github.com/google/uuid"
)

type OtpModel struct {
	DB     *sql.DB
	Mailer *mailer.Mailer
	User   *UserModel
}

func (m *OtpModel) SendOtp(subject, mode string) (*models.Otp, error) {
	stmt := `INSERT INTO otp(subject,verification_mode,otp,expiry,used,session_id) VALUES($1,$2,$3,$4,$5,$6);`

	otp, err := generateOtp(6)

	if err != nil {
		return nil, err
	}

	now := time.Now()
	expiry := now.Add(time.Minute * 30)
	sessionId := strings.ReplaceAll(uuid.New().String(), "-", "")

	_, err = m.DB.Exec(stmt, subject, mode, string(otp), expiry, false, sessionId)

	if err != nil {
		return nil, err
	}

	var emailData struct {
		Otp string
	}

	emailData.Otp = otp

	//send mail
	err = m.Mailer.Send(subject, "user_otp.tmpl", emailData)

	if err != nil {
		return nil, err
	}

	return &models.Otp{
		Expiry:    expiry,
		SessionId: sessionId,
	}, nil

}

func (m *OtpModel) VerifyOtp(otp, sessionId string) (bool, error) {
	stmt := `SELECT otp,session_id,expiry,subject FROM otp WHERE session_id = $1 AND used = false;`

	rows, err := m.DB.Query(stmt, sessionId)

	if err != nil {
		return false, err
	}

	defer rows.Close()

	var data struct {
		sessionId string
		expiry    time.Time
		otp       string
		subject   string
	}

	for rows.Next() {
		err = rows.Scan(&data.otp, &data.sessionId, &data.expiry, &data.subject)

		if err != nil {
			return false, err
		}
	}

	if data.otp != otp {
		return false, nil
	}

	if data.expiry.Before(time.Now()) {
		return false, nil
	}

	updateStmt := `UPDATE otp SET used = $1 WHERE session_id = $2`

	_, err = m.DB.Exec(updateStmt, true, sessionId)

	if err != nil {
		return false, err
	}

	err = m.User.VerifyUser(data.subject)

	if err != nil {
		return false, err
	}

	return true, nil
}

func generateOtp(length int) (string, error) {
	const chars = "0123456789"
	result := ""

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		result += string(chars[num.Int64()])
	}

	return result, nil
}
