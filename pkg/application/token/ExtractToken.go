package token

import "strings"

const bearerPrefix = "Bearer "

// ExtractToken returns the credentials of a RFC 6750 bearer authorization
// header. The scheme is matched case-insensitively, as required by RFC 7235;
// any other scheme is rejected instead of being sliced off blindly.
func ExtractToken(headerValue string) (string, error) {
	if len(headerValue) <= len(bearerPrefix) {
		return "", ErrBearerTokenExtract
	}
	if !strings.EqualFold(headerValue[:len(bearerPrefix)], bearerPrefix) {
		return "", ErrBearerTokenExtract
	}
	tk := strings.TrimSpace(headerValue[len(bearerPrefix):])
	if tk == "" {
		return "", ErrBearerTokenExtract
	}
	return tk, nil
}
