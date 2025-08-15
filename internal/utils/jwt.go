package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/models"
)

type JWTCustomClaims struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	ExpHuman string `json:"exp_human"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a new JWT token for the given user
func GenerateAccessToken(user *models.User, cfg *config.Config) (string, error) {
	expirationTime := time.Now().Add(time.Hour * 1)

	claims := &JWTCustomClaims{
		Id:       user.ID,
		Email:    user.Email,
		ExpHuman: expirationTime.Format("2006-01-02 15:04:05"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.SecretKey))
}
