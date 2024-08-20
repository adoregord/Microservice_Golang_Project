package jasonwebtoken

import (
	"fmt"
	"order_microservice/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var key = []byte("ayam")

type Claimsasdasda struct {
	UserID   int    `json:"userID"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

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

	return klem, nil
}

func CreateMockJWT() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, Claimsasdasda{
		UserID:   1,
		Username: "Didi",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "DIDI KEREN!",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Hour)),
		},
	})

	//token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//tokenString, _ := token.SignedString([]byte("ayam")) // Sama dengan key yang digunakan dalam VerifyJWT
	tokenString, err := token.SignedString(key)
	if err != nil {
		return ""
	}
	return tokenString
}
