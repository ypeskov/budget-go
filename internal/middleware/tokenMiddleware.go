package middleware

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

var jwtSecret = []byte("secret")

func GetUserFromToken(authToken string) (jwt.MapClaims, error) {
	if authToken == "" {
		return nil, errors.New("missing auth-token")
	}

	token, err := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().After(time.Unix(int64(exp), 0)) {
				return nil, errors.New("token has expired")
			}
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func CheckJWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authToken := c.Request().Header.Get("auth-token")

		claims, err := GetUserFromToken(authToken)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": err.Error(),
			})
		}

		c.Set("user", claims)

		return next(c)
	}
}