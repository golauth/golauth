package entity

// Token is the credential pair handed back by login and refresh. RefreshToken
// and ExpiresIn are empty on flows that do not mint a refresh token.
type Token struct {
	AccessToken  string
	RefreshToken string
	// ExpiresIn is the access token lifetime in seconds, for OAuth-shaped clients.
	ExpiresIn int
}
