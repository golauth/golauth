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

// IsActive reports whether the account is allowed to authenticate. It is the
// one predicate the login and refresh flows check; naming it keeps the reason
// for a rejection legible at the call site. The field stays exported for the
// repository row mapping.
func (u *User) IsActive() bool {
	return u.Enabled
}
