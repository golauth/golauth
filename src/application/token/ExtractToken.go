package token

func ExtractToken(headerValue string) (string, error) {
	if len(headerValue) > len("Bearer ") {
		return headerValue[7:], nil
	}
	return "", ErrBearerTokenExtract
}
