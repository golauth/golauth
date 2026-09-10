package token

import "time"

// Token lifetimes. These are the defaults; real values come from
// ACCESS_TOKEN_TTL and REFRESH_TOKEN_TTL, parsed once at start-up and injected
// as a Config. There is deliberately no package-level mutable expiry variable:
// one test mutating it used to leak into the next.
const (
	DefaultAccessTokenTTL  = 15 * time.Minute
	DefaultRefreshTokenTTL = 7 * 24 * time.Hour
)

// Config carries the token lifetimes into the use cases that mint tokens.
type Config struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// DefaultConfig returns the built-in lifetimes.
func DefaultConfig() Config {
	return Config{
		AccessTokenTTL:  DefaultAccessTokenTTL,
		RefreshTokenTTL: DefaultRefreshTokenTTL,
	}
}
