package middleware

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)


func VerifyJWT(tokenStr string) (*Claimsasdasda, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claimsasdasda{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signin method: %v", token.Header["alg"])
		}
		return key, nil
	})

	if err != nil {
		return nil, err
	}

	klem := token.Claims.(*Claimsasdasda)
	fmt.Println(klem.Username)

	return klem, nil
}
