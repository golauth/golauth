package model

import (
	"github.com/cristalhq/jwt/v3"
	"github.com/google/uuid"
)

type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken uuid.UUID `json:"refresh_token"`
}

type Claims struct {
	Username    string   `json:"username"`
	FirstName   string   `json:"firstName"`
	LastName    string   `json:"lastName"`
	Authorities []string `json:"authorities,omitempty"`
	jwt.StandardClaims
}
