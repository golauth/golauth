//go:generate mockgen -source GenerateToken.go -destination mock/GenerateToken_mock.go -package mock
package token

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golauth/golauth/pkg/application/audit"
	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/golauth/golauth/pkg/domain/factory"
	"github.com/golauth/golauth/pkg/domain/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUsernameOrPassword = errors.New("invalid username or password")
	ErrGeneratingToken           = errors.New("error generating token")
)

// dummyBcryptHash is a valid cost-10 bcrypt hash of a value no one knows. It is
// compared against the supplied password when the username does not exist, so
// the unknown-user path pays the same CPU cost as the wrong-password path and
// response time no longer leaks which usernames are real.
const dummyBcryptHash = "$2a$10$A9DfQ7RA4ojiq1Pi0DCyJO5/yz4G2wZl1HhTBnRsizXcyKptHZ5cO"

// comparePassword is a seam for tests to observe that the bcrypt comparison
// runs on the unknown-user path; production always uses bcrypt.
var comparePassword = bcrypt.CompareHashAndPassword

type GenerateToken interface {
	Execute(ctx context.Context, username, password, clientIP, userAgent string) (*entity.Token, error)
}

func NewGenerateToken(repoFactory factory.RepositoryFactory, jwtToken GenerateJwtToken, lockout LockoutPolicy, cfg Config) GenerateToken {
	return generateToken{
		userRepository:         repoFactory.NewUserRepository(),
		roleRepository:         repoFactory.NewRoleRepository(),
		userRoleRepository:     repoFactory.NewUserRoleRepository(),
		loginAttemptRepository: repoFactory.NewLoginAttemptRepository(),
		lockout:                lockout.withDefaults(),
		issuer: tokenIssuer{
			refreshTokenRepository:  repoFactory.NewRefreshTokenRepository(),
			userAuthorityRepository: repoFactory.NewUserAuthorityRepository(),
			jwtToken:                jwtToken,
			cfg:                     cfg,
		},
	}
}

type generateToken struct {
	userRepository         repository.UserRepository
	roleRepository         repository.RoleRepository
	userRoleRepository     repository.UserRoleRepository
	loginAttemptRepository repository.LoginAttemptRepository
	lockout                LockoutPolicy
	issuer                 tokenIssuer
}

func (uc generateToken) Execute(ctx context.Context, username, password, clientIP, userAgent string) (*entity.Token, error) {
	user, err := uc.userRepository.FindByUsername(ctx, username)
	if err != nil {
		// Pay the bcrypt cost even though there is nothing to compare against,
		// so an unknown username is indistinguishable from a wrong password in
		// latency, status and body.
		_ = comparePassword([]byte(dummyBcryptHash), []byte(password))
		logFailedLogin(ctx, username, clientIP, "unknown_user")
		return nil, ErrInvalidUsernameOrPassword
	}

	now := time.Now()
	attempt, err := uc.loginAttemptRepository.Get(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("could not read login attempts: %w", err)
	}

	// A locked account fails before bcrypt runs: this both enforces the lock
	// and keeps a locked account from being a CPU-exhaustion lever.
	if attempt != nil && attempt.Locked(now) {
		logFailedLogin(ctx, username, clientIP, "locked")
		return nil, ErrInvalidUsernameOrPassword
	}

	if comparePassword([]byte(user.Password), []byte(password)) != nil {
		uc.registerFailure(ctx, user.ID, attempt, now)
		logFailedLogin(ctx, username, clientIP, "bad_password")
		return nil, ErrInvalidUsernameOrPassword
	}

	// A deactivated account must not be able to log in. Same error as a wrong
	// password on purpose, so the endpoint cannot be used to probe account
	// state; the real reason is only in the log.
	if !user.Enabled {
		logFailedLogin(ctx, username, clientIP, "disabled")
		return nil, ErrInvalidUsernameOrPassword
	}

	// Correct password on an active account: clear the failure counter.
	if attempt != nil {
		if resetErr := uc.loginAttemptRepository.Reset(ctx, user.ID); resetErr != nil {
			slog.WarnContext(ctx, "could not reset login attempts", "user_id", user.ID.String(), "err", resetErr.Error())
		}
	}

	token, _, err := uc.issuer.issuePair(ctx, user, clientIP, userAgent)
	if err != nil {
		if errors.Is(err, ErrGeneratingToken) {
			return nil, ErrGeneratingToken
		}
		return nil, err
	}

	audit.Event(ctx, audit.LoginSucceeded,
		"user_id", user.ID.String(),
		"username", user.Username,
		"client_ip", clientIP,
	)
	return token, nil
}

// registerFailure records one more consecutive failure and, once the policy
// threshold is crossed, sets a growing lock window.
func (uc generateToken) registerFailure(ctx context.Context, userID uuid.UUID, prev *entity.LoginAttempt, now time.Time) {
	failures := 1
	if prev != nil {
		failures = prev.FailedCount + 1
	}
	var lockedUntil time.Time
	if d := uc.lockout.lockDuration(failures); d > 0 {
		lockedUntil = now.Add(d)
	}
	if err := uc.loginAttemptRepository.RegisterFailure(ctx, userID, lockedUntil); err != nil {
		slog.WarnContext(ctx, "could not register login failure", "user_id", userID.String(), "err", err.Error())
	}
}

// logFailedLogin emits one audit event per rejected login. outcome is the real
// reason (unknown_user, locked, bad_password, disabled); the caller still
// returns the single vague error so the response leaks none of it.
func logFailedLogin(ctx context.Context, username, clientIP, outcome string) {
	audit.Event(ctx, audit.LoginFailed,
		"username", username,
		"client_ip", clientIP,
		"outcome", outcome,
	)
}
