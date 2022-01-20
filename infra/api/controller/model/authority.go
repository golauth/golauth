package model

import (
	"github.com/google/uuid"
	"time"
)

type AuthorityRequest struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type AuthorityResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Enabled      bool      `json:"enabled"`
	CreationDate time.Time `json:"creationDate"`
}
