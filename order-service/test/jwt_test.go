package test

import (
	"order_microservice/internal/domain"
	jasonwebtoken "order_microservice/internal/middleware/jwt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestVerifyJWT_Positive(t *testing.T) {
	// Create a test JWT token
	claims := &domain.JwtClaims{
		UserID: 123,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("ayam"))
	assert.NoError(t, err)

	// Test the VerifyJWT function
	parsedClaims, err := jasonwebtoken.VerifyJWT(tokenStr)
	assert.NoError(t, err)
	assert.Equal(t, claims.UserID, parsedClaims.UserID)
}

func TestVerifyJWT_Negative(t *testing.T) {
	// Test with an invalid JWT token
	tokenStr := "invalidtoken"
	_, err := jasonwebtoken.VerifyJWT(tokenStr)
	assert.Error(t, err)

	// Test with an expired JWT token
	expiredClaims := &domain.JwtClaims{
		UserID: 123,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenStr, _ := expiredToken.SignedString([]byte("ayam"))

	_, err = jasonwebtoken.VerifyJWT(expiredTokenStr)
	assert.Error(t, err)
}
