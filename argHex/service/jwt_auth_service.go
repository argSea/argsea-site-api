package service

import (
	"errors"
	"log"
	"time"

	"github.com/argSea/argsea-site-api/argHex/data_objects"
	"github.com/argSea/argsea-site-api/argHex/in_port"
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

// Generate mints a signed HS256 token for the user, honoring the requested
// expiry and role; pass in_port.PERM_ADMIN to mint an admin token.
func (j jwtAuthService) Generate(id string, expires time.Time, roles []string) (string, error) {
	// create jwt token
	key := j.jwtSecret
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["userID"] = id
	claims["exp"] = expires.Unix()
	claims["iat"] = time.Now().Unix()

	// the token carries a single effective role; default to plain user when none given
	role := in_port.PERM_USER

	if 0 < len(roles) {
		role = roles[0]
	}

	claims["role"] = role

	tokenString, sign_err := token.SignedString(key)

	if nil != sign_err {
		return "", sign_err
	}

	return tokenString, nil
}

// Validate
func (j jwtAuthService) Validate(token string) (data_objects.AuthValidationResponseObject, error) {
	// parse jwt
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return j.jwtSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if nil != err {
		log.Println("Error parsing token: ", err)
		return data_objects.AuthValidationResponseObject{Valid: false}, err
	}

	// a validly-signed token can still be missing the claims we depend on; guard
	// the assertions so a malformed token is rejected rather than panicking
	role, roleOK := claims["role"].(string)
	userID, idOK := claims["userID"].(string)

	if !roleOK || !idOK {
		log.Println("Error validating token: missing role or userID claim")
		return data_objects.AuthValidationResponseObject{Valid: false}, errors.New("token missing required claims")
	}

	validResponse := data_objects.AuthValidationResponseObject{
		Valid:  true,
		Role:   role,
		UserID: userID,
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
