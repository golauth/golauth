package model

import (
	"github.com/golauth/golauth/domain/entity"
	"github.com/google/uuid"
)

type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken uuid.UUID `json:"refresh_token"`
}

func NewTokenResponseFromEntity(e *entity.Token) *TokenResponse {
	return &TokenResponse{AccessToken: e.AccessToken}
}
