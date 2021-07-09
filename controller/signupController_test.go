package controller

import (
	"bytes"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golauth/model"
	"golauth/usecase/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type SignupControllerSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller
	svc      *mock.MockUserService

	ctrl SignupController
}

func TestSignupController(t *testing.T) {
	suite.Run(t, new(SignupControllerSuite))
}

func (s *SignupControllerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())
	s.svc = mock.NewMockUserService(s.mockCtrl)

	s.ctrl = NewSignupController(s.svc)
}

func (s *SignupControllerSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s SignupControllerSuite) TestCreateUser() {
	user := model.User{
		Username:  "admin",
		FirstName: "User",
		LastName:  "Name",
		Email:     "em@il.com",
		Document:  "1234",
		Password:  "4567",
		Enabled:   true,
	}
	savedUser := model.User{
		Username:  "admin",
		FirstName: "User",
		LastName:  "Name",
		Email:     "em@il.com",
		Document:  "1234",
		Password:  "4567",
		Enabled:   true,
	}
	s.svc.EXPECT().CreateUser(user).Return(savedUser, nil).Times(1)

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
