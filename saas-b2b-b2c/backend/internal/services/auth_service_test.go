package services

import (
	"errors"
	"testing"

	"franchise-saas-backend/internal/mocks"
	"franchise-saas-backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// TestAuthService_GenerateTokens_NoDefaultSecret проверяет fail-closed:
// без настроенного jwt_secret токены не выпускаются.
func TestAuthService_GenerateTokens_NoDefaultSecret(t *testing.T) {
	viper.Set("jwt_secret", "")
	defer viper.Set("jwt_secret", "")

	svc := NewAuthServiceWithInterface(nil, nil)
	_, _, err := svc.GenerateTokens(uuid.New(), "user@test.com", models.RoleDealer, nil, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "jwt_secret")
}

// TestAuthService_GenerateTokens_WithSecret проверяет корректную подпись HS256.
func TestAuthService_GenerateTokens_WithSecret(t *testing.T) {
	const secret = "unit-test-secret"
	viper.Set("jwt_secret", secret)
	defer viper.Set("jwt_secret", "")

	userID := uuid.New()
	svc := NewAuthServiceWithInterface(nil, nil)

	access, refresh, err := svc.GenerateTokens(userID, "user@test.com", models.RoleFranchisor, nil, nil)
	require.NoError(t, err)
	require.NotEmpty(t, access)
	require.NotEmpty(t, refresh)

	token, err := jwt.Parse(access, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	require.NoError(t, err)
	require.True(t, token.Valid)

	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)
	assert.Equal(t, userID.String(), claims["user_id"])
	assert.Equal(t, string(models.RoleFranchisor), claims["role"])
}

// TestAuthService_Authenticate_BlockedUser проверяет, что заблокированный
// пользователь не может войти.
func TestAuthService_Authenticate_BlockedUser(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	require.NoError(t, err)

	user := &models.User{
		ID:           uuid.New(),
		Email:        "blocked@test.com",
		PasswordHash: string(hash),
		Role:         models.RoleDealer,
		Status:       "blocked",
	}

	repo := mocks.NewMockUserRepository()
	repo.On("GetUserByEmail", mock.Anything, "blocked@test.com").Return(user, nil)
	repo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)

	svc := NewAuthServiceWithInterface(repo, nil)
	_, err = svc.Authenticate("blocked@test.com", "password123")

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUserBlocked))
	repo.AssertExpectations(t)
}

// TestAuthService_Authenticate_ActiveUser регресс: активный вход работает.
func TestAuthService_Authenticate_ActiveUser(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	require.NoError(t, err)

	user := &models.User{
		ID:           uuid.New(),
		Email:        "active@test.com",
		PasswordHash: string(hash),
		Role:         models.RoleDealer,
		Status:       "active",
	}

	repo := mocks.NewMockUserRepository()
	repo.On("GetUserByEmail", mock.Anything, "active@test.com").Return(user, nil)
	repo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)

	svc := NewAuthServiceWithInterface(repo, nil)
	got, err := svc.Authenticate("active@test.com", "password123")

	require.NoError(t, err)
	assert.Equal(t, user.ID, got.ID)
	repo.AssertExpectations(t)
}

// TestAuthService_Logout_MissingSecret проверяет fail-closed при пустом секрете.
func TestAuthService_Logout_MissingSecret(t *testing.T) {
	viper.Set("jwt_secret", "")
	defer viper.Set("jwt_secret", "")

	svc := NewAuthServiceWithInterface(nil, nil)
	err := svc.Logout("some.jwt.token")
	require.Error(t, err)
}

// TestAuthService_Logout_ValidToken проверяет успешный отзыв refresh-токена.
func TestAuthService_Logout_ValidToken(t *testing.T) {
	const secret = "logout-secret"
	viper.Set("jwt_secret", secret)
	defer viper.Set("jwt_secret", "")

	svc := NewAuthServiceWithInterface(nil, nil)
	_, refresh, err := svc.GenerateTokens(uuid.New(), "logout@test.com", models.RoleDealer, nil, nil)
	require.NoError(t, err)

	require.NoError(t, svc.Logout(refresh))
}

// TestAuthService_RefreshTokens_NoSecret проверяет fail-closed при пустом секрете.
func TestAuthService_RefreshTokens_NoSecret(t *testing.T) {
	viper.Set("jwt_secret", "")
	defer viper.Set("jwt_secret", "")

	svc := NewAuthServiceWithInterface(nil, nil)
	_, _, err := svc.RefreshTokens("irrelevant")
	require.Error(t, err)
}

// TestAuthService_RefreshTokens_Rotation проверяет обновление пары токенов.
func TestAuthService_RefreshTokens_Rotation(t *testing.T) {
	const secret = "refresh-secret"
	viper.Set("jwt_secret", secret)
	defer viper.Set("jwt_secret", "")

	user := &models.User{
		ID:    uuid.New(),
		Email: "rotate@test.com",
		Role:  models.RoleDealer,
	}

	repo := mocks.NewMockUserRepository()
	repo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)

	svc := NewAuthServiceWithInterface(repo, nil)
	_, refresh, err := svc.GenerateTokens(user.ID, user.Email, user.Role, nil, nil)
	require.NoError(t, err)

	newAccess, newRefresh, err := svc.RefreshTokens(refresh)
	require.NoError(t, err)
	assert.NotEmpty(t, newAccess)
	assert.NotEmpty(t, newRefresh)
	repo.AssertExpectations(t)
}
