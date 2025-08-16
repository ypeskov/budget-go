package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/middleware"
	"ypeskov/budget-go/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/api/idtoken"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/services"
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
	log.Debugf("LoginUser request started: %s %s", c.Request().Method, c.Request().URL)

	loginDTO := new(dto.UserLoginDTO)
	if err := c.Bind(loginDTO); err != nil {
		log.Error("Error binding login data: ", err)
		return c.String(http.StatusBadRequest, "Bad request")
	}

	user, err := sm.UserService.LoginUser(loginDTO)
	if err != nil {
		// Handle different error types appropriately
		var userNotFoundError *services.UserNotFoundError
		var invalidCredentialsError *services.InvalidCredentialsError
		var userNotActivatedError *services.UserNotActivatedError
		var userDeletedError *services.UserDeletedError
		switch {
		case errors.As(err, &userNotFoundError), errors.As(err, &invalidCredentialsError):
			log.Error("Login failed: ", err)
			return c.String(http.StatusUnauthorized, "Unauthorized")
		case errors.As(err, &userNotActivatedError):
			log.Error("User not activated: ", err)
			return c.String(http.StatusUnauthorized, "User not activated")
		case errors.As(err, &userDeletedError):
			log.Error("User is deleted: ", err)
			return c.String(http.StatusUnauthorized, "User account is disabled")
		default:
			log.Error("Login error: ", err)
			return c.String(http.StatusInternalServerError, "Internal server error")
		}
	}

	signedToken, err := utils.GenerateAccessToken(user, cfg)
	if err != nil {
		log.Error("Error generating token: ", err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	log.Debug("LoginUser request completed: %s %s", c.Request().Method, c.Request().URL)
	return c.JSON(http.StatusOK, map[string]string{
		"accessToken": signedToken,
		"tokenType":   "Bearer",
	})
}

func RegisterUser(c echo.Context) error {
	log.Debugf("RegisterUser request started: %s %s", c.Request().Method, c.Request().URL)

	u := new(dto.UserRegisterRequestDTO)
	if err := c.Bind(u); err != nil {
		log.Error("Error binding registration data: ", err)
		return c.String(http.StatusBadRequest, "Bad request")
	}

	// Basic validation
	if u.Email == "" || u.Password == "" || u.FirstName == "" || u.LastName == "" {
		log.Error("Missing required registration fields")
		return c.String(http.StatusBadRequest, "All fields are required")
	}

	// RegisterUser user using service layer (handles complete registration flow)
	createdUser, err := sm.UserService.RegisterUser(u, sm.CurrenciesService, sm.ActivationTokenService)
	if err != nil {
		if strings.Contains(err.Error(), "User already exists") {
			log.Error("User already exists with email: ", u.Email)
			return c.String(http.StatusConflict, "User already exists")
		}
		log.Error("Error registering user: ", err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	// Create response DTO
	response := dto.UserRegisterResponseDTO{
		ID:        createdUser.ID,
		Email:     createdUser.Email,
		FirstName: createdUser.FirstName,
		LastName:  createdUser.LastName,
	}

	log.Debug("RegisterUser request completed: %s %s", c.Request().Method, c.Request().URL)
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"user":    response,
		"message": "User registered successfully. Please check your email to activate your account.",
	})
}

func ActivateUser(c echo.Context) error {
	log.Debugf("ActivateUser request started: %s %s", c.Request().Method, c.Request().URL)

	// Extract token from URL parameter
	token := c.Param("token")
	if token == "" {
		log.Error("Missing activation token in URL")
		return c.String(http.StatusBadRequest, "Activation token is required")
	}

	// Validate and use the activation token
	activationToken, err := sm.ActivationTokenService.ValidateAndUseToken(token)
	if err != nil {
		log.Error("Error validating activation token: ", err)
		return c.String(http.StatusBadRequest, "Invalid or expired activation token")
	}

	// Activate the user
	err = sm.UserService.ActivateUser(activationToken.UserID)
	if err != nil {
		log.Error("Error activating user: ", err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	log.Debug("ActivateUser request completed - GET /auth/activate/:token")
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Account activated successfully",
	})
}

func Profile(c echo.Context) error {
	log.Debugf("Profile request started: %s %s", c.Request().Method, c.Request().URL)

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
		log.Error("Failed to cast user to jwt.MapClaims")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or missing user")
	}

	email, emailOk := claims["email"].(string)
	if !emailOk {
		log.Error("Email not found in claims")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user data")
	}

	user, err := sm.UserService.GetUserByEmail(email)
	if err != nil {
		log.Error("Error getting user by email: ", err)
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}
	fmt.Printf("user %+v\n", user)

	// Get user's actual base currency
	baseCurrency, err := sm.UserSettingsService.GetBaseCurrency(user.ID)
	if err != nil {
		log.Error("Error getting user base currency: ", err)
		// Fallback to USD if we can't get the base currency
		baseCurrency.Code = "USD"
	}

	// Get user settings including language
	settings := map[string]string{}
	userSettings, err := sm.UserSettingsService.GetUserSettings(user.ID)
	if err != nil {
		log.Error("Error getting user settings: ", err)
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

	log.Debug("Profile request completed - GET /auth/profile")
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
	log.Debugf("OAuth request started: %s %s", c.Request().Method, c.Request().URL)

	oauthToken := new(dto.OAuthToken)
	if err := c.Bind(oauthToken); err != nil {
		log.Error("Bad request: ", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"detail": "Bad request"})
	}

	ctx := context.Background()
	payload, err := idtoken.Validate(ctx, oauthToken.Credential, cfg.GoogleClientID)
	if err != nil {
		log.Error("Error verifying Google JWT: ", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"detail": "Invalid token"})
	}

	email, emailOk := payload.Claims["email"].(string)
	if !emailOk || email == "" {
		log.Error("No email provided in token")
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"detail": "No email provided"})
	}

	emailVerified, emailVerifiedOk := payload.Claims["email_verified"].(bool)
	if !emailVerifiedOk || !emailVerified {
		log.Errorf("Error while logging in user: [%s]: Email not verified", email)
		return c.JSON(http.StatusUnauthorized, map[string]string{"detail": "Email not verified"})
	}

	givenName, _ := payload.Claims["given_name"].(string)
	familyName, _ := payload.Claims["family_name"].(string)
	if familyName == "" {
		familyName = ""
	}

	user, err := sm.UserService.LoginOrRegisterOAuth(email, givenName, familyName)
	if err != nil {
		if _, ok := err.(*services.UserNotActivatedError); ok {
			log.Errorf("Error while logging in user: [%s]: %v", email, err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"detail": "User not activated"})
		}
		log.Errorf("Error while logging in user: [%s]: %v", email, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": "Internal server error. See logs for details"})
	}

	signedToken, err := utils.GenerateAccessToken(user, cfg)
	if err != nil {
		log.Error("Error generating access token: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": "Internal server error. See logs for details"})
	}

	log.Debug("OAuth request completed - POST /auth/oauth")
	return c.JSON(http.StatusOK, dto.Token{
		AccessToken: signedToken,
		TokenType:   "Bearer",
	})
}

