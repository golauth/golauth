package user

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/golauth/golauth/src/domain/entity"
	"github.com/golauth/golauth/src/domain/repository/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type FindUserByIdSuite struct {
	suite.Suite
	*require.Assertions
	mockCtrl *gomock.Controller

	userRepository *mock.MockUserRepository

	ctx     context.Context
	finding FindUserById
}

func TestFindUserById(t *testing.T) {
	suite.Run(t, new(FindUserByIdSuite))
}

func (s *FindUserByIdSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockCtrl = gomock.NewController(s.T())

	s.userRepository = mock.NewMockUserRepository(s.mockCtrl)

	s.ctx = context.Background()
	s.finding = NewFindUserById(s.userRepository)
}

func (s *FindUserByIdSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *FindUserByIdSuite) TestFindByIdOK() {
	id := uuid.New()
	user := &entity.User{
		ID:           id,
		Username:     "admin",
		FirstName:    "User",
		LastName:     "Name",
		Email:        "em@ail.com",
		Document:     "1234",
		Password:     "1234c",
		Enabled:      true,
		CreationDate: time.Now().AddDate(-1, 0, 0),
	}
	s.userRepository.EXPECT().FindByID(s.ctx, id).Return(user, nil).Times(1)

	output, err := s.finding.Execute(s.ctx, id)
	s.NoError(err)
	s.Equal(user, output)
}

func (s *FindUserByIdSuite) TestFindByIDErr() {
	id := uuid.New()
	s.userRepository.EXPECT().FindByID(s.ctx, id).Return(nil, fmt.Errorf("could not find user")).Times(1)

	resp, err := s.finding.Execute(s.ctx, id)
	s.Zero(resp)
	s.Error(err)
	s.ErrorAs(fmt.Errorf("could not find user"), &err)
}
