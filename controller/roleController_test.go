package controller

import (
	"bytes"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golauth/model"
	"golauth/repository/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type RoleControllerSuite struct {
	suite.Suite
	*require.Assertions
	ctrl  *gomock.Controller
	rRepo *mock.MockRoleRepository

	rc RoleController
}

func TestRoleControllerSuite(t *testing.T) {
	suite.Run(t, new(RoleControllerSuite))
}

func (s *RoleControllerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.rRepo = mock.NewMockRoleRepository(s.ctrl)

	s.rc = NewRoleController(s.rRepo)
}

func (s *RoleControllerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *RoleControllerSuite) TestCreateOk() {
	role := model.Role{
		Name:        "Role",
		Description: "Description",
		Enabled:     true,
	}
	savedRole := model.Role{
		ID:           1,
		Name:         "Role",
		Description:  "Description",
		Enabled:      true,
		CreationDate: time.Now(),
	}
	s.rRepo.EXPECT().Create(role).Return(savedRole, nil).Times(1)

	body, _ := json.Marshal(role)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/roles", strings.NewReader(string(body)))

	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	_ = jsonEncoder.Encode(savedRole)

	s.rc.CreateRole(w, r)
	s.Equal(http.StatusCreated, w.Code)
	s.Equal(bf, w.Body)
}

func (s RoleControllerSuite) TestEditRoleOk() {
	role := model.Role{
		ID:          1,
		Name:        "Role Edited",
		Description: "Description Edited",
		Enabled:     false,
	}
	s.rRepo.EXPECT().Edit(role).Return(nil).Times(1)

	body, _ := json.Marshal(role)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PUT", "/roles/1", strings.NewReader(string(body)))
	vars := map[string]string{
		"id": "1",
	}
	r = mux.SetURLVars(r, vars)

	s.rc.EditRole(w, r)
	s.Equal(http.StatusOK, w.Code)

	var result model.Role
	_ = json.NewDecoder(w.Body).Decode(&result)
	s.Equal(1, result.ID)
	s.Equal("Role Edited", result.Name)
	s.Equal("Description Edited", result.Description)
}

func (s RoleControllerSuite) TestEditRoleNotOk() {

	role := model.Role{
		ID:          1,
		Name:        "Role Edited",
		Description: "Description Edited",
		Enabled:     false,
	}

	body, _ := json.Marshal(role)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PUT", "/roles/abc", strings.NewReader(string(body)))
	vars := map[string]string{
		"id": "abc",
	}
	r = mux.SetURLVars(r, vars)

	s.rc.EditRole(w, r)
	s.Equal(http.StatusInternalServerError, w.Code)

	var result model.Error
	_ = json.NewDecoder(w.Body).Decode(&result)

	s.Equal(http.StatusInternalServerError, result.StatusCode)
	s.Equal("cannot cast [id] to [int]", result.Message)
}
