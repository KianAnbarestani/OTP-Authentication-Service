package services

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	secret []byte
}

func NewAuthService(secret string) *AuthService {
	return &AuthService{secret: []byte(secret)}
}

func (a *AuthService) GenerateToken(userID uint, phone string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"phone": phone,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.secret)
}
