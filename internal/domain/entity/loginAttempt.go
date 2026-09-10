package entity

import (
	"time"

	"github.com/google/uuid"
)

// LoginAttempt is the per-account failure counter that backs the token
// endpoint's lockout. IP rate limiting does not cover a distributed guessing
// attack against a single account; this does.
type LoginAttempt struct {
	UserID      uuid.UUID
	FailedCount int
	LastFailure time.Time
	LockedUntil time.Time
}

// Locked reports whether the account is under an active lock at the given time.
func (a LoginAttempt) Locked(now time.Time) bool {
	return a.LockedUntil.After(now)
}
