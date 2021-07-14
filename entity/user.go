package entity

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID           uuid.UUID
	Username     string
	FirstName    string
	LastName     string
	Email        string
	Document     string
	Password     string
	Enabled      bool
	CreationDate time.Time
}
