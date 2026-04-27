package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"franchise-saas-backend/internal/cache"
	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	tenantRepo *repository.TenantRepository
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		userRepo:   repository.NewUserRepository(db),
		tenantRepo: repository.NewTenantRepository(db),
	}
}

// CreateUserWithTenant - Создает сеть (Tenant) и Владельца (User)
func (s *AuthService) CreateUserWithTenant(user *models.User, password string, companyName string) (*models.User, error) {
	ctx := context.Background()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = string(hashedPassword)

	tenant := &models.Tenant{Name: companyName, Status: "active"}
	if _, err := s.tenantRepo.CreateTenant(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	user.TenantID = &tenant.ID
	user.Role = models.RoleFranchisor
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

// CreateUser - Создание пользователя без создания сети
func (s *AuthService) CreateUser(user *models.User, password string) (*models.User, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = string(hashedPass)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := s.userRepo.CreateUser(context.Background(), user); err != nil {
		return nil, err
	}
	return user, nil
}

// Authenticate - Проверка логина/пароля
func (s *AuthService) Authenticate(email, password string) (*models.User, error) {
	user, err := s.userRepo.GetUserByEmail(context.Background(), email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Проверка пароля
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, errors.New("invalid credentials")
	}

	// ВАЖНОЕ ИЗМЕНЕНИЕ:
	// После успешной проверки пароля, мы заново запрашиваем пользователя по ID.
	// Это нужно, чтобы гарантированно получить актуальное поле SalonID,
	// которое могло измениться (назначение менеджера) после создания пользователя.
	freshUser, err := s.userRepo.GetUserByID(context.Background(), user.ID)
	if err != nil {
		return nil, errors.New("failed to fetch user details")
	}

	return freshUser, nil
}

// GenerateTokens - Генерация токенов
func (s *AuthService) GenerateTokens(userID uuid.UUID, email string, role models.Role, tenantID *uuid.UUID, salonID *uuid.UUID) (string, string, error) {
	secret := viper.GetString("jwt_secret")
	if secret == "" {
		secret = "CHANGE_ME_JWT_SECRET_change_in_production"
	}

	tidStr := ""
	if tenantID != nil {
		tidStr = tenantID.String()
	}

	sidStr := ""
	if salonID != nil {
		sidStr = salonID.String()
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   userID.String(),
		"email":     email,
		"role":      role,
		"tenant_id": tidStr,
		"salon_id":  sidStr,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	})

	jti := uuid.New().String()
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"jti":     jti,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	accessStr, _ := accessToken.SignedString([]byte(secret))
	refreshStr, _ := refreshToken.SignedString([]byte(secret))

	return accessStr, refreshStr, nil
}

// RefreshTokens - Обновление токенов с ротацией
func (s *AuthService) RefreshTokens(oldRefreshToken string) (string, string, error) {
	secret := viper.GetString("jwt_secret")

	token, err := jwt.Parse(oldRefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("invalid token claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return "", "", errors.New("user_id not found in token")
	}

	tokenID, _ := claims["jti"].(string)
	if tokenID != "" {
		cache.RevokeRefreshToken(context.Background(), tokenID)
	}

	uid, _ := uuid.Parse(userIDStr)
	user, err := s.userRepo.GetUserByID(context.Background(), uid)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	return s.GenerateTokens(user.ID, user.Email, user.Role, user.TenantID, user.SalonID)
}

// GetUserByID - Получение пользователя по ID
func (s *AuthService) GetUserByID(id uuid.UUID) (*models.User, error) {
	return s.userRepo.GetUserByID(context.Background(), id)
}
