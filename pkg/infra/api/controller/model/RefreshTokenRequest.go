package model

// RefreshTokenRequest is the body of POST /auth/token/refresh and
// POST /auth/logout: the opaque refresh-token string the client was given at
// login.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}
