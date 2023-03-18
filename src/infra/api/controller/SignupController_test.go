package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	userMock "github.com/golauth/golauth/src/application/user/mock"
	"github.com/golauth/golauth/src/domain/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type SignupControllerSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl   *gomock.Controller
	ctx        context.Context
	createUser *userMock.MockCreateUser
	app        *fiber.App
	ctrl       SignupController
}

func TestSignupController(t *testing.T) {
	suite.Run(t, new(SignupControllerSuite))
}

func (s *SignupControllerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	s.ctx = context.Background()
	s.createUser = userMock.NewMockCreateUser(s.mockCtrl)

	s.ctrl = NewSignupController(s.createUser)
	s.app = fiber.New()
	s.app.Post("/users", s.ctrl.CreateUser)
}

func (s *SignupControllerSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *SignupControllerSuite) TestCreateUserOK() {
	input := &entity.User{
		Username:  "admin",
		FirstName: "User",
		LastName:  "Name",
		Email:     "em@il.com",
		Document:  "1234",
		Password:  "4567",
		Enabled:   true,
	}
	savedUser := &entity.User{
		ID:           uuid.New(),
		Username:     "admin",
		FirstName:    "User",
		LastName:     "Name",
		Email:        "em@il.com",
		Document:     "1234",
		Enabled:      true,
		CreationDate: time.Now().Add(-5 * time.Second),
	}
	s.createUser.EXPECT().Execute(s.ctx, input).Return(savedUser, nil).Times(1)

	body, _ := json.Marshal(input)
	r, _ := http.NewRequest("POST", "/users", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	resp, _ := s.app.Test(r, -1)

	s.Equal(http.StatusCreated, resp.StatusCode)
	var output entity.User
	_ = json.NewDecoder(resp.Body).Decode(&output)
	s.Equal(savedUser.ID, output.ID)
}

func (s *SignupControllerSuite) TestCreateUserErrBadRequest() {
	body, _ := json.Marshal("invalid json")
	r, _ := http.NewRequest("POST", "/users", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *SignupControllerSuite) TestCreateUserErrSvc() {
	user := &entity.User{
		Username:  "admin",
		FirstName: "User",
		LastName:  "Name",
		Email:     "em@il.com",
		Document:  "1234",
		Password:  "4567",
		Enabled:   true,
	}
	errMessage := "could not create new user"
	s.createUser.EXPECT().Execute(s.ctx, user).Return(nil, fmt.Errorf(errMessage)).Times(1)

	body, _ := json.Marshal(user)
	r, _ := http.NewRequest("POST", "/users", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	resp, _ := s.app.Test(r, -1)
	s.Equal(http.StatusInternalServerError, resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	s.Equal(errMessage, string(b))
}
