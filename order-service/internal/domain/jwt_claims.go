package domain

import "github.com/golang-jwt/jwt/v5"

type JwtClaims struct {
	UserID   int    `json:"userID"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
