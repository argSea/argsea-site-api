package utility

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if nil != err {
		log.Printf("Error hashing password: %v", err)
		return "", err
	}

	return string(bcryptPassword), nil
}
