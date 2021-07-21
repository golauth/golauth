//go:generate mockgen -source interface.go -destination mock/interface_mock.go -package mock
package token

import (
	"golauth/entity"
	"net/http"
)

type UseCase interface {
	ValidateToken(token string) error
	ExtractToken(r *http.Request) (string, error)
	GenerateJwtToken(user entity.User, authorities []string) (string, error)
}
