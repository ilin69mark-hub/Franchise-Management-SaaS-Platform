package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"franchise-saas-backend/internal/middleware"
	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDurableHandlerDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	statements := []string{
		`CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT NOT NULL, password_hash TEXT, role TEXT NOT NULL, status TEXT, auth_version INTEGER NOT NULL DEFAULT 1, tenant_id TEXT, salon_id TEXT, managed_by TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE tenants (id TEXT PRIMARY KEY, name TEXT, status TEXT, plan_id TEXT, timezone TEXT, paid_until DATETIME, grace_period_days INTEGER DEFAULT 7, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE auth_sessions (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, auth_version INTEGER NOT NULL, current_refresh_jti TEXT NOT NULL UNIQUE, chain_started_at DATETIME NOT NULL, chain_expires_at DATETIME NOT NULL, refresh_expires_at DATETIME NOT NULL, last_token_issued_at DATETIME NOT NULL, revoked_at DATETIME, revoke_reason TEXT, created_at DATETIME NOT NULL, updated_at DATETIME NOT NULL)`,
	}
	for _, statement := range statements {
		require.NoError(t, db.Exec(statement).Error)
	}
	return db
}

func newDurableHandlerSession(t *testing.T) (*gorm.DB, *services.AuthService, *models.User, string, string) {
	t.Helper()
	viper.Set("jwt_secret", "durable-handler-logout-secret")
	t.Cleanup(func() { viper.Set("jwt_secret", "") })
	db := newDurableHandlerDB(t)
	tenantID := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO tenants (id,name,status,created_at,updated_at) VALUES (?,?,?,?,?)`, tenantID.String(), "handler", "active", time.Now(), time.Now()).Error)
	user := &models.User{ID: uuid.New(), Email: uuid.NewString() + "@test.com", PasswordHash: "hash", Role: models.RoleFranchisor, Status: "active", AuthVersion: 1, TenantID: &tenantID}
	require.NoError(t, db.Exec(`INSERT INTO users (id,email,password_hash,role,status,auth_version,tenant_id,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?)`, user.ID.String(), user.Email, user.PasswordHash, user.Role, user.Status, user.AuthVersion, tenantID.String(), time.Now(), time.Now()).Error)
	service := services.NewAuthService(db)
	access, refresh, err := service.IssueSession(context.Background(), user)
	require.NoError(t, err)
	return db, service, user, access, refresh
}

func expiredAccessToken(t *testing.T, access string) string {
	t.Helper()
	claims := jwt.MapClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(access, claims)
	require.NoError(t, err)
	claims["exp"] = time.Now().Add(-time.Hour).Unix()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("durable-handler-logout-secret"))
	require.NoError(t, err)
	return token
}

func TestAuthHandlerRefreshAcceptsEmptyBodyWithCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, service, _, _, refresh := newDurableHandlerSession(t)
	r := gin.New()
	r.POST("/auth/refresh", NewAuthHandler(service).RefreshToken)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refresh})
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)
	require.Equal(t, http.StatusOK, response.Code)
}

func TestAuthHandlerRefreshWithoutTokenDoesNotClearCookies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, service, _, _, _ := newDurableHandlerSession(t)
	r := gin.New()
	r.POST("/auth/refresh", NewAuthHandler(service).RefreshToken)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)
	require.Equal(t, http.StatusUnauthorized, response.Code)
	require.Empty(t, response.Result().Cookies())
}

func TestAuthHandlerLogoutRevokesSessionWithExpiredAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, service, user, access, _ := newDurableHandlerSession(t)
	expired := expiredAccessToken(t, access)
	claims := jwt.MapClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(access, claims)
	require.NoError(t, err)
	sid, err := uuid.Parse(claims["sid"].(string))
	require.NoError(t, err)

	r := gin.New()
	r.POST("/auth/logout", middleware.CSRF(), NewAuthHandler(service).Logout)
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: expired})
	req.AddCookie(&http.Cookie{Name: middleware.CSRFCookie, Value: "csrf-token"})
	req.Header.Set(middleware.CSRFHeader, "csrf-token")
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)
	require.Equal(t, http.StatusOK, response.Code)
	_, err = service.ResolveAccessIdentity(context.Background(), user.ID, sid, 1)
	require.Error(t, err)
}
