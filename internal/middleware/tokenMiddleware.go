package middleware

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

// Your secret key for signing the JWTs
var jwtSecret = []byte("secret")

func CheckJWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authToken := c.Request().Header.Get("auth-token")
		if authToken == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Missing auth-token header",
			})
		}

		token, err := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Invalid or expired token",
			})
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().After(time.Unix(int64(exp), 0)) {
					return c.JSON(http.StatusUnauthorized, map[string]string{
						"message": "Token has expired",
					})
				}
			}

			c.Set("user", claims)

			return next(c)
		}

		return c.JSON(http.StatusUnauthorized, map[string]string{
			"message": "Invalid token",
		})
	}
}
