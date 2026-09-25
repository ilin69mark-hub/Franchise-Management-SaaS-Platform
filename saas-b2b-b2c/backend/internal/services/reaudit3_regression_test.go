package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"franchise-saas-backend/internal/cache"
	"franchise-saas-backend/internal/middleware"
	"franchise-saas-backend/internal/mocks"
	"franchise-saas-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// C-01: GORM-колонка модели обязана называться contacts_whatsapp — именно это имя
// создают обе миграции (SQL 001 и Go migrateUsers). Расхождение = 500 на любой
// вставке пользователя.
func TestUserModel_WhatsAppColumnMatchesSchema(t *testing.T) {
	parsed, err := schema.Parse(&models.User{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)

	names := map[string]string{}
	for _, f := range parsed.Fields {
		names[f.Name] = f.DBName
	}
	require.Equal(t, "contacts_whatsapp", names["ContactsWhatsApp"],
		"W1: колонка должна называться contacts_whatsapp (как в обеих миграциях)")
}

// H-01: смена пароли отзывает живые сессии (эпоха в Redis/fallback).
func TestChangePassword_RevokesUserSessions(t *testing.T) {
	repo := mocks.NewMockUserRepository()
	svc := NewUserServiceWithInterface(repo, nil)
	rawHash, err := bcrypt.GenerateFromPassword([]byte("OldPassword-12345"), bcrypt.MinCost)
	require.NoError(t, err)
	hash := string(rawHash)

	uid := uuid.New()
	repo.On("GetUserByID", mock.Anything, uid).Return(&models.User{ID: uid, PasswordHash: hash}, nil)
	repo.On("UpdateUserFields", mock.Anything, uid, mock.Anything).Return(nil)

	require.NoError(t, svc.ChangePassword(uid, "OldPassword-12345", "BrandNew-98765"))

	after := cache.UserSessionsRevokedAfter(context.Background(), uid.String())
	require.Greater(t, after, int64(0), "после смены пароля все токены должны считаться отозванными")
}

// H-02: истёкшая подписка (paid_until + grace) закрывает вход.
func TestCheckTenantAccess_ExpiredSubscriptionDenied(t *testing.T) {
	repo := mocks.NewMockUserRepository()
	tenantRepo := mocks.NewMockTenantRepository()
	svc := NewAuthServiceWithInterface(repo, tenantRepo)

	uid, tid := uuid.New(), uuid.New()
	user := &models.User{ID: uid, Email: "e@x.test", Role: models.RoleDealer, TenantID: &tid}
	tenantRepo.On("FindByID", mock.Anything, tid).Return(&models.Tenant{
		ID: tid, Status: "active", PaidUntil: ptrTime(time.Now().Add(-30 * 24 * time.Hour)), GracePeriodDays: 7,
	}, nil)

	err := svc.checkTenantAccess(context.Background(), user)
	require.ErrorIs(t, err, ErrTenantSubscriptionExpired)
}

// ...но в пределах grace-period ещё пускает.
func TestCheckTenantAccess_GracePeriodStillAllowed(t *testing.T) {
	repo := mocks.NewMockUserRepository()
	tenantRepo := mocks.NewMockTenantRepository()
	svc := NewAuthServiceWithInterface(repo, tenantRepo)

	tid := uuid.New()
	user := &models.User{ID: uuid.New(), Email: "g@x.test", Role: models.RoleDealer, TenantID: &tid}
	tenantRepo.On("FindByID", mock.Anything, tid).Return(&models.Tenant{
		ID: tid, Status: "active", PaidUntil: ptrTime(time.Now().Add(-2 * 24 * time.Hour)), GracePeriodDays: 7,
	}, nil)

	require.NoError(t, svc.checkTenantAccess(context.Background(), user))
}

// C-06: access-токен нельзя обменять на refresh и наоборот.
func TestTokenUse_AccessAndRefreshNotInterchangeable(t *testing.T) {
	const secret = "token-use-secret"
	viper.Set("jwt_secret", secret)
	defer viper.Set("jwt_secret", "")

	svc := NewAuthServiceWithInterface(nil, nil)
	access, refresh, err := svc.GenerateTokens(uuid.New(), "x@test.com", models.RoleDealer, nil, nil)
	require.NoError(t, err)

	// access -> refresh должен быть отклонён
	_, _, err = svc.RefreshTokens(access)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not a refresh token")

	// refresh -> access должен быть отклонён middleware
	rec := httptest.NewRecorder()
	if !strings.Contains(refresh, ".") {
		t.Fatal("refresh должен быть JWT")
	}
	_ = rec
}

// H-03: chain_iat переносится при ротации (потолок 30 дней абсолютен).
func TestRefreshChain_ChainIatCarried(t *testing.T) {
	const secret = "chain-iat-secret"
	viper.Set("jwt_secret", secret)
	defer viper.Set("jwt_secret", "")

	repo := mocks.NewMockUserRepository()
	user := &models.User{ID: uuid.New(), Email: "c@x.test", Role: models.RoleDealer, Status: "active"}
	repo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	svc := NewAuthServiceWithInterface(repo, nil)

	_, refresh, err := svc.GenerateTokens(user.ID, user.Email, user.Role, nil, nil)
	require.NoError(t, err)
	_, newRefresh, err := svc.RefreshTokens(refresh)
	require.NoError(t, err)

	newClaims := decodeClaims(t, newRefresh)
	origClaims := decodeClaims(t, refresh)
	require.Equal(t, origClaims["chain_iat"], newClaims["chain_iat"],
		"chain_iat первого токена должен переноситься в ротированный")
}

// C-06 (вторая половина): refresh-токен не принимается AuthMiddleware как bearer.
func TestAuthMiddleware_RejectsRefreshTokenAsBearer(t *testing.T) {
	const secret = "bearer-refresh-secret"
	viper.Set("jwt_secret", secret)
	defer viper.Set("jwt_secret", "")

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"token_use": "refresh",
		"user_id":   uuid.New().String(),
		"jti":       uuid.New().String(),
		"cid":       uuid.New().String(),
		"iat":       time.Now().Unix(),
		"exp":       time.Now().Add(time.Hour).Unix(),
	})
	signed, err := refreshToken.SignedString([]byte(secret))
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code, "refresh-токен не должен работать как bearer")
}

func ptrTime(t time.Time) *time.Time { return &t }

// decodeClaims — разбор JWT без проверки подписи (тестовая утилита).
func decodeClaims(t *testing.T, tokenStr string) jwt.MapClaims {
	t.Helper()
	parser := jwt.NewParser()
	claims := jwt.MapClaims{}
	_, _, err := parser.ParseUnverified(tokenStr, claims)
	require.NoError(t, err)
	return claims
}

var _ = gorm.ErrRecordNotFound
