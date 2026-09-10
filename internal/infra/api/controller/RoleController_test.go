package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/internal/domain/apperr"
	"github.com/golauth/golauth/internal/domain/entity"
	factoryMock "github.com/golauth/golauth/internal/domain/factory/mock"
	repoMock "github.com/golauth/golauth/internal/domain/repository/mock"
	"github.com/golauth/golauth/internal/infra/api/controller/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type RoleControllerSuite struct {
	suite.Suite
	*require.Assertions
	ctrl        *gomock.Controller
	repoFactory *factoryMock.MockRepositoryFactory
	roleRepo    *repoMock.MockRoleRepository
	app         *fiber.App
	rc          RoleController
}

func TestRoleControllerSuite(t *testing.T) {
	suite.Run(t, new(RoleControllerSuite))
}

func (s *RoleControllerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.roleRepo = repoMock.NewMockRoleRepository(s.ctrl)
	s.repoFactory = factoryMock.NewMockRepositoryFactory(s.ctrl)
	s.repoFactory.EXPECT().NewRoleRepository().AnyTimes().Return(s.roleRepo)

	s.rc = NewRoleController(s.repoFactory)
	s.app = newErrApp()
	s.app.Post("/roles", s.rc.Create)
	s.app.Put("/roles/:id", s.rc.Edit)
	s.app.Patch("/roles/:id/change-status", s.rc.ChangeStatus)
	s.app.Get("/roles/:name", s.rc.FindByName)
}

func (s *RoleControllerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *RoleControllerSuite) TestCreateRoleOk() {
	input := model.RoleRequest{Name: "New Role", Description: "New Role Description"}
	body, _ := json.Marshal(input)
	r, _ := http.NewRequest("POST", "/roles", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	s.roleRepo.EXPECT().Create(r.Context(), gomock.Any()).Return(&entity.Role{ID: uuid.New(), Name: input.Name, Description: input.Description}, nil)

	resp, err := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)
	var result model.RoleResponse
	_ = json.NewDecoder(resp.Body).Decode(&result)
	s.NotZero(result.ID)
}

// An unparseable body is 400 invalid_input, not the old 500.
func (s *RoleControllerSuite) TestCreateRoleBadBody() {
	r, _ := http.NewRequest("POST", "/roles", strings.NewReader("{not json"))
	r.Header.Set("Content-Type", "application/json")

	resp, _ := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Equal(http.StatusBadRequest, resp.StatusCode)
	c, _ := readContract(resp)
	s.Equal("invalid_input", c.Error.Code)
}

func (s *RoleControllerSuite) TestEditRoleOk() {
	role := model.RoleRequest{
		ID:          uuid.New(),
		Name:        "Role Edited",
		Description: "Description Edited",
	}

	body, _ := json.Marshal(role)
	r, _ := http.NewRequest("PUT", fmt.Sprintf("/roles/%s", role.ID), strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	s.roleRepo.EXPECT().ExistsById(r.Context(), role.ID).Return(true, nil).Times(1)
	s.roleRepo.EXPECT().Edit(r.Context(), gomock.Any()).Return(nil).Times(1)

	resp, err := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	var result entity.Role
	_ = json.NewDecoder(resp.Body).Decode(&result)
	s.Equal(role.ID, result.ID)
	s.Equal("Role Edited", result.Name)
	s.Equal("Description Edited", result.Description)
}

func (s *RoleControllerSuite) TestEditRoleErrParseUUID() {
	role := model.RoleRequest{
		ID:          uuid.New(),
		Name:        "Role Edited",
		Description: "Description Edited",
	}

	body, _ := json.Marshal(role)
	r, _ := http.NewRequest("PUT", "/roles/abc", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	resp, _ := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Equal(http.StatusBadRequest, resp.StatusCode)
	c, raw := readContract(resp)
	s.Equal("invalid_input", c.Error.Code)
	s.NotContains(raw, "invalid UUID length")
}

func (s *RoleControllerSuite) TestEditRoleNotOk() {
	roleId := uuid.New()
	role := model.RoleRequest{
		ID:          roleId,
		Name:        "Role Edited",
		Description: "Description Edited",
	}
	errMessage := "could not edit role"

	body, _ := json.Marshal(role)
	r, _ := http.NewRequest("PUT", fmt.Sprintf("/roles/%s", roleId), strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	s.roleRepo.EXPECT().ExistsById(r.Context(), roleId).Return(true, nil).Times(1)
	s.roleRepo.EXPECT().Edit(r.Context(), gomock.Any()).Return(errors.New(errMessage)).Times(1)

	resp, _ := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Equal(http.StatusInternalServerError, resp.StatusCode)
	c, raw := readContract(resp)
	s.Equal("internal_error", c.Error.Code)
	s.NotContains(raw, errMessage)
}

// A missing role (apperr.ErrNotFound from the use case) is 404, not 500.
func (s *RoleControllerSuite) TestEditRoleNotFound() {
	roleId := uuid.New()
	role := model.RoleRequest{ID: roleId, Name: "X", Description: "Y"}
	body, _ := json.Marshal(role)
	r, _ := http.NewRequest("PUT", fmt.Sprintf("/roles/%s", roleId), strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	s.roleRepo.EXPECT().ExistsById(r.Context(), roleId).Return(false, nil).Times(1)

	resp, _ := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Equal(http.StatusNotFound, resp.StatusCode)
	c, _ := readContract(resp)
	s.Equal("not_found", c.Error.Code)
}

func (s *RoleControllerSuite) TestChangeStatusOk() {
	roleId := uuid.New()
	changeStatus := model.RoleChangeStatus{Enabled: false}

	body, _ := json.Marshal(changeStatus)
	r, _ := http.NewRequest("PATCH", fmt.Sprintf("/roles/%s/change-status", roleId), strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	s.roleRepo.EXPECT().ExistsById(r.Context(), roleId).Return(true, nil).Times(1)
	s.roleRepo.EXPECT().ChangeStatus(r.Context(), roleId, changeStatus.Enabled).Return(nil).Times(1)

	resp, _ := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Equal(http.StatusNoContent, resp.StatusCode)
}

func (s *RoleControllerSuite) TestChangeStatusErrParseUUID() {
	changeStatus := model.RoleChangeStatus{Enabled: false}
	body, _ := json.Marshal(changeStatus)

	r, _ := http.NewRequest("PATCH", "/roles/abc/change-status", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	resp, _ := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Equal(http.StatusBadRequest, resp.StatusCode)
	c, _ := readContract(resp)
	s.Equal("invalid_input", c.Error.Code)
}

func (s *RoleControllerSuite) TestChangeStatusErrSvc() {
	roleId := uuid.New()
	changeStatus := model.RoleChangeStatus{Enabled: false}
	errMessage := "could not change status for role"
	body, _ := json.Marshal(changeStatus)

	r, _ := http.NewRequest("PATCH", fmt.Sprintf("/roles/%s/change-status", roleId), strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	s.roleRepo.EXPECT().ExistsById(r.Context(), roleId).Return(true, nil).Times(1)
	s.roleRepo.EXPECT().ChangeStatus(r.Context(), roleId, changeStatus.Enabled).Return(errors.New(errMessage)).Times(1)

	resp, _ := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Equal(http.StatusInternalServerError, resp.StatusCode)
	c, raw := readContract(resp)
	s.Equal("internal_error", c.Error.Code)
	s.NotContains(raw, errMessage)
}

func (s *RoleControllerSuite) TestFindByNameOk() {
	roleId := uuid.New()
	roleName := "ROLE_NAME"
	roleEntity := &entity.Role{
		ID:           roleId,
		Name:         roleName,
		Description:  "Role description",
		Enabled:      true,
		CreationDate: time.Now(),
	}

	r, _ := http.NewRequest("GET", fmt.Sprintf("/roles/%s", roleName), nil)
	r.Header.Set("Content-Type", "application/json")

	s.roleRepo.EXPECT().FindByName(r.Context(), roleName).Return(roleEntity, nil).Times(1)

	resp, _ := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})

	s.Equal(http.StatusOK, resp.StatusCode)
	var result model.RoleResponse
	s.NoError(json.NewDecoder(resp.Body).Decode(&result))
	s.Equal(roleId, result.ID)
	s.Equal(roleName, result.Name)
}

func (s *RoleControllerSuite) TestFindByNameErrSvc() {
	roleName := "ROLE_NAME"
	errMessage := "could not find role by name: pq: timeout"

	r, _ := http.NewRequest("GET", fmt.Sprintf("/roles/%s", roleName), nil)
	r.Header.Set("Content-Type", "application/json")

	s.roleRepo.EXPECT().FindByName(r.Context(), roleName).Return(nil, errors.New(errMessage)).Times(1)

	resp, _ := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})

	s.Equal(http.StatusInternalServerError, resp.StatusCode)
	c, raw := readContract(resp)
	s.Equal("internal_error", c.Error.Code)
	s.NotContains(raw, "pq:")
	s.NotContains(raw, errMessage)
}

// A role the repository does not have (apperr.ErrNotFound) is 404.
func (s *RoleControllerSuite) TestFindByNameNotFound() {
	roleName := "GHOST"
	r, _ := http.NewRequest("GET", fmt.Sprintf("/roles/%s", roleName), nil)
	r.Header.Set("Content-Type", "application/json")

	s.roleRepo.EXPECT().FindByName(r.Context(), roleName).
		Return(nil, fmt.Errorf("role %q: %w", roleName, apperr.ErrNotFound)).Times(1)

	resp, _ := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Equal(http.StatusNotFound, resp.StatusCode)
	c, _ := readContract(resp)
	s.Equal("not_found", c.Error.Code)
}
