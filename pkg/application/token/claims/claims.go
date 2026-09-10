// Package claims holds the identity a verified token carries. It lives in the
// application layer -- the layer that owns the concept -- and is its own small
// package so that the token use cases and their generated mocks can both depend
// on it without a cycle.
package claims

import "github.com/cristalhq/jwt/v3"

// Claims is the set of assertions a golauth access token carries. The JSON tags
// are the on-the-wire contract and must not change by accident.
//
// jwt.StandardClaims is still embedded for its IsValidAt expiry check; replacing
// it with plain Subject/ExpiresAt fields converted at the library edge is the
// tracked follow-up (Plan 09, step 2).
type Claims struct {
	Username    string   `json:"username"`
	FirstName   string   `json:"firstName"`
	LastName    string   `json:"lastName"`
	Authorities []string `json:"authorities,omitempty"`
	jwt.StandardClaims
}
