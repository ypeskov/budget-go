package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/dto"
	appErrors "ypeskov/budget-go/internal/errors"
	"ypeskov/budget-go/internal/logger"
	"ypeskov/budget-go/internal/middleware"
	"ypeskov/budget-go/internal/services"
	"ypeskov/budget-go/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/api/idtoken"

	"github.com/labstack/echo/v4"
)

type JWTCustomClaims struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}
type ProfileDTO struct {
	ID           int               `json:"id"`
	FirstName    string            `json:"first_name"`
	LastName     string            `json:"last_name"`
	Email        string            `json:"email"`
	Settings     map[string]string `json:"settings"`
	BaseCurrency string            `json:"baseCurrency"`
}

var (
	cfg *config.Config
	sm  *services.Manager
)

func RegisterAuthRoutes(g *echo.Group, cfgGlobal *config.Config, manager *services.Manager) {
	cfg = cfgGlobal
	sm = manager

	g.POST("/login", LoginUser)
	g.POST("/register", RegisterUser)
	g.POST("/oauth", OAuth)
	g.GET("/activate/:token", ActivateUser)
	g.GET("/profile", Profile)
}

func LoginUser(c echo.Context) error {
	logger.Debug("LoginUser request started", "method", c.Request().Method, "url", c.Request().URL)

	loginDTO := new(dto.UserLoginDTO)
	if err := c.Bind(loginDTO); err != nil {
		logger.Error("Error binding login data", "error", err)
		return c.String(http.StatusBadRequest, "Bad request")
	}

	user, err := sm.UserService.LoginUser(loginDTO)
	if err != nil {
		switch {
		// group common "unauthorized" errors
		case errors.As(err, new(*appErrors.UserNotFoundError)),
			errors.As(err, new(*appErrors.InvalidCredentialsError)):
			logger.Warn("Login failed", "error", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"detail": "Invalid email or password",
			})

		case errors.As(err, new(*appErrors.UserNotActivatedError)):
			logger.Warn("User not activated", "error", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"detail": "User not activated",
			})

		case errors.As(err, new(*appErrors.UserDeletedError)):
			logger.Warn("User is deleted", "error", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"detail": "User account is deleted",
			})

		default:
			logger.Error("Unexpected login error", "error", err)
			return c.JSON(http.StatusInternalServerError, "Internal server error")
		}
	}

	signedToken, err := utils.GenerateAccessToken(user, cfg)
	if err != nil {
		logger.Error("Error generating token", "error", err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	logger.Debug("LoginUser request completed", "method", c.Request().Method, "url", c.Request().URL)
	return c.JSON(http.StatusOK, map[string]string{
		"accessToken": signedToken,
		"tokenType":   "Bearer",
	})
}

func RegisterUser(c echo.Context) error {
	logger.Debug("RegisterUser request started", "method", c.Request().Method, "url", c.Request().URL)

	u := new(dto.UserRegisterRequestDTO)
	if err := c.Bind(u); err != nil {
		logger.Error("Error binding registration data", "error", err)
		return c.String(http.StatusBadRequest, "Bad request")
	}

	// Basic validation
	if u.Email == "" || u.Password == "" || u.FirstName == "" || u.LastName == "" {
		logger.Error("Missing required registration fields")
		return c.String(http.StatusBadRequest, "All fields are required")
	}

	// RegisterUser user using service layer (handles complete registration flow)
	createdUser, err := sm.UserService.RegisterUser(u, sm.CurrenciesService, sm.ActivationTokenService)
	if err != nil {
		if strings.Contains(err.Error(), "User already exists") {
			logger.Error("User already exists with email", "email", u.Email)
			return c.String(http.StatusConflict, "User already exists")
		}
		logger.Error("Error registering user", "error", err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	// Create response DTO
	response := dto.UserRegisterResponseDTO{
		ID:        createdUser.ID,
		Email:     createdUser.Email,
		FirstName: createdUser.FirstName,
		LastName:  createdUser.LastName,
	}

	logger.Debug("RegisterUser request completed", "method", c.Request().Method, "url", c.Request().URL)
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"user":    response,
		"message": "User registered successfully. Please check your email to activate your account.",
	})
}

func ActivateUser(c echo.Context) error {
	logger.Debug("ActivateUser request started", "method", c.Request().Method, "url", c.Request().URL)

	// Extract token from URL parameter
	token := c.Param("token")
	if token == "" {
		logger.Error("Missing activation token in URL")
		return c.String(http.StatusBadRequest, "Activation token is required")
	}

	// Validate and use the activation token
	activationToken, err := sm.ActivationTokenService.ValidateAndUseToken(token)
	if err != nil {
		logger.Error("Error validating activation token", "error", err)
		return c.String(http.StatusBadRequest, "Invalid or expired activation token")
	}

	// Activate the user
	err = sm.UserService.ActivateUser(activationToken.UserID)
	if err != nil {
		logger.Error("Error activating user", "error", err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	logger.Debug("ActivateUser request completed - GET /auth/activate/:token")
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Account activated successfully",
	})
}

func Profile(c echo.Context) error {
	logger.Debug("Profile request started", "method", c.Request().Method, "url", c.Request().URL)

	cfg := c.Get("config").(*config.Config)

	// Accept both Authorization: Bearer and legacy auth-token header
	var token string
	if authz := c.Request().Header.Get("Authorization"); authz != "" {
		if parts := strings.SplitN(authz, " ", 2); len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			token = parts[1]
		}
	}
	if token == "" {
		token = c.Request().Header.Get("auth-token")
	}

	claims, err := middleware.GetUserFromToken(token, cfg)
	if err != nil || claims == nil {
		logger.Error("Failed to cast user to jwt.MapClaims")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or missing user")
	}

	email, emailOk := claims["email"].(string)
	if !emailOk {
		logger.Error("Email not found in claims")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user data")
	}

	user, err := sm.UserService.GetUserByEmail(email)
	if err != nil {
		logger.Error("Error getting user by email", "error", err)
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}
	fmt.Printf("user %+v\n", user)

	// Get user's actual base currency
	baseCurrency, err := sm.UserSettingsService.GetBaseCurrency(user.ID)
	if err != nil {
		logger.Error("Error getting user base currency", "error", err)
		// Fallback to USD if we can't get the base currency
		baseCurrency.Code = "USD"
	}

	// Get user settings including language
	settings := map[string]string{}
	userSettings, err := sm.UserSettingsService.GetUserSettings(user.ID)
	if err != nil {
		logger.Error("Error getting user settings", "error", err)
		// Fallback to default language if we can't get settings
		settings["language"] = "en"
	} else {
		// Extract language from settings, default to "en" if not found
		if lang, ok := userSettings.Settings["language"].(string); ok {
			settings["language"] = lang
		} else {
			settings["language"] = "en"
		}
	}

	logger.Debug("Profile request completed - GET /auth/profile")
	return c.JSON(http.StatusOK, ProfileDTO{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		Settings:     settings,
		BaseCurrency: baseCurrency.Code,
	})
}

func OAuth(c echo.Context) error {
	logger.Debug("OAuth request started", "method", c.Request().Method, "url", c.Request().URL)

	oauthToken := new(dto.OAuthToken)
	if err := c.Bind(oauthToken); err != nil {
		logger.Error("Bad request", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"detail": "Bad request"})
	}

	ctx := context.Background()
	payload, err := idtoken.Validate(ctx, oauthToken.Credential, cfg.GoogleClientID)
	if err != nil {
		logger.Error("Error verifying Google JWT", "error", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"detail": "Invalid token"})
	}

	email, emailOk := payload.Claims["email"].(string)
	if !emailOk || email == "" {
		logger.Error("No email provided in token")
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"detail": "No email provided"})
	}

	emailVerified, emailVerifiedOk := payload.Claims["email_verified"].(bool)
	if !emailVerifiedOk || !emailVerified {
		logger.Error("Error while logging in user: Email not verified", "email", email)
		return c.JSON(http.StatusUnauthorized, map[string]string{"detail": "Email not verified"})
	}

	givenName, _ := payload.Claims["given_name"].(string)
	familyName, _ := payload.Claims["family_name"].(string)
	if familyName == "" {
		familyName = ""
	}

	user, err := sm.UserService.LoginOrRegisterOAuth(email, givenName, familyName)
	if err != nil {
		if _, ok := err.(*appErrors.UserNotActivatedError); ok {
			logger.Error("Error while logging in user", "email", email, "error", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"detail": "User not activated"})
		}
		logger.Error("Error while logging in user", "email", email, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": "Internal server error. See logs for details"})
	}

	signedToken, err := utils.GenerateAccessToken(user, cfg)
	if err != nil {
		logger.Error("Error generating access token", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": "Internal server error. See logs for details"})
	}

	logger.Debug("OAuth request completed - POST /auth/oauth")
	return c.JSON(http.StatusOK, dto.Token{
		AccessToken: signedToken,
		TokenType:   "Bearer",
	})
}
