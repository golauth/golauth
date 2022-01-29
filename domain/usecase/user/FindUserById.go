//go:generate mockgen -source FindUserById.go -destination mock/FindUserById_mock.go -package mock
package user

import (
	"context"
	"github.com/golauth/golauth/domain/repository"
	"github.com/golauth/golauth/infra/api/controller/model"
	"github.com/google/uuid"
)

type FindUserById interface {
	Execute(ctx context.Context, id uuid.UUID) (*model.UserResponse, error)
}

func NewFindUserById(repo repository.UserRepository) FindUserById {
	return findUserById{repo: repo}
}

type findUserById struct {
	repo repository.UserRepository
}

func (uc findUserById) Execute(ctx context.Context, id uuid.UUID) (*model.UserResponse, error) {
	user, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &model.UserResponse{
		ID:           user.ID,
		Username:     user.Username,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		Document:     user.Document,
		Enabled:      user.Enabled,
		CreationDate: user.CreationDate,
	}, nil
}
