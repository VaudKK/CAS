package data

import (
	"fmt"
	"os"
	"time"

	"github.com/VaudKK/CAS/pkg/models"
	"github.com/VaudKK/CAS/utils"
	"github.com/golang-jwt/jwt/v5"
)

var AnonymousUser = &models.User{}

type Token struct {
	Token  string `json:"token"`
	Expiry int64  `json:"expiry"`
}

type CustomClaims struct {
	Identifier int
	Exp int64
	jwt.RegisteredClaims
}

var logger = utils.GetLoggerInstance()

var secretKey = []byte(os.Getenv("TOKEN_KEY"))

func CreateToken(userID int) (Token, error) {
	expiry := time.Now().Add(time.Hour * 1).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS512,
		jwt.MapClaims{
			"identifier": userID,
			"exp":        expiry,
		})

	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		logger.ErrorLog.Println(err)
		return Token{}, err
	}

	return Token{
		Token:  tokenString,
		Expiry: expiry,
	}, nil
}

func VerifyToken(tokenString string) (int,error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{},func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		logger.ErrorLog.Println(err)
		return -1,err
	}

	if claim,ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claim.Identifier,nil
	}else{
		return -1, fmt.Errorf("invalid token")
	}
}


// Check if a User instance is the AnonymousUser.
func IsAnonymous(u *models.User) bool {
	return u == AnonymousUser
}