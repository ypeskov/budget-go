package dto

type OAuthToken struct {
	Credential string `json:"credential"`
}

type GoogleClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type Token struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
}

