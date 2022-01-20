//go:generate mockgen -source roleService.go -destination mock/roleService_mock.go -package mock
package usecase

import (
	"fmt"
	"github.com/golauth/golauth/api/handler/model"
	"github.com/golauth/golauth/domain/entity"
	"github.com/golauth/golauth/domain/repository"
	"github.com/google/uuid"
)

type RoleService interface {
	Create(req model.RoleRequest) (model.RoleResponse, error)
	Edit(id uuid.UUID, req model.RoleRequest) error
	ChangeStatus(id uuid.UUID, enabled bool) error
	FindByName(name string) (model.RoleResponse, error)
}

type roleService struct {
	repo repository.RoleRepository
}

func NewRoleService(r repository.RoleRepository) RoleService {
	return roleService{repo: r}
}

func (s roleService) Create(req model.RoleRequest) (model.RoleResponse, error) {
	data := entity.Role{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     true,
	}
	role, err := s.repo.Create(data)
	if err != nil {
		return model.RoleResponse{}, err
	}
	return model.RoleResponse{
		ID:           role.ID,
		Name:         role.Name,
		Description:  role.Description,
		Enabled:      role.Enabled,
		CreationDate: role.CreationDate,
	}, nil
}

func (s roleService) Edit(id uuid.UUID, req model.RoleRequest) error {
	exists, err := s.repo.ExistsById(id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("role with id %s does not exists", id)
	}
	if id != req.ID {
		return fmt.Errorf("path id[%s] and object_id[%s] does not match", id, req.ID)
	}
	data := entity.Role{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}
	return s.repo.Edit(data)
}

func (s roleService) ChangeStatus(id uuid.UUID, enabled bool) error {
	exists, err := s.repo.ExistsById(id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("role with id %s does not exists", id)
	}
	return s.repo.ChangeStatus(id, enabled)
}

func (s roleService) FindByName(name string) (model.RoleResponse, error) {
	role, err := s.repo.FindByName(name)
	if err != nil {
		return model.RoleResponse{}, err
	}
	return model.RoleResponse{
		ID:           role.ID,
		Name:         role.Name,
		Description:  role.Description,
		Enabled:      role.Enabled,
		CreationDate: role.CreationDate,
	}, nil
}
