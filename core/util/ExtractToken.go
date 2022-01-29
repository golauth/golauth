package util

import (
	"net/http"
)

func ExtractToken(r *http.Request) (string, error) {
	authorization := r.Header.Get("Authorization")
	if len(authorization) > len("Bearer ") {
		return authorization[7:], nil
	}
	return "", ErrBearerTokenExtract
}
