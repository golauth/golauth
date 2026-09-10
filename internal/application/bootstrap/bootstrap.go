// Package bootstrap makes sure a fresh installation has exactly one way in:
// an administrator created from configuration, never a credential baked into
// the repository.
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/golauth/golauth/internal/application/user"
	"github.com/golauth/golauth/internal/domain/entity"
	"github.com/golauth/golauth/internal/domain/repository"
)

// ErrNoAdmin is returned when the database holds no administrator and the
// BOOTSTRAP_ADMIN_* variables are unset: there is nothing to keep and no
// instruction to create one, so the caller must abort the boot rather than come
// up unreachable.
var ErrNoAdmin = errors.New(
	"no administrator exists and BOOTSTRAP_ADMIN_USER / BOOTSTRAP_ADMIN_PASSWORD are not set")

const adminRoleName = "ADMIN"

// Config is the bootstrap-admin input, read from the environment by the caller.
// Email is optional and defaults to <username>@bootstrap.local.
type Config struct {
	Username string
	Password string
	Email    string
}

// AdminChecker reports whether the service already has a usable administrator.
type AdminChecker interface {
	AdminExists(ctx context.Context) (bool, error)
}

// EnsureAdmin guarantees the service has at least one administrator:
//
//   - one already exists            -> nothing changes
//   - none, Config is complete      -> one is created, granted ADMIN, and logged
//   - none, Config is incomplete    -> ErrNoAdmin
//
// The user is created through the same CreateUser use case as a public signup,
// so the bootstrap password must satisfy the full password policy; it is then
// also granted the ADMIN role.
func EnsureAdmin(
	ctx context.Context,
	checker AdminChecker,
	createUser user.CreateUser,
	roleRepo repository.RoleRepository,
	grantRole user.AddUserRole,
	cfg Config,
) error {
	exists, err := checker.AdminExists(ctx)
	if err != nil {
		return fmt.Errorf("bootstrap: checking for an administrator: %w", err)
	}
	if exists {
		slog.InfoContext(ctx, "bootstrap: an administrator already exists; nothing to do")
		return nil
	}

	cfg.Username = strings.TrimSpace(cfg.Username)
	if cfg.Username == "" || cfg.Password == "" {
		return ErrNoAdmin
	}
	email := strings.TrimSpace(cfg.Email)
	if email == "" {
		email = cfg.Username + "@bootstrap.local"
	}

	created, err := createUser.Execute(ctx, &entity.User{
		Username:  cfg.Username,
		Password:  cfg.Password,
		Email:     email,
		FirstName: "Bootstrap",
		LastName:  "Administrator",
		Document:  "bootstrap",
	})
	if err != nil {
		return fmt.Errorf("bootstrap: creating the administrator: %w", err)
	}

	adminRole, err := roleRepo.FindByName(ctx, adminRoleName)
	if err != nil {
		return fmt.Errorf("bootstrap: loading the %s role: %w", adminRoleName, err)
	}
	if err := grantRole.Execute(ctx, created.ID, adminRole.ID); err != nil {
		return fmt.Errorf("bootstrap: granting %s: %w", adminRoleName, err)
	}

	slog.WarnContext(ctx, "bootstrap: created the initial administrator from BOOTSTRAP_ADMIN_*",
		"username", created.Username)
	return nil
}
