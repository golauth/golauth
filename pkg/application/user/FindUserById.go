//go:generate mockgen -source FindUserById.go -destination mock/FindUserById_mock.go -package mock
package user

import (
	"context"
	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/golauth/golauth/pkg/domain/repository"
	"github.com/google/uuid"
)

type FindUserById interface {
	Execute(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

func NewFindUserById(repo repository.UserRepository) FindUserById {
	return findUserById{repo: repo}
}

type findUserById struct {
	repo repository.UserRepository
}

func (uc findUserById) Execute(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}
