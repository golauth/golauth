package repository

import (
	"fmt"

	"github.com/golauth/golauth/pkg/domain/apperr"
)

// ErrUserAlreadyExists is returned by UserRepository.Create when the write
// collides with a unique index (username or email). It wraps apperr.ErrAlreadyExists,
// so both errors.Is(err, ErrUserAlreadyExists) and errors.Is(err, apperr.ErrAlreadyExists)
// match, and the central error handler maps it to HTTP 409.
var ErrUserAlreadyExists = fmt.Errorf("user already exists: %w", apperr.ErrAlreadyExists)
