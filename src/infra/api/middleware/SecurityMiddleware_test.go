package middleware

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golauth/golauth/src/application/token"
	"github.com/golauth/golauth/src/application/user/mock"
	"github.com/golauth/golauth/src/domain/entity"
	mock2 "github.com/golauth/golauth/src/domain/factory/mock"
	mock3 "github.com/golauth/golauth/src/domain/repository/mock"
	"github.com/golauth/golauth/src/infra/api/controller"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"testing"
)

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
	app.Use(NewSecurityMiddleware(token.NewValidateToken(key), "/").Apply())
	app.Get("/users/:id", userController.FindById)

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

		req, err := http.NewRequest("GET", "/users/37fe41b4-24bf-4da9-9124-615cc72865a5", nil)
		req.Header.Set("Content-Type", "application/json")
		bearerTk := fmt.Sprintf("Bearer %s", tk.AccessToken)
		req.Header.Set("Authorization", bearerTk)
		assert.NoError(t, err)

		findUserById.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(&entity.User{ID: uuid.MustParse("37fe41b4-24bf-4da9-9124-615cc72865a5")}, nil)

		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("invalid token", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/users/37fe41b4-24bf-4da9-9124-615cc72865a5", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer 123456")
		assert.NoError(t, err)

		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
