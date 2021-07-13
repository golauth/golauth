package model

import (
	"github.com/google/uuid"
	"time"
)

type RoleRequest struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type RoleResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Enabled      bool      `json:"enabled"`
	CreationDate time.Time `json:"creationDate"`
}

type RoleChangeStatus struct {
	Enabled bool `json:"enabled"`
}
