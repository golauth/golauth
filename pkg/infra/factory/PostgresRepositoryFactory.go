package factory

import (
	"github.com/golauth/golauth/pkg/domain/factory"
	"github.com/golauth/golauth/pkg/domain/repository"
	"github.com/golauth/golauth/pkg/infra/database"
	"github.com/golauth/golauth/pkg/infra/repository/postgres"
)

type PostgresRepositoryFactory struct {
	db database.Database
}

func NewPostgresRepositoryFactory(db database.Database) factory.RepositoryFactory {
	return PostgresRepositoryFactory{db: db}
}

func (p PostgresRepositoryFactory) NewRoleRepository() repository.RoleRepository {
	return postgres.NewRoleRepository(p.db)
}

func (p PostgresRepositoryFactory) NewUserAuthorityRepository() repository.UserAuthorityRepository {
	return postgres.NewUserAuthorityRepository(p.db)
}

func (p PostgresRepositoryFactory) NewUserRepository() repository.UserRepository {
	return postgres.NewUserRepository(p.db)
}

func (p PostgresRepositoryFactory) NewUserRoleRepository() repository.UserRoleRepository {
	return postgres.NewUserRoleRepository(p.db)
}
