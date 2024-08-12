package middleware

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var key = []byte("ayam")

type Claimsasdasda struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateJwt(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, Claimsasdasda{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "DIDI KEREN!",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	})

	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
