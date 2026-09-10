package bootstrap

import (
	"context"
	"errors"
	"testing"

	usermock "github.com/golauth/golauth/internal/application/user/mock"
	"github.com/golauth/golauth/internal/domain/entity"
	repomock "github.com/golauth/golauth/internal/domain/repository/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type deps struct {
	checker  *repomock.MockUserRepository
	create   *usermock.MockCreateUser
	roleRepo *repomock.MockRoleRepository
	grant    *usermock.MockAddUserRole
}

func newDeps(t *testing.T) deps {
	t.Helper()
	ctrl := gomock.NewController(t)
	return deps{
		checker:  repomock.NewMockUserRepository(ctrl),
		create:   usermock.NewMockCreateUser(ctrl),
		roleRepo: repomock.NewMockRoleRepository(ctrl),
		grant:    usermock.NewMockAddUserRole(ctrl),
	}
}

func (d deps) run(cfg Config) error {
	return EnsureAdmin(context.Background(), d.checker, d.create, d.roleRepo, d.grant, cfg)
}

// No admin in the database and no bootstrap variables: boot must fail, and the
// message must say exactly which variables to set.
func TestEnsureAdmin_NoAdminNoConfig_Fails(t *testing.T) {
	d := newDeps(t)
	d.checker.EXPECT().AdminExists(gomock.Any()).Return(false, nil)

	err := d.run(Config{})
	require.ErrorIs(t, err, ErrNoAdmin)
	require.Contains(t, err.Error(), "BOOTSTRAP_ADMIN_USER")
	require.Contains(t, err.Error(), "BOOTSTRAP_ADMIN_PASSWORD")
}

// A password present but no username still fails: both are required.
func TestEnsureAdmin_PartialConfig_Fails(t *testing.T) {
	d := newDeps(t)
	d.checker.EXPECT().AdminExists(gomock.Any()).Return(false, nil)

	require.ErrorIs(t, d.run(Config{Password: "a-strong-passphrase"}), ErrNoAdmin)
}

// With both variables set and no admin present, exactly one user is created,
// the ADMIN role is looked up once and granted once. A missing Email is
// derived, not required.
func TestEnsureAdmin_CreatesExactlyOneAdmin(t *testing.T) {
	d := newDeps(t)
	adminID := uuid.New()
	roleID := uuid.New()

	d.checker.EXPECT().AdminExists(gomock.Any()).Return(false, nil)
	d.create.EXPECT().Execute(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, u *entity.User) (*entity.User, error) {
			require.Equal(t, "root", u.Username)
			require.Equal(t, "root@bootstrap.local", u.Email)
			require.NotEmpty(t, u.FirstName)
			require.NotEmpty(t, u.LastName)
			require.NotEmpty(t, u.Document)
			u.ID = adminID
			return u, nil
		}).Times(1)
	d.roleRepo.EXPECT().FindByName(gomock.Any(), "ADMIN").
		Return(&entity.Role{ID: roleID, Name: "ADMIN"}, nil).Times(1)
	d.grant.EXPECT().Execute(gomock.Any(), adminID, roleID).Return(nil).Times(1)

	require.NoError(t, d.run(Config{Username: "root", Password: "a-strong-passphrase"}))
}

// A second boot, with an admin already present, creates nothing.
func TestEnsureAdmin_ExistingAdmin_NoOp(t *testing.T) {
	d := newDeps(t)
	d.checker.EXPECT().AdminExists(gomock.Any()).Return(true, nil)
	// No create/roleRepo/grant expectations: any call fails the test.

	require.NoError(t, d.run(Config{Username: "root", Password: "a-strong-passphrase"}))
}

// A weak bootstrap password is rejected by the CreateUser policy and surfaces
// as a boot failure rather than a silent weak account.
func TestEnsureAdmin_PropagatesCreateUserError(t *testing.T) {
	d := newDeps(t)
	d.checker.EXPECT().AdminExists(gomock.Any()).Return(false, nil)
	d.create.EXPECT().Execute(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("password: must be at least 12 characters")).Times(1)

	err := d.run(Config{Username: "root", Password: "short"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "bootstrap: creating the administrator")
}

// An error from the admin check aborts the boot; nothing is created.
func TestEnsureAdmin_PropagatesCheckError(t *testing.T) {
	d := newDeps(t)
	d.checker.EXPECT().AdminExists(gomock.Any()).Return(false, errors.New("db down"))

	require.Error(t, d.run(Config{Username: "root", Password: "a-strong-passphrase"}))
}
