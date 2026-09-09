package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	userMock "github.com/golauth/golauth/pkg/application/user/mock"
	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
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

	resp, _ := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})

	s.Equal(http.StatusCreated, resp.StatusCode)
	var output entity.User
	_ = json.NewDecoder(resp.Body).Decode(&output)
	s.Equal(savedUser.ID, output.ID)
}

// Signup answers unauthenticated callers, so the response must not carry the
// stored password hash back out. The entity was previously serialized as-is.
func (s *SignupControllerSuite) TestCreateUserDoesNotLeakPasswordHash() {
	const hash = "$2a$10$VNkiJ40.00IfVjxo8ILyauLUbnxMcKK2G/FbbwdsTYb.lCuZEbh22"
	input := &entity.User{Username: "admin", Email: "em@il.com", Password: "4567", Enabled: true}
	savedUser := &entity.User{ID: uuid.New(), Username: "admin", Email: "em@il.com", Password: hash, Enabled: true}
	s.createUser.EXPECT().Execute(s.ctx, input).Return(savedUser, nil).Times(1)

	body, _ := json.Marshal(input)
	r, _ := http.NewRequest("POST", "/users", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	resp, _ := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Equal(http.StatusCreated, resp.StatusCode)

	raw, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.NotContains(string(raw), hash)
	s.NotContains(strings.ToLower(string(raw)), `"password"`)
}

func (s *SignupControllerSuite) TestCreateUserErrBadRequest() {
	body, _ := json.Marshal("invalid json")
	r, _ := http.NewRequest("POST", "/users", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	resp, _ := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
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
	s.createUser.EXPECT().Execute(s.ctx, user).Return(nil, errors.New(errMessage)).Times(1)

	body, _ := json.Marshal(user)
	r, _ := http.NewRequest("POST", "/users", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")

	resp, _ := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.Equal(http.StatusInternalServerError, resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	s.Equal(errMessage, string(b))
}
