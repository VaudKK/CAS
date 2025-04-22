package data

import (
	"errors"

	"github.com/VaudKK/CAS/pkg/validator"
)

var (
	ErrorNoRecords = errors.New("record not found")
)

func ValidateEmail(v *validator.Validator,email string){
	v.Check(email != "","email","must be provided")
	v.Check(validator.Matches(email,validator.EmailRX),"email","must be a valid email address")
}

func ValidatePasswordPlainText(v *validator.Validator,password string){
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUserName(v *validator.Validator,username string){
	v.Check(username != "","username","must be provided")
	v.Check(validator.Matches(username,validator.UserNameRX),"username", "must be words and spaces only with a max length of 50")
}