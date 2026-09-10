package repository

import "errors"

// ErrUserAlreadyExists is returned by UserRepository.Create when the write
// collides with a unique index (username or email). It is a client error --
// the controller maps it to HTTP 409 -- not an internal failure. Repositories
// wrap it with %w so callers can match it with errors.Is.
var ErrUserAlreadyExists = errors.New("user already exists")
