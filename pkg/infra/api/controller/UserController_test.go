package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golauth/golauth/pkg/application/user/mock"
	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/golauth/golauth/pkg/infra/api/controller/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"strings"
	"testing"
	time "time"
)

type UserControllerSuite struct {
	suite.Suite
	*require.Assertions
	ctrl         *gomock.Controller
	findUserById *mock.MockFindUserById
	addUserRole  *mock.MockAddUserRole
	uc           UserController
	app          *fiber.App
}

func TestUserControllerSuite(t *testing.T) {
	suite.Run(t, new(UserControllerSuite))
}

func (s *UserControllerSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.ctrl = gomock.NewController(s.T())
	s.findUserById = mock.NewMockFindUserById(s.ctrl)
	s.addUserRole = mock.NewMockAddUserRole(s.ctrl)

	s.uc = NewUserController(s.findUserById, s.addUserRole)
	s.app = fiber.New()
	s.app.Get("/users/:id", s.uc.FindById)
	s.app.Post("/users/:id/add-role", s.uc.AddRole)
}

func (s *UserControllerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *UserControllerSuite) TestFindByIDOk() {
	user := &entity.User{
		ID:           uuid.New(),
		Username:     "admin",
		FirstName:    "User",
		LastName:     "Name",
		Email:        "em@il.com",
		Document:     "1234",
		Enabled:      true,
		CreationDate: time.Now().AddDate(0, 0, -4),
	}

	r, _ := http.NewRequest("GET", fmt.Sprintf("/users/%s", user.ID), nil)
	r.Header.Set("Content-Type", "application/json")

	s.findUserById.EXPECT().Execute(r.Context(), user.ID).Return(user, nil).Times(1)

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusOK, resp.StatusCode)
	var userResponse model.UserResponse
	s.NoError(json.NewDecoder(resp.Body).Decode(&userResponse))
	s.Equal(user.ID, userResponse.ID)
	s.Equal(user.Username, userResponse.Username)
	s.Equal(user.FirstName, userResponse.FirstName)
	s.Equal(user.LastName, userResponse.LastName)
	s.Equal(user.Email, userResponse.Email)
	s.Equal(user.Document, userResponse.Document)
	s.True(userResponse.Enabled)
}

func (s *UserControllerSuite) TestAddRoleOk() {
	userId := uuid.New()
	userRole := entity.UserRole{RoleID: uuid.New(), UserID: userId, CreationDate: time.Now()}
	body, _ := json.Marshal(userRole)

	r, _ := http.NewRequest("POST", fmt.Sprintf("/users/%s/add-role", userId), strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	s.addUserRole.EXPECT().Execute(r.Context(), userRole.UserID, userRole.RoleID).Return(nil).Times(1)

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusCreated, resp.StatusCode)
}

func (s *UserControllerSuite) TestFindByIDErrParseUUID() {
	r, _ := http.NewRequest("GET", fmt.Sprintf("/users/abc"), nil)
	r.Header.Set("Content-Type", "application/json")

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *UserControllerSuite) TestFindByIDErrSvc() {
	id := uuid.New()
	errMessage := "could not find user by id"

	r, _ := http.NewRequest("GET", fmt.Sprintf("/users/%s", id), nil)
	r.Header.Set("Content-Type", "application/json")

	s.findUserById.EXPECT().Execute(r.Context(), id).Return(nil, errors.New(errMessage)).Times(1)

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusInternalServerError, resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	s.Contains(string(b), errMessage)
}

func (s *UserControllerSuite) TestAddRoleErrSvc() {
	userId := uuid.New()
	roleId := uuid.New()
	errMessage := "could not add role to user"
	userRole := entity.UserRole{RoleID: roleId, UserID: userId, CreationDate: time.Now()}
	body, err := json.Marshal(userRole)
	s.NoError(err)

	r, _ := http.NewRequest("POST", fmt.Sprintf("/users/%s/add-role", userId), strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	s.addUserRole.EXPECT().Execute(r.Context(), userId, roleId).Return(errors.New(errMessage)).Times(1)

	resp, err := s.app.Test(r, -1)
	s.Equal(http.StatusInternalServerError, resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	s.Contains(string(b), errMessage)
}
