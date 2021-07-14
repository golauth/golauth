package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golauth/entity"
	"golauth/model"
	mockSvc "golauth/usecase/mock"
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
	//urRepo  *mock.MockUserRoleRepository
	uc UserController
}

func TestUserControllerSuite(t *testing.T) {
	suite.Run(t, new(UserControllerSuite))
}

func (s *UserControllerSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.ctrl = gomock.NewController(s.T())
	s.userSvc = mockSvc.NewMockUserService(s.ctrl)
	//s.urRepo = mock.NewMockUserRoleRepository(s.ctrl)

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
