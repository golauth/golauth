package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/golauth/golauth/domain/entity"
	mockSvc "github.com/golauth/golauth/domain/usecase/mock"
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

type UserControllerSuite struct {
	suite.Suite
	*require.Assertions
	ctrl    *gomock.Controller
	userSvc *mockSvc.MockUserService
	uc      UserController
}

func TestUserControllerSuite(t *testing.T) {
	suite.Run(t, new(UserControllerSuite))
}

func (s *UserControllerSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.ctrl = gomock.NewController(s.T())
	s.userSvc = mockSvc.NewMockUserService(s.ctrl)

	s.uc = NewUserController(s.userSvc)
}

func (s *UserControllerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *UserControllerSuite) TestFindByIDOk() {
	user := model.UserResponse{
		ID:           uuid.New(),
		Username:     "admin",
		FirstName:    "User",
		LastName:     "Name",
		Email:        "em@il.com",
		Document:     "1234",
		Enabled:      true,
		CreationDate: time.Now().AddDate(0, 0, -4),
	}
	s.userSvc.EXPECT().FindByID(user.ID).Return(user, nil).Times(1)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", fmt.Sprintf("users/%s", user.ID), nil)

	vars := map[string]string{
		"id": user.ID.String(),
	}
	r = mux.SetURLVars(r, vars)

	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	_ = jsonEncoder.Encode(user)
	s.uc.FindById(w, r)
	s.Equal(http.StatusOK, w.Code)
	s.Equal(bf, w.Body)
}

func (s *UserControllerSuite) TestAddRoleOk() {
	userId := uuid.New()
	userRole := entity.UserRole{RoleID: uuid.New(), UserID: userId, CreationDate: time.Now()}
	s.userSvc.EXPECT().AddUserRole(userRole.UserID, userRole.RoleID).Return(nil).Times(1)

	body, _ := json.Marshal(userRole)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", fmt.Sprintf("users/%s/add-role", userId), strings.NewReader(string(body)))
	vars := map[string]string{
		"id": userId.String(),
	}
	r = mux.SetURLVars(r, vars)

	s.uc.AddRole(w, r)
	s.Equal(http.StatusCreated, w.Code)
}

func (s *UserControllerSuite) TestFindByIDErrParseUUID() {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", fmt.Sprintf("users/abc"), nil)

	vars := map[string]string{
		"id": "abc",
	}
	r = mux.SetURLVars(r, vars)

	s.uc.FindById(w, r)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *UserControllerSuite) TestFindByIDErrSvc() {
	id := uuid.New()
	errMessage := "could not find user by id"
	s.userSvc.EXPECT().FindByID(id).Return(model.UserResponse{}, fmt.Errorf(errMessage)).Times(1)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", fmt.Sprintf("users/%s", id), nil)

	vars := map[string]string{
		"id": id.String(),
	}
	r = mux.SetURLVars(r, vars)

	s.uc.FindById(w, r)
	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), errMessage)
}

func (s *UserControllerSuite) TestAddRoleErrSvc() {
	userId := uuid.New()
	roleId := uuid.New()
	errMessage := "could not add role to user"
	userRole := entity.UserRole{RoleID: roleId, UserID: userId, CreationDate: time.Now()}
	body, err := json.Marshal(userRole)
	s.NoError(err)
	s.userSvc.EXPECT().AddUserRole(userId, roleId).Return(fmt.Errorf(errMessage)).Times(1)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", fmt.Sprintf("users/%s/add-role", userId), strings.NewReader(string(body)))

	vars := map[string]string{
		"id": userId.String(),
	}
	r = mux.SetURLVars(r, vars)

	s.uc.AddRole(w, r)
	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), errMessage)
}
