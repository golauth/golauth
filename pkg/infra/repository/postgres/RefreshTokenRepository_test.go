package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/golauth/golauth/pkg/domain/repository"
	"github.com/golauth/golauth/pkg/infra/database"
	"github.com/golauth/golauth/tests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const seededUserID = "8c61f220-8bb8-48b9-b225-d54dfa6503db" // admin, from add-users.sql

type RefreshTokenRepositorySuite struct {
	suite.Suite
	*require.Assertions
	db     database.Database
	repo   repository.RefreshTokenRepository
	userID uuid.UUID
}

func TestRefreshTokenRepository(t *testing.T) {
	ctxContainer, err := tests.ContainerDBStart("./../../../..")
	assert.NoError(t, err)
	s := new(RefreshTokenRepositorySuite)
	suite.Run(t, s)
	tests.ContainerDBStop(ctxContainer)
}

func (s *RefreshTokenRepositorySuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.db = database.NewPGDatabase()
	s.repo = NewRefreshTokenRepository(s.db)
	s.userID = uuid.MustParse(seededUserID)
	s.NoError(tests.DatasetTest(s.db, "./../../../..", "clear-data.sql", "add-users.sql"))
}

func (s *RefreshTokenRepositorySuite) TearDownTest() { s.db.Close() }

func (s *RefreshTokenRepositorySuite) newToken(hash string, expiresAt time.Time) *entity.RefreshToken {
	created, err := s.repo.Create(context.Background(), &entity.RefreshToken{
		UserID:    s.userID,
		TokenHash: hash,
		IssuedAt:  time.Now(),
		ExpiresAt: expiresAt,
		UserAgent: "curl/8",
		ClientIP:  "203.0.113.9",
	})
	s.NoError(err)
	s.NotEqual(uuid.Nil, created.ID)
	return created
}

func (s *RefreshTokenRepositorySuite) TestCreateAndFindByHash() {
	s.newToken("hash-aaa", time.Now().Add(time.Hour))

	got, err := s.repo.FindByHash(context.Background(), "hash-aaa")
	s.NoError(err)
	s.Require().NotNil(got)
	s.Equal(s.userID, got.UserID)
	s.Equal("curl/8", got.UserAgent)
	s.Equal("203.0.113.9", got.ClientIP)
	s.Nil(got.RevokedAt)
	s.Nil(got.ReplacedBy)
	s.True(got.Usable(time.Now()))
}

func (s *RefreshTokenRepositorySuite) TestFindByHashMissingReturnsNilNil() {
	got, err := s.repo.FindByHash(context.Background(), "does-not-exist")
	s.NoError(err)
	s.Nil(got)
}

func (s *RefreshTokenRepositorySuite) TestReplaceMarksRotated() {
	old := s.newToken("hash-old", time.Now().Add(time.Hour))
	next := s.newToken("hash-new", time.Now().Add(time.Hour))

	s.NoError(s.repo.Replace(context.Background(), old.ID, next.ID))

	got, err := s.repo.FindByHash(context.Background(), "hash-old")
	s.NoError(err)
	s.Require().NotNil(got)
	s.True(got.Rotated())
	s.Require().NotNil(got.ReplacedBy)
	s.Equal(next.ID, *got.ReplacedBy)
	s.Require().NotNil(got.RevokedAt)
	s.False(got.Usable(time.Now()))
}

func (s *RefreshTokenRepositorySuite) TestRevoke() {
	tk := s.newToken("hash-rev", time.Now().Add(time.Hour))

	s.NoError(s.repo.Revoke(context.Background(), tk.ID))

	got, err := s.repo.FindByHash(context.Background(), "hash-rev")
	s.NoError(err)
	s.Require().NotNil(got)
	s.True(got.Revoked())
}

func (s *RefreshTokenRepositorySuite) TestRevokeAllForUser() {
	s.newToken("hash-1", time.Now().Add(time.Hour))
	s.newToken("hash-2", time.Now().Add(time.Hour))
	already := s.newToken("hash-3", time.Now().Add(time.Hour))
	s.NoError(s.repo.Revoke(context.Background(), already.ID))

	n, err := s.repo.RevokeAllForUser(context.Background(), s.userID)
	s.NoError(err)
	s.Equal(int64(2), n, "only the not-yet-revoked rows count")

	for _, h := range []string{"hash-1", "hash-2", "hash-3"} {
		got, err := s.repo.FindByHash(context.Background(), h)
		s.NoError(err)
		s.Require().NotNil(got)
		s.True(got.Revoked())
	}
}

func (s *RefreshTokenRepositorySuite) TestDeleteExpired() {
	s.newToken("hash-live", time.Now().Add(time.Hour))
	s.newToken("hash-dead-1", time.Now().Add(-time.Hour))
	s.newToken("hash-dead-2", time.Now().Add(-time.Minute))

	n, err := s.repo.DeleteExpired(context.Background(), time.Now())
	s.NoError(err)
	s.Equal(int64(2), n)

	live, err := s.repo.FindByHash(context.Background(), "hash-live")
	s.NoError(err)
	s.NotNil(live)
	dead, err := s.repo.FindByHash(context.Background(), "hash-dead-1")
	s.NoError(err)
	s.Nil(dead)
}

// A unique index guards the token hash: a second insert of the same digest fails.
func (s *RefreshTokenRepositorySuite) TestDuplicateHashRejected() {
	s.newToken("hash-dup", time.Now().Add(time.Hour))
	_, err := s.repo.Create(context.Background(), &entity.RefreshToken{
		UserID:    s.userID,
		TokenHash: "hash-dup",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	})
	s.Error(err)
}
