package model

import (
	"github.com/google/uuid"
	"time"
)

type RoleResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Enabled      bool      `json:"enabled"`
	CreationDate time.Time `json:"creationDate"`
}
