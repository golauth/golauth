package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/golauth/golauth/domain/entity"
	factoryMock "github.com/golauth/golauth/domain/factory/mock"
	repoMock "github.com/golauth/golauth/domain/repository/mock"
	"github.com/golauth/golauth/domain/usecase/mock"
	"github.com/golauth/golauth/infra/api/controller/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type RoleControllerSuite struct {
	suite.Suite
	*require.Assertions
	ctrl        *gomock.Controller
	roleSvc     *mock.MockRoleService
	repoFactory *factoryMock.MockRepositoryFactory
	roleRepo    *repoMock.MockRoleRepository
	rc          RoleController
}

func TestRoleControllerSuite(t *testing.T) {
	suite.Run(t, new(RoleControllerSuite))
}

func (s *RoleControllerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.roleSvc = mock.NewMockRoleService(s.ctrl)
	s.roleRepo = repoMock.NewMockRoleRepository(s.ctrl)
	s.repoFactory = factoryMock.NewMockRepositoryFactory(s.ctrl)
	s.repoFactory.EXPECT().NewRoleRepository().AnyTimes().Return(s.roleRepo)

	s.rc = NewRoleController(s.roleSvc, s.repoFactory)
}

func (s *RoleControllerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s RoleControllerSuite) TestCreateRoleOk() {
	input := model.RoleRequest{Name: "New Role", Description: "New Role Description"}
	body, _ := json.Marshal(input)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/roles", strings.NewReader(string(body)))
	s.roleRepo.EXPECT().Create(r.Context(), gomock.Any()).Return(&entity.Role{ID: uuid.New(), Name: input.Name, Description: input.Description}, nil)

	s.rc.Create(w, r)
	s.Equal(http.StatusCreated, w.Code)
	var result model.RoleResponse
	_ = json.NewDecoder(w.Body).Decode(&result)
	s.NotZero(result.ID)
}

func (s RoleControllerSuite) TestEditRoleOk() {
	role := model.RoleRequest{
		ID:          uuid.New(),
		Name:        "Role Edited",
		Description: "Description Edited",
	}

	body, _ := json.Marshal(role)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PUT", fmt.Sprintf("/roles/%s", role.ID), strings.NewReader(string(body)))
	vars := map[string]string{
		"id": role.ID.String(),
	}
	r = mux.SetURLVars(r, vars)
	s.roleRepo.EXPECT().ExistsById(r.Context(), role.ID).Return(true, nil).Times(1)
	s.roleRepo.EXPECT().Edit(r.Context(), gomock.Any()).Return(nil).Times(1)

	s.rc.Edit(w, r)
	s.Equal(http.StatusOK, w.Code)

	var result entity.Role
	_ = json.NewDecoder(w.Body).Decode(&result)
	s.Equal(role.ID, result.ID)
	s.Equal("Role Edited", result.Name)
	s.Equal("Description Edited", result.Description)
}

func (s RoleControllerSuite) TestEditRoleErrParseUUID() {

	role := model.RoleRequest{
		ID:          uuid.New(),
		Name:        "Role Edited",
		Description: "Description Edited",
	}

	body, _ := json.Marshal(role)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PUT", "/roles/abc", strings.NewReader(string(body)))
	vars := map[string]string{
		"id": "abc",
	}
	r = mux.SetURLVars(r, vars)

	s.rc.Edit(w, r)
	s.Equal(http.StatusInternalServerError, w.Code)
	expected := errors.New("cannot cast [id] to [uuid]")
	s.ErrorAs(errors.New(w.Body.String()), &expected)
}

func (s RoleControllerSuite) TestEditRoleNotOk() {
	roleId := uuid.New()
	role := model.RoleRequest{
		ID:          roleId,
		Name:        "Role Edited",
		Description: "Description Edited",
	}
	errMessage := "could not edit role"

	body, _ := json.Marshal(role)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PUT", fmt.Sprintf("/roles/%s", roleId), strings.NewReader(string(body)))
	vars := map[string]string{
		"id": roleId.String(),
	}
	r = mux.SetURLVars(r, vars)
	s.roleRepo.EXPECT().ExistsById(r.Context(), roleId).Return(true, nil).Times(1)
	s.roleRepo.EXPECT().Edit(r.Context(), gomock.Any()).Return(fmt.Errorf(errMessage)).Times(1)

	s.rc.Edit(w, r)
	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), errMessage)
}

func (s RoleControllerSuite) TestChangeStatusOk() {
	roleId := uuid.New()
	changeStatus := model.RoleChangeStatus{Enabled: false}

	body, _ := json.Marshal(changeStatus)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PATCH", fmt.Sprintf("/roles/%s/change-status", roleId), strings.NewReader(string(body)))
	vars := map[string]string{
		"id": roleId.String(),
	}
	r = mux.SetURLVars(r, vars)
	s.roleRepo.EXPECT().ExistsById(r.Context(), roleId).Return(true, nil).Times(1)
	s.roleRepo.EXPECT().ChangeStatus(r.Context(), roleId, changeStatus.Enabled).Return(nil).Times(1)

	s.rc.ChangeStatus(w, r)
	s.Equal(http.StatusOK, w.Code)
}

func (s RoleControllerSuite) TestChangeStatusErrParseUUID() {
	changeStatus := model.RoleChangeStatus{Enabled: false}
	body, _ := json.Marshal(changeStatus)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PATCH", "/roles/abc/change-status", strings.NewReader(string(body)))
	vars := map[string]string{
		"id": "abc",
	}
	r = mux.SetURLVars(r, vars)

	s.rc.ChangeStatus(w, r)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s RoleControllerSuite) TestChangeStatusErrSvc() {
	roleId := uuid.New()
	changeStatus := model.RoleChangeStatus{Enabled: false}
	errMessage := "could not change status for role"

	body, _ := json.Marshal(changeStatus)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PATCH", fmt.Sprintf("/roles/%s/change-status", roleId), strings.NewReader(string(body)))
	vars := map[string]string{
		"id": roleId.String(),
	}
	r = mux.SetURLVars(r, vars)
	s.roleRepo.EXPECT().ExistsById(r.Context(), roleId).Return(true, nil).Times(1)
	s.roleRepo.EXPECT().ChangeStatus(r.Context(), roleId, changeStatus.Enabled).Return(fmt.Errorf(errMessage)).Times(1)

	s.rc.ChangeStatus(w, r)
	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), errMessage)
}

func (s RoleControllerSuite) TestFindByNameOk() {
	roleId := uuid.New()
	roleName := "ROLE_NAME"
	resp := model.RoleResponse{
		ID:           roleId,
		Name:         roleName,
		Description:  "Role description",
		Enabled:      true,
		CreationDate: time.Now(),
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", fmt.Sprintf("/roles/%s", roleName), nil)
	vars := map[string]string{
		"name": roleName,
	}
	r = mux.SetURLVars(r, vars)
	s.roleSvc.EXPECT().FindByName(r.Context(), roleName).Return(resp, nil).Times(1)

	s.rc.FindByName(w, r)

	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	_ = jsonEncoder.Encode(resp)

	s.Equal(http.StatusOK, w.Code)
	s.Equal(bf, w.Body)
}

func (s RoleControllerSuite) TestFindByNameErrSvc() {
	roleName := "ROLE_NAME"
	errMessage := "could not find role by name"

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", fmt.Sprintf("/roles/%s", roleName), nil)
	vars := map[string]string{
		"name": roleName,
	}
	r = mux.SetURLVars(r, vars)
	s.roleSvc.EXPECT().FindByName(r.Context(), roleName).Return(model.RoleResponse{}, fmt.Errorf(errMessage)).Times(1)

	s.rc.FindByName(w, r)

	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), errMessage)
}
