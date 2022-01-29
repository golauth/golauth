package entity

import (
	"github.com/google/uuid"
	"time"
)

type Role struct {
	ID           uuid.UUID
	Name         string
	Description  string
	Enabled      bool
	CreationDate time.Time
}

func NewRole(name string, description string) *Role {
	return &Role{
		Name:        name,
		Description: description,
		Enabled:     true,
	}
}
