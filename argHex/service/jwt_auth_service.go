package service

import (
	"log"
	"time"

	"github.com/argSea/portfolio_blog_api/argHex/data_objects"
	"github.com/argSea/portfolio_blog_api/argHex/in_port"
	"github.com/golang-jwt/jwt/v5"
)

type jwtAuthService struct {
	jwtSecret []byte
}

func NewJWTAuthService(secret []byte) in_port.AuthService {
	return jwtAuthService{
		jwtSecret: secret,
	}
}

// NewAuth
func (j jwtAuthService) Generate(id string, expires time.Time, roles []string) (string, error) {
	// create jwt token
	key := j.jwtSecret
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["userID"] = id

	time_now := time.Now()
	time_30_days := time_now.AddDate(0, 0, 30)
	claims["exp"] = time_30_days.Unix()
	claims["iat"] = time_now.Unix()
	claims["role"] = "user"

	// later we can add a claim to make my user an admin
	// claims["role"] = "admin"
	tokenString, _ := token.SignedString(key)

	return tokenString, nil
}

// Validate
func (j jwtAuthService) Validate(token string) (data_objects.AuthValidationResponseObject, error) {
	// parse jwt
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return j.jwtSecret, nil
	})

	if nil != err {
		log.Println("Error parsing token: ", err)
		return data_objects.AuthValidationResponseObject{Valid: false}, err
	}

	validResponse := data_objects.AuthValidationResponseObject{
		Valid:  true,
		Role:   claims["role"].(string),
		UserID: claims["userID"].(string),
	}

	return validResponse, nil
}

// check if user is authorized
func (j jwtAuthService) IsAuthorized(id string, token string, roles ...string) bool {
	validResponse, err := j.Validate(token)

	if nil != err {
		log.Println("Error validating token: ", err)
		return false
	}

	if !validResponse.Valid {
		return false
	}

	role := validResponse.Role
	userID := validResponse.UserID

	if role != in_port.PERM_ADMIN {
		if userID != id {
			return false
		}
	}

	// check if user has required role
	for _, r := range roles {
		if role == r {
			return true
		}
	}

	return false
}
