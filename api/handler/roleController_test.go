package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golauth/entity"
	"golauth/model"
	"golauth/usecase/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type RoleControllerSuite struct {
	suite.Suite
	*require.Assertions
	ctrl    *gomock.Controller
	roleSvc *mock.MockRoleService

	rc RoleController
}

func TestRoleControllerSuite(t *testing.T) {
	suite.Run(t, new(RoleControllerSuite))
}

func (s *RoleControllerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.roleSvc = mock.NewMockRoleService(s.ctrl)

	s.rc = NewRoleController(s.roleSvc)
}

func (s *RoleControllerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *RoleControllerSuite) TestCreateOk() {
	role := model.RoleRequest{
		Name:        "Role",
		Description: "Description",
	}
	savedRole := model.RoleResponse{
		ID:           uuid.New(),
		Name:         "Role",
		Description:  "Description",
		Enabled:      true,
		CreationDate: time.Now(),
	}
	s.roleSvc.EXPECT().Create(role).Return(savedRole, nil).Times(1)

	body, _ := json.Marshal(role)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/roles", strings.NewReader(string(body)))

	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	_ = jsonEncoder.Encode(savedRole)

	s.rc.Create(w, r)
	s.Equal(http.StatusCreated, w.Code)
	s.Equal(bf, w.Body)
}

func (s RoleControllerSuite) TestEditRoleOk() {
	role := model.RoleRequest{
		ID:          uuid.New(),
		Name:        "Role Edited",
		Description: "Description Edited",
	}
	s.roleSvc.EXPECT().Edit(role.ID, role).Return(nil).Times(1)

	body, _ := json.Marshal(role)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PUT", fmt.Sprintf("/roles/%s", role.ID), strings.NewReader(string(body)))
	vars := map[string]string{
		"id": role.ID.String(),
	}
	r = mux.SetURLVars(r, vars)

	s.rc.Edit(w, r)
	s.Equal(http.StatusOK, w.Code)

	var result entity.Role
	_ = json.NewDecoder(w.Body).Decode(&result)
	s.Equal(role.ID, result.ID)
	s.Equal("Role Edited", result.Name)
	s.Equal("Description Edited", result.Description)
}

func (s RoleControllerSuite) TestEditRoleNotOk() {

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
