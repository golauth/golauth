package model

import (
	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/google/uuid"
	"time"
)

type UserResponse struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Email        string    `json:"email"`
	Document     string    `json:"document"`
	Enabled      bool      `json:"enabled"`
	CreationDate time.Time `json:"creationDate"`
}

func NewUserResponseFromEntity(e *entity.User) *UserResponse {
	return &UserResponse{
		ID:           e.ID,
		Username:     e.Username,
		FirstName:    e.FirstName,
		LastName:     e.LastName,
		Email:        e.Email,
		Document:     e.Document,
		Enabled:      e.Enabled,
		CreationDate: e.CreationDate,
	}
}
