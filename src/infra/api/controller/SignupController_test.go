package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	userMock "github.com/golauth/golauth/src/application/user/mock"
	"github.com/golauth/golauth/src/domain/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
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

	ctrl SignupController
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
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/users", strings.NewReader(string(body)))

	s.ctrl.CreateUser(w, r)
	s.Equal(http.StatusCreated, w.Code)
	var output entity.User
	_ = json.Unmarshal(w.Body.Bytes(), &output)
	s.Equal(savedUser.ID, output.ID)
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
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/users", strings.NewReader(string(body)))

	s.ctrl.CreateUser(w, r)
	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), errMessage)
}
