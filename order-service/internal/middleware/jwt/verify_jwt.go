package jasonwebtoken

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"order_microservice/internal/domain"
)

var key = []byte("ayam")

func VerifyJWT(tokenStr string) (*domain.JwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &domain.JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signin method: %v", token.Header["alg"])
		}
		return key, nil
	})

	if err != nil {
		return nil, err
	}

	klem := token.Claims.(*domain.JwtClaims)
	fmt.Println(klem.Username)

	return klem, nil
}
