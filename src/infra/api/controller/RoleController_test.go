package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/golauth/golauth/src/domain/entity"
	factoryMock "github.com/golauth/golauth/src/domain/factory/mock"
	repoMock "github.com/golauth/golauth/src/domain/repository/mock"
	"github.com/golauth/golauth/src/infra/api/controller/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
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
	s.app = fiber.New()
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

	resp, err := s.app.Test(r, -1)
	s.NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)
	var result model.RoleResponse
	_ = json.NewDecoder(resp.Body).Decode(&result)
	s.NotZero(result.ID)
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

	resp, err := s.app.Test(r, -1)
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

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusInternalServerError, resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	s.Contains(string(b), "invalid UUID length")
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
	s.roleRepo.EXPECT().Edit(r.Context(), gomock.Any()).Return(fmt.Errorf(errMessage)).Times(1)

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusInternalServerError, resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	s.Contains(string(b), errMessage)
}

func (s *RoleControllerSuite) TestChangeStatusOk() {
	roleId := uuid.New()
	changeStatus := model.RoleChangeStatus{Enabled: false}

	body, _ := json.Marshal(changeStatus)
	r, _ := http.NewRequest("PATCH", fmt.Sprintf("/roles/%s/change-status", roleId), strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	s.roleRepo.EXPECT().ExistsById(r.Context(), roleId).Return(true, nil).Times(1)
	s.roleRepo.EXPECT().ChangeStatus(r.Context(), roleId, changeStatus.Enabled).Return(nil).Times(1)

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusNoContent, resp.StatusCode)
}

func (s *RoleControllerSuite) TestChangeStatusErrParseUUID() {
	changeStatus := model.RoleChangeStatus{Enabled: false}
	body, _ := json.Marshal(changeStatus)

	r, _ := http.NewRequest("PATCH", "/roles/abc/change-status", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *RoleControllerSuite) TestChangeStatusErrSvc() {
	roleId := uuid.New()
	changeStatus := model.RoleChangeStatus{Enabled: false}
	errMessage := "could not change status for role"
	body, _ := json.Marshal(changeStatus)

	r, _ := http.NewRequest("PATCH", fmt.Sprintf("/roles/%s/change-status", roleId), strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	s.roleRepo.EXPECT().ExistsById(r.Context(), roleId).Return(true, nil).Times(1)
	s.roleRepo.EXPECT().ChangeStatus(r.Context(), roleId, changeStatus.Enabled).Return(fmt.Errorf(errMessage)).Times(1)

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusInternalServerError, resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	s.Contains(string(b), errMessage)
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

	resp, _ := s.app.Test(r, -1)

	s.Equal(http.StatusOK, resp.StatusCode)
	var result model.RoleResponse
	s.NoError(json.NewDecoder(resp.Body).Decode(&result))
	s.Equal(roleId, result.ID)
	s.Equal(roleName, result.Name)
}

func (s *RoleControllerSuite) TestFindByNameErrSvc() {
	roleName := "ROLE_NAME"
	errMessage := "could not find role by name"

	r, _ := http.NewRequest("GET", fmt.Sprintf("/roles/%s", roleName), nil)
	r.Header.Set("Content-Type", "application/json")

	s.roleRepo.EXPECT().FindByName(r.Context(), roleName).Return(nil, fmt.Errorf(errMessage)).Times(1)

	resp, _ := s.app.Test(r, -1)

	s.Equal(http.StatusInternalServerError, resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	s.Contains(string(b), errMessage)
}
