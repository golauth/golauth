package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	userApp "github.com/golauth/golauth/pkg/application/user"
	userMock "github.com/golauth/golauth/pkg/application/user/mock"
	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/golauth/golauth/pkg/domain/repository"
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

// post sends body to the signup route and returns the response.
func (s *SignupControllerSuite) post(body string) *http.Response {
	r, _ := http.NewRequest("POST", "/users", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	resp, err := s.app.Test(r, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	s.NoError(err)
	return resp
}

func (s *SignupControllerSuite) TestCreateUserOK() {
	// ToEntity() never carries an enabled flag: the request model no longer has
	// the field, so the use case always receives Enabled == false.
	want := &entity.User{
		Username:  "admin",
		FirstName: "User",
		LastName:  "Name",
		Email:     "em@il.com",
		Document:  "1234",
		Password:  "supersecret123",
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
	s.createUser.EXPECT().Execute(s.ctx, want).Return(savedUser, nil).Times(1)

	resp := s.post(`{"username":"admin","firstName":"User","lastName":"Name","email":"em@il.com","document":"1234","password":"supersecret123"}`)

	s.Equal(http.StatusCreated, resp.StatusCode)
	var output entity.User
	_ = json.NewDecoder(resp.Body).Decode(&output)
	s.Equal(savedUser.ID, output.ID)
}

// An "enabled" key in the body is ignored, not honoured: the field is gone from
// the request model, so the use case still receives Enabled == false.
func (s *SignupControllerSuite) TestCreateUserIgnoresEnabledField() {
	want := &entity.User{
		Username: "admin", FirstName: "User", LastName: "Name",
		Email: "em@il.com", Document: "1234", Password: "supersecret123",
	}
	s.createUser.EXPECT().Execute(s.ctx, want).Return(&entity.User{ID: uuid.New()}, nil).Times(1)

	resp := s.post(`{"username":"admin","firstName":"User","lastName":"Name","email":"em@il.com","document":"1234","password":"supersecret123","enabled":true}`)
	s.Equal(http.StatusCreated, resp.StatusCode)
}

// Signup answers unauthenticated callers, so the response must not carry the
// stored password hash back out, and the raw JSON must have no password key.
func (s *SignupControllerSuite) TestCreateUserDoesNotLeakPasswordHash() {
	const hash = "$2a$10$VNkiJ40.00IfVjxo8ILyauLUbnxMcKK2G/FbbwdsTYb.lCuZEbh22"
	savedUser := &entity.User{ID: uuid.New(), Username: "admin", Email: "em@il.com", Password: hash, Enabled: true}
	s.createUser.EXPECT().Execute(s.ctx, gomock.Any()).Return(savedUser, nil).Times(1)

	resp := s.post(`{"username":"admin","firstName":"User","lastName":"Name","email":"em@il.com","document":"1234","password":"supersecret123"}`)
	s.Equal(http.StatusCreated, resp.StatusCode)

	raw, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.NotContains(string(raw), hash)
	s.NotContains(strings.ToLower(string(raw)), `"password"`)
}

func (s *SignupControllerSuite) TestCreateUserErrBadRequest() {
	resp := s.post(`"invalid json"`)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

// A *user.ValidationError becomes a 400 whose body carries the field list.
func (s *SignupControllerSuite) TestCreateUserValidationErrorIsBadRequestWithFields() {
	ve := &userApp.ValidationError{Fields: []userApp.FieldError{
		{Field: "password", Message: "must be at least 12 characters"},
		{Field: "email", Message: "is not a valid e-mail address"},
	}}
	s.createUser.EXPECT().Execute(s.ctx, gomock.Any()).Return(nil, ve).Times(1)

	resp := s.post(`{"username":"admin","password":"short"}`)
	s.Equal(http.StatusBadRequest, resp.StatusCode)

	var body struct {
		Fields []userApp.FieldError `json:"fields"`
	}
	s.NoError(json.NewDecoder(resp.Body).Decode(&body))
	s.Len(body.Fields, 2)
	s.Equal("password", body.Fields[0].Field)
	s.Equal("must be at least 12 characters", body.Fields[0].Message)
}

// A duplicate username or e-mail is a 409, not a 500.
func (s *SignupControllerSuite) TestCreateUserDuplicateIsConflict() {
	dup := fmt.Errorf("could not save user: %w", fmt.Errorf("admin: %w", repository.ErrUserAlreadyExists))
	s.createUser.EXPECT().Execute(s.ctx, gomock.Any()).Return(nil, dup).Times(1)

	resp := s.post(`{"username":"admin","firstName":"User","lastName":"Name","email":"em@il.com","document":"1234","password":"supersecret123"}`)
	s.Equal(http.StatusConflict, resp.StatusCode)

	raw, _ := io.ReadAll(resp.Body)
	s.NotContains(string(raw), "23505")
	s.NotContains(strings.ToLower(string(raw)), "sql")
}

func (s *SignupControllerSuite) TestCreateUserErrSvc() {
	errMessage := "could not create new user"
	s.createUser.EXPECT().Execute(s.ctx, gomock.Any()).Return(nil, errors.New(errMessage)).Times(1)

	resp := s.post(`{"username":"admin","firstName":"User","lastName":"Name","email":"em@il.com","document":"1234","password":"supersecret123"}`)
	s.Equal(http.StatusInternalServerError, resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	s.Equal(errMessage, string(b))
}
