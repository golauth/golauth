package middleware

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/application/token"
	"github.com/golauth/golauth/pkg/application/user/mock"
	"github.com/golauth/golauth/pkg/domain/entity"
	mock2 "github.com/golauth/golauth/pkg/domain/factory/mock"
	mock3 "github.com/golauth/golauth/pkg/domain/repository/mock"
	"github.com/golauth/golauth/pkg/infra/api/controller"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// pathPrefix mirrors the prefix the router actually mounts the API under.
// The suite used to pass "/", which made every allowlist key ("//token") a
// string no real request could produce -- so the public-path logic was never
// really exercised here.
const pathPrefix = "/auth"

func TestSecurityMiddleware(t *testing.T) {
	username := "admin"
	password := "admin123"
	passwordEncoded := "$2a$10$VNkiJ40.00IfVjxo8ILyauLUbnxMcKK2G/FbbwdsTYb.lCuZEbh22"
	ctrl := gomock.NewController(t)
	findUserById := mock.NewMockFindUserById(ctrl)
	addUserRole := mock.NewMockAddUserRole(ctrl)
	userController := controller.NewUserController(findUserById, addUserRole)

	key := token.GeneratePrivateKey()

	app := fiber.New()
	app.Use(NewSecurityMiddleware(token.NewValidateToken(key), pathPrefix).Apply())
	app.Get(pathPrefix+"/users/:id", userController.FindById)
	app.Post(pathPrefix+"/token", func(ctx fiber.Ctx) error {
		return ctx.SendStatus(http.StatusOK)
	})

	t.Run("valid token", func(t *testing.T) {
		userRepository := mock3.NewMockUserRepository(ctrl)
		userAuthorityRepository := mock3.NewMockUserAuthorityRepository(ctrl)
		roleRepository := mock3.NewMockRoleRepository(ctrl)
		userRoleRepository := mock3.NewMockUserRoleRepository(ctrl)

		repoFactory := mock2.NewMockRepositoryFactory(ctrl)
		repoFactory.EXPECT().NewUserRepository().Return(userRepository)
		repoFactory.EXPECT().NewUserAuthorityRepository().Return(userAuthorityRepository)
		repoFactory.EXPECT().NewRoleRepository().Return(roleRepository)
		repoFactory.EXPECT().NewUserRoleRepository().Return(userRoleRepository)

		userRepository.EXPECT().FindByUsername(gomock.Any(), "admin").Return(&entity.User{Username: username, Password: passwordEncoded}, nil)
		userAuthorityRepository.EXPECT().FindAuthoritiesByUserID(gomock.Any(), gomock.Any()).Return([]string{"ADMIN"}, nil)

		generateJwtToken := token.NewGenerateJwtToken(key)
		generateToken := token.NewGenerateToken(repoFactory, generateJwtToken)

		tk, err := generateToken.Execute(context.Background(), username, password)
		assert.NoError(t, err)

		req, err := http.NewRequest("GET", pathPrefix+"/users/37fe41b4-24bf-4da9-9124-615cc72865a5", nil)
		req.Header.Set("Content-Type", "application/json")
		bearerTk := fmt.Sprintf("Bearer %s", tk.AccessToken)
		req.Header.Set("Authorization", bearerTk)
		assert.NoError(t, err)

		findUserById.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(&entity.User{ID: uuid.MustParse("37fe41b4-24bf-4da9-9124-615cc72865a5")}, nil)

		resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("invalid token", func(t *testing.T) {
		req, err := http.NewRequest("GET", pathPrefix+"/users/37fe41b4-24bf-4da9-9124-615cc72865a5", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer 123456")
		assert.NoError(t, err)

		resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	// A missing or non-bearer header is an authentication failure, not a server
	// fault; it used to surface as 500.
	t.Run("missing and malformed headers are unauthorized", func(t *testing.T) {
		for name, header := range map[string]string{
			"absent": "",
			"basic":  "Basic dXNlcjpwYXNzd29yZA==",
			"empty":  "Bearer ",
		} {
			req, _ := http.NewRequest("GET", pathPrefix+"/users/37fe41b4-24bf-4da9-9124-615cc72865a5", nil)
			if header != "" {
				req.Header.Set("Authorization", header)
			}
			resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
			assert.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "header case %q", name)
		}
	})

	// The allowlist is matched against the request path. It previously compared
	// the absolute URI ("http://host/auth/token"), so no entry ever matched.
	t.Run("public path is served without a token", func(t *testing.T) {
		for _, target := range []string{
			pathPrefix + "/token",
			pathPrefix + "/token/",
			pathPrefix + "/token?redirect=/somewhere",
		} {
			req, _ := http.NewRequest("POST", target, nil)
			resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode, "target %q", target)
		}
	})
}

func TestIsPrivateURI(t *testing.T) {
	s := NewSecurityMiddleware(nil, pathPrefix)

	for path, private := range map[string]bool{
		"/auth/token":            false,
		"/auth/signup":           false,
		"/auth/check_token":      false,
		"/auth/users/some-id":    true,
		"/auth/roles/ADMIN":      true,
		"/auth/roles":            true,
		"/token":                 true,
		"/auth":                  true,
		"/anything/unmapped":     true,
		"/auth/token/../users/x": true,
	} {
		assert.Equal(t, private, s.isPrivateURI(path), "path %q", path)
	}
}
