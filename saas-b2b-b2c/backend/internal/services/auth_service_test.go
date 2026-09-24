package services

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"franchise-saas-backend/internal/cache"
	"franchise-saas-backend/internal/mocks"
	"franchise-saas-backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
	_, err = svc.Authenticate(context.Background(), "blocked@test.com", "password123", "10.9.9.8")

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
	got, err := svc.Authenticate(context.Background(), "active@test.com", "password123", "10.9.9.7")

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

// TestAuthService_GenerateTokens_ExpiresCappedAt24h: F12 — JWT_EXPIRES=720h режется до 24h.
func TestAuthService_GenerateTokens_ExpiresCappedAt24h(t *testing.T) {
	viper.Set("jwt_secret", "cap-secret")
	viper.Set("JWT_EXPIRES", "720h")
	defer viper.Set("jwt_secret", "")
	defer viper.Set("JWT_EXPIRES", "")

	svc := NewAuthServiceWithInterface(nil, nil)
	access, _, err := svc.GenerateTokens(uuid.New(), "cap@test.com", models.RoleDealer, nil, nil)
	require.NoError(t, err)

	parser := jwt.NewParser()
	claims := jwt.MapClaims{}
	_, _, err = parser.ParseUnverified(access, claims)
	require.NoError(t, err)
	exp, ok := claims["exp"].(float64)
	require.True(t, ok, "exp must be present")
	ttl := int64(exp) - time.Now().Unix()
	assert.LessOrEqual(t, ttl, int64((24*time.Hour + 5*time.Minute).Seconds()), "access TTL must be capped at 24h")
	assert.Greater(t, ttl, int64((23 * time.Hour).Seconds()))
}

// TestAuthService_Logout_WithoutRedis_Revokes: F1 — logout без Redis реально отзывает
// (раньше был no-op: повторный RefreshTokens проходил).
func TestAuthService_Logout_WithoutRedis_Revokes(t *testing.T) {
	viper.Set("jwt_secret", "noredis-secret")
	defer viper.Set("jwt_secret", "")

	user := &models.User{ID: uuid.New(), Email: "noredis@test.com", Role: models.RoleDealer}
	repo := mocks.NewMockUserRepository()
	repo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	svc := NewAuthServiceWithInterface(repo, nil)

	_, refresh, err := svc.GenerateTokens(user.ID, user.Email, user.Role, nil, nil)
	require.NoError(t, err)
	require.NoError(t, svc.Logout(refresh))

	_, _, err = svc.RefreshTokens(refresh)
	require.Error(t, err, "logged-out refresh must not rotate without Redis")
	assert.Contains(t, err.Error(), "revoked")
}

// TestAuthService_RefreshTokens_ReuseDenied: F1 — reuse после ротации запрещён без Redis.
func TestAuthService_RefreshTokens_ReuseDenied(t *testing.T) {
	viper.Set("jwt_secret", "reuse-secret")
	defer viper.Set("jwt_secret", "")

	user := &models.User{ID: uuid.New(), Email: "reuse2@test.com", Role: models.RoleDealer}
	repo := mocks.NewMockUserRepository()
	repo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	svc := NewAuthServiceWithInterface(repo, nil)

	_, refresh, err := svc.GenerateTokens(user.ID, user.Email, user.Role, nil, nil)
	require.NoError(t, err)
	_, _, err = svc.RefreshTokens(refresh)
	require.NoError(t, err)
	_, _, err = svc.RefreshTokens(refresh)
	require.Error(t, err, "refresh reuse must be denied")
}

// TestAuthService_RefreshTokens_StaleChainDenied: F12 — цепочке старше 30d отказ.
func TestAuthService_RefreshTokens_StaleChainDenied(t *testing.T) {
	const secret = "stale-secret"
	viper.Set("jwt_secret", secret)
	defer viper.Set("jwt_secret", "")

	svc := NewAuthServiceWithInterface(nil, nil)
	stale := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": uuid.New().String(),
		"jti":     uuid.New().String(),
		"iat":     time.Now().Add(-31 * 24 * time.Hour).Unix(),
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	staleStr, err := stale.SignedString([]byte(secret))
	require.NoError(t, err)

	_, _, err = svc.RefreshTokens(staleStr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chain expired")
}

// TestAuthService_CreateUser_ShortPasswordRejected: F6 — пароль "1" не проходит.
func TestAuthService_CreateUser_ShortPasswordRejected(t *testing.T) {
	svc := NewAuthServiceWithInterface(mocks.NewMockUserRepository(), nil)
	_, err := svc.CreateUser(&models.User{Email: "short@test.com"}, "1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least 12")
}

// TestAuthService_CreateUser_DuplicateEmail409: F6 — дубль маппится в ErrEmailTaken.
func TestAuthService_CreateUser_DuplicateEmail409(t *testing.T) {
	repo := mocks.NewMockUserRepository()
	existing := &models.User{ID: uuid.New(), Email: "taken@test.com"}
	stubHIBPEmpty(t)
	repo.On("GetUserByEmail", mock.Anything, "taken@test.com").Return(existing, nil)
	svc := NewAuthServiceWithInterface(repo, nil)
	_, err := svc.CreateUser(&models.User{Email: "taken@test.com"}, "long-enough-password-1")
	require.ErrorIs(t, err, ErrEmailTaken)
	repo.AssertNotCalled(t, "CreateUser", mock.Anything, mock.Anything)
}

// TestAuthService_CreateUser_RaceDuplicateMapped: F6 — гонка pre-check тоже даёт 409.
func TestAuthService_CreateUser_RaceDuplicateMapped(t *testing.T) {
	repo := mocks.NewMockUserRepository()
	stubHIBPEmpty(t)
	repo.On("GetUserByEmail", mock.Anything, "race@test.com").Return(nil, gorm.ErrRecordNotFound)
	repo.On("CreateUser", mock.Anything, mock.Anything).Return(errors.New(`duplicate key value violates unique constraint "users_email_key"`))
	svc := NewAuthServiceWithInterface(repo, nil)
	_, err := svc.CreateUser(&models.User{Email: "race@test.com"}, "long-enough-password-1")
	require.ErrorIs(t, err, ErrEmailTaken)
}

// TestAuthService_Authenticate_LockoutAfter10Fails: F9 — 11-я попытка с неверным
// паролем даёт ErrTooManyAttempts (раньше можно было перебирать бесконечно).
func TestAuthService_Authenticate_LockoutAfter10Fails(t *testing.T) {
	ctx := context.Background()
	lockIP := "10.10.10.10"
	email := "lockout-" + uuid.New().String() + "@test.com"
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password-12"), bcrypt.MinCost)
	require.NoError(t, err)
	user := &models.User{ID: uuid.New(), Email: email, PasswordHash: string(hash), Status: "active"}

	repo := mocks.NewMockUserRepository()
	repo.On("GetUserByEmail", mock.Anything, email).Return(user, nil)
	repo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	svc := NewAuthServiceWithInterface(repo, nil)

	for i := 0; i < 10; i++ {
		_, err := svc.Authenticate(ctx, email, "wrong-password-12", lockIP)
		require.Error(t, err)
		require.NotContains(t, err.Error(), "too many", "первые 10 попыток — invalid credentials")
	}
	_, err = svc.Authenticate(ctx, email, "wrong-password-12", lockIP)
	require.ErrorIs(t, err, ErrTooManyAttempts)
}

// TestAuthService_Authenticate_SuccessResetsCounter: F9 — успех сбрасывает счётчик.
func TestAuthService_Authenticate_SuccessResetsCounter(t *testing.T) {
	ctx := context.Background()
	resetIP := "10.10.10.11"
	email := "reset-" + uuid.New().String() + "@test.com"
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password-12"), bcrypt.MinCost)
	require.NoError(t, err)
	user := &models.User{ID: uuid.New(), Email: email, PasswordHash: string(hash), Status: "active"}

	repo := mocks.NewMockUserRepository()
	repo.On("GetUserByEmail", mock.Anything, email).Return(user, nil)
	repo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	svc := NewAuthServiceWithInterface(repo, nil)

	for i := 0; i < 3; i++ {
		_, _ = svc.Authenticate(ctx, email, "wrong-password-12", resetIP)
	}
	_, err = svc.Authenticate(ctx, email, "correct-password-12", resetIP)
	require.NoError(t, err)
	// После успеха счётчик сброшен — ещё 9 неудач не должны лочить.
	for i := 0; i < 9; i++ {
		_, err := svc.Authenticate(ctx, email, "wrong-password-12", resetIP)
		require.NotContains(t, err.Error(), "too many")
	}
}

// RE-AUDIT: lockout изолирован по IP — злоумышленник не лочит чужой вход.
func TestAuthService_Authenticate_LockoutIsolatedByIP(t *testing.T) {
	ctx := context.Background()
	attackerIP, victimIP := "10.10.10.12", "10.10.10.13"
	email := "victim-" + uuid.New().String() + "@test.com"
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password-12"), bcrypt.MinCost)
	require.NoError(t, err)
	user := &models.User{ID: uuid.New(), Email: email, PasswordHash: string(hash), Status: "active"}

	repo := mocks.NewMockUserRepository()
	repo.On("GetUserByEmail", mock.Anything, email).Return(user, nil)
	repo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	svc := NewAuthServiceWithInterface(repo, nil)

	for i := 0; i < 10; i++ {
		_, _ = svc.Authenticate(ctx, email, "wrong-password-12", attackerIP)
	}
	_, err = svc.Authenticate(ctx, email, "wrong-password-12", attackerIP)
	require.ErrorIs(t, err, ErrTooManyAttempts)

	// Тот же email с другого IP — не залочен (только invalid credentials).
	_, err = svc.Authenticate(ctx, email, "wrong-password-12", victimIP)
	require.Error(t, err)
	require.NotContains(t, err.Error(), "too many")
}

// RE-AUDIT: заблокированный пользователь не продлевает сессию через refresh.
func TestAuthService_RefreshTokens_BlockedUserDenied(t *testing.T) {
	const secret = "blocked-refresh-secret"
	viper.Set("jwt_secret", secret)
	defer viper.Set("jwt_secret", "")

	user := &models.User{ID: uuid.New(), Email: "blocked-r@test.com", Role: models.RoleDealer, Status: "blocked"}
	repo := mocks.NewMockUserRepository()
	repo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	svc := NewAuthServiceWithInterface(repo, nil)

	_, refresh, err := svc.GenerateTokens(user.ID, user.Email, user.Role, nil, nil)
	require.NoError(t, err)
	_, _, err = svc.RefreshTokens(refresh)
	require.ErrorIs(t, err, ErrUserBlocked)
}

// RE-AUDIT: пароль длиннее 72 байт отклоняется до bcrypt (раньше — 500).
func TestAuthService_CreateUser_PasswordOver72Rejected(t *testing.T) {
	svc := NewAuthServiceWithInterface(mocks.NewMockUserRepository(), nil)
	_, err := svc.CreateUser(&models.User{Email: "long72@test.com"}, longPassword73())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at most 72")
}

func longPassword73() string {
	b := make([]byte, 73)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

// RE-AUDIT: провал создания юзера чистит созданный tenant (без сирот).
func TestAuthService_CreateUserWithTenant_CompensatesOrphan(t *testing.T) {
	userRepo := mocks.NewMockUserRepository()
	tenantRepo := mocks.NewMockTenantRepository()
	svc := NewAuthServiceWithInterface(userRepo, tenantRepo)

	createdTenant := &models.Tenant{ID: uuid.New(), Name: "Acme"}
	tenantRepo.On("CreateTenant", mock.Anything, mock.Anything).Return(createdTenant, nil).Run(func(args mock.Arguments) {
		// эмулируем GORM: входной структуре ID не проставляем — сервис обязан
		// брать ID из возвращённого значения, а не из мутации.
	})
	userRepo.On("GetUserByEmail", mock.Anything, "orphan@test.com").Return(nil, gorm.ErrRecordNotFound)
	stubHIBPEmpty(t)
	userRepo.On("CreateUser", mock.Anything, mock.Anything).Return(errors.New("db down"))
	tenantRepo.On("DeleteTenant", mock.Anything, mock.Anything).Return(nil)

	_, err := svc.CreateUserWithTenant(
		&models.User{Email: "orphan@test.com", FirstName: "A", LastName: "B"},
		"long-enough-password-1", "Acme",
	)
	require.Error(t, err)
	tenantRepo.AssertCalled(t, "DeleteTenant", mock.Anything, createdTenant.ID)
}

// stubHIBPEmpty — подмена HIBP API пустышкой (без совпадений): тесты не ходят в сеть.
func stubHIBPEmpty(t *testing.T) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(""))
	}))
	t.Cleanup(func() {
		srv.Close()
		hibpBaseURL = "https://api.pwnedpasswords.com"
	})
	hibpBaseURL = srv.URL
}

func TestCheckHIBP_ExposedDetected(t *testing.T) {
	pwd := "test-pwned-password-12"
	sum := sha1.Sum([]byte(pwd))
	hexsum := strings.ToUpper(hex.EncodeToString(sum[:]))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.True(t, strings.HasSuffix(r.URL.Path, "/"+hexsum[:5]))
		_, _ = w.Write([]byte(hexsum[5:] + ":42\n"))
	}))
	defer srv.Close()
	prev := hibpBaseURL
	hibpBaseURL = srv.URL
	defer func() { hibpBaseURL = prev }()

	exposed, checked := checkHIBP(pwd)
	require.True(t, checked)
	require.True(t, exposed)
	require.Error(t, rejectPwnedPassword(pwd))
}

func TestCheckHIBP_CleanPasses(t *testing.T) {
	stubHIBPEmpty(t)
	exposed, checked := checkHIBP("definitely-not-pwned-password-12")
	require.True(t, checked)
	require.False(t, exposed)
	require.NoError(t, rejectPwnedPassword("definitely-not-pwned-password-12"))
}

func TestCheckHIBP_FailOpenOnOutage(t *testing.T) {
	prev := hibpBaseURL
	hibpBaseURL = "http://127.0.0.1:1"
	defer func() { hibpBaseURL = prev }()

	exposed, checked := checkHIBP("anything-12-chars")
	require.False(t, checked)
	require.False(t, exposed)
	require.NoError(t, rejectPwnedPassword("anything-12-chars"))
}

// RE-AUDIT: reuse убивает всю цепочку — даже легитимный текущий токен мёртв.
func TestAuthService_RefreshTokens_ReuseKillsChain(t *testing.T) {
	viper.Set("jwt_secret", "killchain-secret")
	defer viper.Set("jwt_secret", "")

	user := &models.User{ID: uuid.New(), Email: "killchain@test.com", Role: models.RoleDealer}
	repo := mocks.NewMockUserRepository()
	repo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	svc := NewAuthServiceWithInterface(repo, nil)

	_, r1, err := svc.GenerateTokens(user.ID, user.Email, user.Role, nil, nil)
	require.NoError(t, err)
	_, r2, err := svc.RefreshTokens(r1)
	require.NoError(t, err)

	_, _, err = svc.RefreshTokens(r1) // reuse украденного/старого
	require.Error(t, err)

	_, _, err = svc.RefreshTokens(r2) // легитимный текущий тоже мёртв
	require.Error(t, err, "kill-chain: текущий токен цепочки отозван")
}

// RE-AUDIT: параллельный double-refresh даёт ровно одну валидную пару.
func TestAuthService_RefreshTokens_ParallelRaceSingleWinner(t *testing.T) {
	viper.Set("jwt_secret", "race-secret")
	defer viper.Set("jwt_secret", "")

	user := &models.User{ID: uuid.New(), Email: "prace@test.com", Role: models.RoleDealer}
	repo := mocks.NewMockUserRepository()
	repo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	svc := NewAuthServiceWithInterface(repo, nil)

	_, r1, err := svc.GenerateTokens(user.ID, user.Email, user.Role, nil, nil)
	require.NoError(t, err)

	const n = 10
	wins := make(chan struct{}, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, _, err := svc.RefreshTokens(r1); err == nil {
				wins <- struct{}{}
			}
		}()
	}
	wg.Wait()
	close(wins)
	require.Len(t, wins, 1, "ровно один победитель гонки")
}

// RE-AUDIT: вход при заблокированной сети запрещён (раньше BlockTenant был надписью).
func TestAuthService_Authenticate_BlockedTenantDenied(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password-12"), bcrypt.MinCost)
	require.NoError(t, err)
	user := &models.User{ID: uuid.New(), Email: "tblocked@test.com", PasswordHash: string(hash), Status: "active", Role: models.RoleDealer, TenantID: &tenantID}

	userRepo := mocks.NewMockUserRepository()
	userRepo.On("GetUserByEmail", mock.Anything, user.Email).Return(user, nil)
	userRepo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	tenantRepo := mocks.NewMockTenantRepository()
	tenantRepo.On("FindByID", mock.Anything, tenantID).Return(&models.Tenant{ID: tenantID, Status: "blocked"}, nil)
	svc := NewAuthServiceWithInterface(userRepo, tenantRepo)

	_, err = svc.Authenticate(ctx, user.Email, "correct-password-12", "10.20.30.40")
	require.ErrorIs(t, err, ErrTenantBlocked)
}

// RE-AUDIT: refresh при заблокированной сети запрещён.
func TestAuthService_RefreshTokens_BlockedTenantDenied(t *testing.T) {
	const secret = "tenant-refresh-secret"
	viper.Set("jwt_secret", secret)
	defer viper.Set("jwt_secret", "")

	tenantID := uuid.New()
	user := &models.User{ID: uuid.New(), Email: "tblocked-r@test.com", Role: models.RoleDealer, Status: "active", TenantID: &tenantID}
	userRepo := mocks.NewMockUserRepository()
	userRepo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	tenantRepo := mocks.NewMockTenantRepository()
	tenantRepo.On("FindByID", mock.Anything, tenantID).Return(&models.Tenant{ID: tenantID, Status: "suspended"}, nil)
	svc := NewAuthServiceWithInterface(userRepo, tenantRepo)

	_, refresh, err := svc.GenerateTokens(user.ID, user.Email, user.Role, &tenantID, nil)
	require.NoError(t, err)
	_, _, err = svc.RefreshTokens(refresh)
	require.ErrorIs(t, err, ErrTenantBlocked)
}

// RE-AUDIT: super_admin без сети — вне tenant-проверки.
func TestAuthService_Authenticate_SuperAdminNoTenantBypass(t *testing.T) {
	ctx := context.Background()
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password-12"), bcrypt.MinCost)
	require.NoError(t, err)
	user := &models.User{ID: uuid.New(), Email: "sa@test.com", PasswordHash: string(hash), Status: "active", Role: models.RoleSuperAdmin}

	userRepo := mocks.NewMockUserRepository()
	userRepo.On("GetUserByEmail", mock.Anything, user.Email).Return(user, nil)
	userRepo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	svc := NewAuthServiceWithInterface(userRepo, nil)

	got, err := svc.Authenticate(ctx, user.Email, "correct-password-12", "10.20.30.41")
	require.NoError(t, err)
	require.Equal(t, user.ID, got.ID)
}

// S4: email нормализуется и уникален регистронезависимо.
func TestAuthService_CreateUser_NormalizesEmail(t *testing.T) {
	stubHIBPEmpty(t)
	repo := mocks.NewMockUserRepository()
	repo.On("GetUserByEmail", mock.Anything, "user@test.com").Return(nil, gorm.ErrRecordNotFound)
	var got *models.User
	repo.On("CreateUser", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		got = args.Get(1).(*models.User)
	}).Return(nil)
	svc := NewAuthServiceWithInterface(repo, nil)

	_, err := svc.CreateUser(&models.User{Email: "  User@Test.COM "}, "long-enough-password-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "user@test.com", got.Email)
}

func TestAuthService_CreateUser_CaseVariantDuplicate409(t *testing.T) {
	stubHIBPEmpty(t)
	repo := mocks.NewMockUserRepository()
	repo.On("GetUserByEmail", mock.Anything, "user@test.com").Return(&models.User{Email: "user@test.com"}, nil)
	svc := NewAuthServiceWithInterface(repo, nil)

	_, err := svc.CreateUser(&models.User{Email: "USER@test.com"}, "long-enough-password-1")
	require.ErrorIs(t, err, ErrEmailTaken)
}

func TestAuthService_Authenticate_CaseInsensitiveLogin(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password-12"), bcrypt.MinCost)
	require.NoError(t, err)
	user := &models.User{ID: uuid.New(), Email: "user@test.com", PasswordHash: string(hash), Status: "active"}
	repo := mocks.NewMockUserRepository()
	repo.On("GetUserByEmail", mock.Anything, "user@test.com").Return(user, nil)
	repo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	svc := NewAuthServiceWithInterface(repo, nil)

	got, err := svc.Authenticate(context.Background(), "USER@Test.COM", "correct-password-12", "10.10.10.20")
	require.NoError(t, err)
	require.Equal(t, user.ID, got.ID)
}

// S1: CAPTCHA требуется после порога неудач; без конфига — пропускается.
func TestVerifyCaptchaToken_SkipsWhenNotConfigured(t *testing.T) {
	t.Setenv("CAPTCHA_SECRET", "")
	t.Setenv("CAPTCHA_VERIFY_URL", "")
	require.NoError(t, verifyCaptchaToken(context.Background(), "10.0.0.1", ""))
}

func TestVerifyCaptchaToken_RequiredWhenConfigured(t *testing.T) {
	t.Setenv("CAPTCHA_SECRET", "secret-key")
	t.Setenv("CAPTCHA_VERIFY_URL", "https://captcha.invalid/verify")
	require.ErrorIs(t, verifyCaptchaToken(context.Background(), "10.0.0.1", ""), ErrCaptchaRequired)
}

func TestVerifyCaptchaToken_SuccessFromProvider(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		require.Equal(t, "secret-key", r.Form.Get("secret"))
		require.Equal(t, "token-42", r.Form.Get("response"))
		require.Equal(t, "10.0.0.9", r.Form.Get("remoteip"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer srv.Close()

	t.Setenv("CAPTCHA_SECRET", "secret-key")
	t.Setenv("CAPTCHA_VERIFY_URL", srv.URL)
	require.NoError(t, verifyCaptchaToken(context.Background(), "10.0.0.9", "token-42"))
}

func TestVerifyCaptchaToken_ProviderFailClosed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":false}`))
	}))
	defer srv.Close()

	t.Setenv("CAPTCHA_SECRET", "secret-key")
	t.Setenv("CAPTCHA_VERIFY_URL", srv.URL)
	require.ErrorIs(t, verifyCaptchaToken(context.Background(), "10.0.0.9", "token-42"), ErrCaptchaRequired)
}

func TestAuthService_Authenticate_CaptchaRequiredAfterThreshold(t *testing.T) {
	stubHIBPEmpty(t)
	repo := mocks.NewMockUserRepository()
	svc := NewAuthServiceWithInterface(repo, nil)
	ctx := context.Background()
	email := "captcha@test.com"
	ip := "10.55.55.55"
	failKey := loginFailKey(ip, email)

	for i := 0; i < captchaThreshold; i++ {
		repo.On("GetUserByEmail", mock.Anything, email).Return(nil, gorm.ErrRecordNotFound).Once()
		_, err := svc.Authenticate(ctx, email, "wrong-password-12", ip)
		require.Error(t, err)
	}
	require.Equal(t, captchaThreshold, cache.PeekLimit(ctx, failKey))

	t.Setenv("CAPTCHA_SECRET", "secret-key")
	t.Setenv("CAPTCHA_VERIFY_URL", "https://captcha.invalid/verify")

	_, err := svc.Authenticate(ctx, email, "correct-password-12", ip)
	require.ErrorIs(t, err, ErrCaptchaRequired, "S1: без токена вход не проверяем вовсе")
}
