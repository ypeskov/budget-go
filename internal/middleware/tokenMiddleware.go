package middleware

import (
    "errors"
    "net/http"
    "strings"
    "time"
    "ypeskov/budget-go/internal/config"
    "ypeskov/budget-go/internal/services"

    "github.com/golang-jwt/jwt/v5"
    "github.com/labstack/echo/v4"
)

// GetUserFromToken parses and validates the JWT token, returning claims if valid.
func GetUserFromToken(authToken string, cfg *config.Config) (jwt.MapClaims, error) {
	if authToken == "" {
		return nil, errors.New("missing auth-token")
	}

	// Parse the token
	token, err := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.SecretKey), nil
	})
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	// Validate token and extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check expiration
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().After(time.Unix(int64(exp), 0)) {
				return nil, errors.New("token has expired")
			}
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// AuthMiddleware validates the JWT token, retrieves the user, and sets both claims and user in context.
func AuthMiddleware(sm *services.Manager, cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
            // Extract the token from headers: prefer Authorization: Bearer, fallback to auth-token
            var authToken string
            if authz := c.Request().Header.Get("Authorization"); authz != "" {
                parts := strings.SplitN(authz, " ", 2)
                if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
                    authToken = parts[1]
                }
            }
            if authToken == "" {
                authToken = c.Request().Header.Get("auth-token")
            }
            if authToken == "" {
                return c.JSON(http.StatusUnauthorized, map[string]string{
                    "message": "Missing token",
                })
            }

			// Parse and validate the token
			claims, err := GetUserFromToken(authToken, cfg)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": err.Error(),
				})
			}

			// Retrieve the user from the claims
			email, ok := claims["email"].(string)
			if !ok || email == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "Invalid token payload",
				})
			}

			user, err := sm.UserService.GetUserByEmail(email)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"message": "Failed to retrieve user",
				})
			}

			// Save claims and user to the context
			// c.Set("user", claims)
			c.Set("authenticated_user", user)

			// Continue to the next handler
			return next(c)
		}
	}
}
