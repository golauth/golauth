// Package apperr holds the small set of sentinel errors that cross layers.
// A use case or repository wraps one of these with %w to say what kind of
// failure happened; the HTTP error handler is the only place that turns the
// kind into a status code and a client-visible message. Nothing else in a
// wrapped chain -- driver text, query fragments, ids -- ever reaches a client.
package apperr

import "errors"

var (
	// ErrNotFound: the addressed resource does not exist. -> 404
	ErrNotFound = errors.New("resource not found")
	// ErrAlreadyExists: the write collides with a uniqueness constraint. -> 409
	ErrAlreadyExists = errors.New("resource already exists")
	// ErrInvalidInput: the request is syntactically or semantically malformed
	// (bad uuid, unparseable body, mismatched ids). -> 400
	ErrInvalidInput = errors.New("invalid input")
	// ErrUnauthorized: authentication is missing or invalid. -> 401
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden: authenticated but not allowed. -> 403
	ErrForbidden = errors.New("forbidden")
)
