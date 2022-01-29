package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	mock2 "github.com/golauth/golauth/domain/usecase/user/mock"
	"github.com/golauth/golauth/infra/api/controller/model"
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
	createUser *mock2.MockCreateUser

	ctrl SignupController
}

func TestSignupController(t *testing.T) {
	suite.Run(t, new(SignupControllerSuite))
}

func (s *SignupControllerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	s.ctx = context.Background()
	s.createUser = mock2.NewMockCreateUser(s.mockCtrl)

	s.ctrl = NewSignupController(s.createUser)
}

func (s *SignupControllerSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s SignupControllerSuite) TestCreateUserOK() {
	user := model.UserRequest{
		Username:  "admin",
		FirstName: "User",
		LastName:  "Name",
		Email:     "em@il.com",
		Document:  "1234",
		Password:  "4567",
		Enabled:   true,
	}
	savedUser := model.UserResponse{
		ID:           uuid.New(),
		Username:     "admin",
		FirstName:    "User",
		LastName:     "Name",
		Email:        "em@il.com",
		Document:     "1234",
		Enabled:      true,
		CreationDate: time.Now().Add(-5 * time.Second),
	}
	s.createUser.EXPECT().Execute(s.ctx, user).Return(savedUser, nil).Times(1)

	body, _ := json.Marshal(user)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/users", strings.NewReader(string(body)))

	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	_ = jsonEncoder.Encode(savedUser)

	s.ctrl.CreateUser(w, r)
	s.Equal(http.StatusCreated, w.Code)
	s.Equal(bf, w.Body)
}

func (s SignupControllerSuite) TestCreateUserErrSvc() {
	user := model.UserRequest{
		Username:  "admin",
		FirstName: "User",
		LastName:  "Name",
		Email:     "em@il.com",
		Document:  "1234",
		Password:  "4567",
		Enabled:   true,
	}
	errMessage := "could not create new user"
	s.createUser.EXPECT().Execute(s.ctx, user).Return(model.UserResponse{}, fmt.Errorf(errMessage)).Times(1)

	body, _ := json.Marshal(user)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/users", strings.NewReader(string(body)))

	s.ctrl.CreateUser(w, r)
	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), errMessage)
}