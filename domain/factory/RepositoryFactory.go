//go:generate mockgen -source RepositoryFactory.go -destination mock/RepositoryFactory_mock.go -package mock
package factory

import (
	"github.com/golauth/golauth/domain/repository"
)

type RepositoryFactory interface {
	NewRoleRepository() repository.RoleRepository
	NewUserAuthorityRepository() repository.UserAuthorityRepository
	NewUserRepository() repository.UserRepository
	NewUserRoleRepository() repository.UserRoleRepository
}
