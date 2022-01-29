//go:generate mockgen -source interface.go -destination mock/interface_mock.go -package mock
package token

import (
	"github.com/golauth/golauth/domain/entity"
)

type UseCase interface {
	GenerateJwtToken(user entity.User, authorities []string) (string, error)
}
