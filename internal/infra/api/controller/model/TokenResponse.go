package model

import (
	"github.com/golauth/golauth/internal/domain/entity"
)

// TokenResponse is the OAuth-shaped body returned by login and refresh.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func NewTokenResponseFromEntity(e *entity.Token) *TokenResponse {
	return &TokenResponse{
		AccessToken:  e.AccessToken,
		RefreshToken: e.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    e.ExpiresIn,
	}
}
