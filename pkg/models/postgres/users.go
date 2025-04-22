package postgres

import (
	"database/sql"
	"errors"

	"github.com/VaudKK/CAS/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	DB *sql.DB
}


func (m *UserModel) CreateUser(user *models.User)(int,error){

	usr,err := m.getUserByEmail(user.Email)

	if err != nil {
		return 0,err
	}

	if usr != nil {
		return 0, errors.New("user already exists")
	}

	stmt := `INSERT INTO users (username, email, organization_id, password,active) VALUES ($1, $2, $3, $4, $5)`

	hashPassword, err := hashPassword(user.Password)

	if err != nil {
		return 0, err
	}

	result,err := m.DB.Exec(stmt,user.UserName,user.Email,user.OrganizationId,hashPassword,true)

	if err != nil {
		return 0,err
	}

	rowAffected, err := result.RowsAffected()

	if err != nil {
		return 0, err
	}

	return int(rowAffected), nil
}

func (m *UserModel) VerifyUser(username,password string) (bool,error){
	stmt := `SELECT email,password FROM users WHERE email = $1 AND active = true`

	row,err := m.DB.Query(stmt,username)

	if err != nil {
		return false, err
	}

	defer row.Close()

	if row.Next() {
		var email,passwordHash string
		err = row.Scan(&email,&passwordHash)

		if err != nil {
			return false, err
		}

		if checkPassword(password,passwordHash) {
			return true, nil
		}
	}

	return false, nil
}

func (m *UserModel) getUserByEmail(email string)(*models.User,error){
	stmt :=  `SELECT id, username FROM users WHERE email = $1`

	row,err := m.DB.Query(stmt,email)

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
	}else{
		return nil, nil
	}
}

func hashPassword (password string)(string,error){
	bytes, err := bcrypt.GenerateFromPassword([]byte(password),14)
	return string(bytes),err
}

func checkPassword(password,hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash),[]byte(password))
	return err == nil
}