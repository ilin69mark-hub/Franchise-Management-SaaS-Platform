package services

import (
	"context"
	"net/http"
	"net/http/httptest"
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
)

// C-1: refresh не должен выдавать новую сессию после отзыва (смена пароля/роли).
func TestRefreshTokens_DeniedAfterSessionRevoked(t *testing.T) {
	const secret = "refresh-revoked-secret"
	viper.Set("jwt_secret", secret)
	defer viper.Set("jwt_secret", "")

	repo := mocks.NewMockUserRepository()
	user := &models.User{ID: uuid.New(), Email: "rev@x.test", Role: models.RoleFranchisor, Status: "active"}
	repo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	svc := NewAuthServiceWithInterface(repo, nil)

	_, refresh, err := svc.GenerateTokens(user.ID, user.Email, user.Role, nil, nil)
	require.NoError(t, err)

	// Эмулируем смену пароля/роли: эпоха отзыва ставится "сейчас".
	cache.RevokeUserSessions(context.Background(), user.ID.String())

	_, _, err = svc.RefreshTokens(refresh)
	require.Error(t, err, "REAUDIT-4: refresh обязан уважать отзыв сессий")
	require.Contains(t, err.Error(), "revoked")
}

// C-1b: валидная (не отозванная) сессия по-прежнему обновляется.
func TestRefreshTokens_AllowsLiveSession(t *testing.T) {
	const secret = "refresh-live-secret"
	viper.Set("jwt_secret", secret)
	defer viper.Set("jwt_secret", "")

	repo := mocks.NewMockUserRepository()
	user := &models.User{ID: uuid.New(), Email: "live@x.test", Role: models.RoleFranchisor, Status: "active"}
	repo.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	svc := NewAuthServiceWithInterface(repo, nil)

	_, refresh, err := svc.GenerateTokens(user.ID, user.Email, user.Role, nil, nil)
	require.NoError(t, err)
	_, newRefresh, err := svc.RefreshTokens(refresh)
	require.NoError(t, err)
	require.NotEmpty(t, newRefresh)
}

// C-2: отзыв не должен "исчезать" при недоступном Redis (fail-open).
func TestRevocationSurvivesRedisOutage(t *testing.T) {
	uid := "user-" + uuid.New().String()
	cache.RevokeUserSessions(context.Background(), uid)

	// Симулируем деградацию Redis: временно отключаем клиент.
	orig := cache.Client
	cache.Client = nil
	defer func() { cache.Client = orig }()

	got := cache.UserSessionsRevokedAfter(context.Background(), uid)
	require.Greater(t, got, int64(0), "REAUDIT-4: локальный маркер обязан пережить потерю Redis")
}

// H-2: blocked-пользователь не проходит middleware (роль/tenant из БД).
func TestAuthMiddleware_BlockedUserRejected(t *testing.T) {
	const secret = "blocked-user-secret"
	viper.Set("jwt_secret", secret)
	defer viper.Set("jwt_secret", "")

	// Резолвер БД говорит "пользователя нет/заблокирован".
	middleware.SetIdentityResolver(func(ctx context.Context, userID, sessionID string, authVersion int64) (string, string, bool) {
		return "", "", false
	})
	defer middleware.SetIdentityResolver(nil)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"token_use": "access",
		"user_id":   uuid.New().String(),
		"role":      "franchiser",
		"tenant_id": uuid.New().String(),
		"sid":       uuid.New().String(),
		"av":        int64(1),
		"jti":       uuid.New().String(),
		"iat":       time.Now().Unix(),
		"exp":       time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code, "заблокированный/удалённый пользователь не проходит")
}

// H-10: assignee не может редактировать план цели.
func TestGoalService_AssigneeCannotEditPlan(t *testing.T) {
	repo := new(MockGoalRepo)
	svc := NewGoalService(repo)

	assignerID, assigneeID := uuid.New().String(), uuid.New().String()
	goal := &models.Goal{
		ID: uuid.New(), AssignerID: uuid.MustParse(assignerID), AssigneeID: uuid.MustParse(assigneeID),
		Role: string(models.RoleDealer), SalesPlan: 1000, Period: models.PeriodMonth,
	}
	repo.On("GetByID", mock.Anything, goal.ID.String()).Return(goal, nil)

	zero := 0.0
	updated, err := svc.UpdateGoal(context.Background(), goal.ID.String(), UpdateGoalDTO{SalesPlan: &zero}, assigneeID, "", string(models.RoleDealer))
	require.Error(t, err, "REAUDIT-4: assignee не управляет собственным планом")
	require.Nil(t, updated)
}

// REAUDIT-4: квота max_users соблюдается (tenant с лимитом 1 не может создать второго).
func TestEnforceUserQuota_BlocksOverLimit(t *testing.T) {
	db := newTestSQLiteDB(t)
	require.NoError(t, db.Exec(`CREATE TABLE tenants (
		id TEXT PRIMARY KEY, name TEXT, status TEXT, plan_id TEXT,
		max_users INTEGER DEFAULT 10, paid_until DATETIME, grace_period_days INTEGER DEFAULT 7,
		deleted_at DATETIME, created_at DATETIME, updated_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE plans (id TEXT PRIMARY KEY, name TEXT, price REAL, max_users INTEGER, max_salons INTEGER)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY, tenant_id TEXT, role TEXT, email TEXT)`).Error)

	tenantID, planID := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO plans (id,name,max_users,max_salons) VALUES (?, 'P', 1, 1)`, planID.String()).Error)
	require.NoError(t, db.Exec(`INSERT INTO tenants (id,name,status,plan_id,max_users) VALUES (?, 'T', 'active', ?, 10)`, tenantID.String(), planID.String()).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id,tenant_id,role) VALUES (?, ?, 'dealer')`, uuid.New().String(), tenantID.String()).Error)

	err := enforceUserQuota(db, tenantID)
	require.Error(t, err, "REAUDIT-4: лимит плана (1) обязан блокировать превышение")
}

// Без плана и без max_users — лимита нет.
func TestEnforceUserQuota_NoLimitConfigured(t *testing.T) {
	db := newTestSQLiteDB(t)
	require.NoError(t, db.Exec(`CREATE TABLE tenants (id TEXT PRIMARY KEY, name TEXT, status TEXT, plan_id TEXT, max_users INTEGER, paid_until DATETIME, grace_period_days INTEGER, deleted_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY, tenant_id TEXT)`).Error)
	tenantID := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO tenants (id,name,status) VALUES (?, 'T2', 'active')`, tenantID.String()).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id,tenant_id) VALUES (?, ?)`, uuid.New().String(), tenantID.String()).Error)
	require.NoError(t, enforceUserQuota(db, tenantID))
}

// REAUDIT-4: managed_by обязан быть в той же сети даже для super_admin.
func TestValidateManagedBy_RejectsCrossTenantAndCycles(t *testing.T) {
	repo := mocks.NewMockUserRepository()
	svc := NewUserServiceWithInterface(repo, nil)

	tenantA, tenantB := uuid.New(), uuid.New()
	boss, peer, sub, outsider := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	repo.On("GetUserByID", mock.Anything, boss).Return(&models.User{ID: boss, TenantID: &tenantA, Role: models.RoleFranchisor}, nil)
	repo.On("GetUserByID", mock.Anything, peer).Return(&models.User{ID: peer, TenantID: &tenantA, Role: models.RoleDealer, ManagedBy: &boss}, nil)
	repo.On("GetUserByID", mock.Anything, sub).Return(&models.User{ID: sub, TenantID: &tenantA, Role: models.RoleDealer, ManagedBy: &peer}, nil)
	repo.On("GetUserByID", mock.Anything, outsider).Return(&models.User{ID: outsider, TenantID: &tenantB, Role: models.RoleDealer}, nil)

	// норма: подчинённого можно назначить руководителем другого подчинённого
	require.NoError(t, svc.validateManagedBy(sub, boss))
	// запрет: руководитель из другой сети
	require.Error(t, svc.validateManagedBy(boss, outsider), "cross-tenant managed_by запрещён")
	// запрет: сам себе
	require.Error(t, svc.validateManagedBy(peer, peer), "self managed_by запрещён")
	// запрет: цикл (подчинённый становится руководителем начальника)
	require.Error(t, svc.validateManagedBy(peer, sub), "цикл иерархии запрещён")
}
