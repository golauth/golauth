package handler

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

type UserControllerSuite struct {
	suite.Suite
	*require.Assertions
	ctrl   *gomock.Controller
	uRepo  *mock.MockUserRepository
	urRepo *mock.MockUserRoleRepository
	uc     UserController
}

func TestUserControllerSuite(t *testing.T) {
	suite.Run(t, new(UserControllerSuite))
}

func (s *UserControllerSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.ctrl = gomock.NewController(s.T())
	s.uRepo = mock.NewMockUserRepository(s.ctrl)
	s.urRepo = mock.NewMockUserRoleRepository(s.ctrl)

	s.uc = NewUserController(s.uRepo, s.urRepo)
}

func (s *UserControllerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *UserControllerSuite) TestFindByUsernameOk() {
	user := model.User{
		ID:           0,
		Username:     "admin",
		FirstName:    "User",
		LastName:     "Name",
		Email:        "em@il.com",
		Document:     "1234",
		Enabled:      true,
		CreationDate: time.Now(),
	}
	s.uRepo.EXPECT().FindByUsername("admin").Return(user, nil).Times(1)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "users/admin", nil)

	vars := map[string]string{
		"username": "admin",
	}
	r = mux.SetURLVars(r, vars)

	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	_ = jsonEncoder.Encode(user)
	s.uc.FindByUsername(w, r)
	s.Equal(http.StatusOK, w.Code)
	s.Equal(bf, w.Body)
}

func (s *UserControllerSuite) TestAddRoleOk() {
	userRole := model.UserRole{RoleID: 2, UserID: 2, CreationDate: time.Now()}
	s.urRepo.EXPECT().AddUserRole(userRole.UserID, userRole.RoleID).Return(userRole, nil).Times(1)

	body, _ := json.Marshal(userRole)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "users/admin/add-role", strings.NewReader(string(body)))
	vars := map[string]string{
		"username": "admin",
	}
	r = mux.SetURLVars(r, vars)

	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	_ = jsonEncoder.Encode(userRole)
	s.uc.AddRole(w, r)
	s.Equal(http.StatusCreated, w.Code)
	s.Equal(bf, w.Body)
}
