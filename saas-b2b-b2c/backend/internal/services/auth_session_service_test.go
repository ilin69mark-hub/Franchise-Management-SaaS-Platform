package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"franchise-saas-backend/internal/cache"
	"franchise-saas-backend/internal/middleware"
	"franchise-saas-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newDurableAuthSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := newTestSQLiteDB(t)
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

func insertDurableAuthUser(t *testing.T, db *gorm.DB, tenantID *uuid.UUID) *models.User {
	t.Helper()
	user := &models.User{
		ID:           uuid.New(),
		Email:        "durable-" + uuid.NewString() + "@test.com",
		PasswordHash: "hash",
		Role:         models.RoleFranchisor,
		Status:       "active",
		AuthVersion:  1,
		TenantID:     tenantID,
	}
	require.NoError(t, db.Exec(`INSERT INTO users (id,email,password_hash,role,status,auth_version,tenant_id,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?)`, user.ID.String(), user.Email, user.PasswordHash, user.Role, user.Status, user.AuthVersion, func() any {
		if tenantID == nil {
			return nil
		}
		return tenantID.String()
	}(), time.Now(), time.Now()).Error)
	return user
}

func durableTokenClaims(t *testing.T, tokenString string) jwt.MapClaims {
	t.Helper()
	claims := jwt.MapClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(tokenString, claims)
	require.NoError(t, err)
	return claims
}

func TestAuthServiceDurableSessionLogoutAndRedisLoss(t *testing.T) {
	viper.Set("jwt_secret", "durable-session-secret")
	defer viper.Set("jwt_secret", "")
	db := newDurableAuthSQLiteDB(t)
	tenantID := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO tenants (id,name,status,created_at,updated_at) VALUES (?,?,?,?,?)`, tenantID.String(), "durable", "active", time.Now(), time.Now()).Error)
	user := insertDurableAuthUser(t, db, &tenantID)
	svc := NewAuthService(db)

	access, refresh, err := svc.IssueSession(context.Background(), user)
	require.NoError(t, err)
	claims := durableTokenClaims(t, access)
	sid, err := uuid.Parse(claims["sid"].(string))
	require.NoError(t, err)
	_, err = svc.ResolveAccessIdentity(context.Background(), user.ID, sid, 1)
	require.NoError(t, err)
	rotatedAccess, rotatedRefresh, err := svc.RefreshTokens(refresh)
	require.NoError(t, err)
	require.NotEmpty(t, rotatedAccess)
	require.NotEmpty(t, rotatedRefresh)
	require.Equal(t, durableTokenClaims(t, refresh)["chain_iat"], durableTokenClaims(t, rotatedRefresh)["chain_iat"])
	_, _, err = svc.RefreshTokens(refresh)
	require.ErrorIs(t, err, ErrRefreshReuse)
	_, err = svc.ResolveAccessIdentity(context.Background(), user.ID, sid, 1)
	require.Error(t, err)

	oldRedis := cache.Client
	cache.Client = nil
	defer func() { cache.Client = oldRedis }()
	_, err = svc.ResolveAccessIdentity(context.Background(), user.ID, sid, 1)
	require.Error(t, err)
	_, _, err = svc.RefreshTokens(rotatedRefresh)
	require.Error(t, err)
}

func TestAuthServiceLogoutTokenAcceptsExpiredAccess(t *testing.T) {
	viper.Set("jwt_secret", "durable-expired-access-logout-secret")
	defer viper.Set("jwt_secret", "")
	db := newDurableAuthSQLiteDB(t)
	tenantID := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO tenants (id,name,status,created_at,updated_at) VALUES (?,?,?,?,?)`, tenantID.String(), "logout-expiry", "active", time.Now(), time.Now()).Error)
	user := insertDurableAuthUser(t, db, &tenantID)
	svc := NewAuthService(db)
	access, _, err := svc.IssueSession(context.Background(), user)
	require.NoError(t, err)
	claims := durableTokenClaims(t, access)
	claims["exp"] = time.Now().Add(-time.Hour).Unix()
	expiredAccess, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("durable-expired-access-logout-secret"))
	require.NoError(t, err)

	recognized, err := svc.LogoutToken(context.Background(), expiredAccess)
	require.NoError(t, err)
	require.True(t, recognized)
	_, err = svc.ResolveAccessIdentity(context.Background(), user.ID, uuid.MustParse(claims["sid"].(string)), 1)
	require.Error(t, err)
}

func TestAuthServiceDurableSessionRevokeAllAndDeviceIsolation(t *testing.T) {
	viper.Set("jwt_secret", "durable-revoke-secret")
	defer viper.Set("jwt_secret", "")
	db := newDurableAuthSQLiteDB(t)
	tenantID := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO tenants (id,name,status,created_at,updated_at) VALUES (?,?,?,?,?)`, tenantID.String(), "durable", "active", time.Now(), time.Now()).Error)
	user := insertDurableAuthUser(t, db, &tenantID)
	svc := NewAuthService(db)

	accessA, _, err := svc.IssueSession(context.Background(), user)
	require.NoError(t, err)
	_, _, err = svc.IssueSession(context.Background(), user)
	require.NoError(t, err)
	claimsA := durableTokenClaims(t, accessA)
	sidA, err := uuid.Parse(claimsA["sid"].(string))
	require.NoError(t, err)
	var sessionB models.AuthSession
	require.NoError(t, db.Where("user_id = ? AND id <> ?", user.ID, sidA).First(&sessionB).Error)
	_, err = svc.ResolveAccessIdentity(context.Background(), user.ID, sidA, 1)
	require.NoError(t, err)
	_, err = svc.ResolveAccessIdentity(context.Background(), user.ID, sessionB.ID, 1)
	require.NoError(t, err)
	require.NoError(t, svc.LogoutSession(context.Background(), user.ID, sidA))
	require.NoError(t, svc.LogoutSession(context.Background(), user.ID, sidA))
	_, err = svc.ResolveAccessIdentity(context.Background(), user.ID, sidA, 1)
	require.Error(t, err)
	_, err = svc.ResolveAccessIdentity(context.Background(), user.ID, sessionB.ID, 1)
	require.NoError(t, err)

	require.NoError(t, svc.RevokeAllUserSessions(context.Background(), user.ID, "security_change"))
	_, err = svc.ResolveAccessIdentity(context.Background(), user.ID, sessionB.ID, 1)
	require.Error(t, err)
	user, err = svc.GetUserByID(user.ID)
	require.NoError(t, err)
	_, _, err = svc.IssueSession(context.Background(), user)
	require.NoError(t, err)
}

func TestAuthServiceDurableSessionRejectsInvalidAuthVersion(t *testing.T) {
	viper.Set("jwt_secret", "durable-auth-version-secret")
	defer viper.Set("jwt_secret", "")
	db := newDurableAuthSQLiteDB(t)
	tenantID := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO tenants (id,name,status,created_at,updated_at) VALUES (?,?,?,?,?)`, tenantID.String(), "auth-version", "active", time.Now(), time.Now()).Error)
	user := insertDurableAuthUser(t, db, &tenantID)
	require.NoError(t, db.Exec(`UPDATE users SET auth_version = 0 WHERE id = ?`, user.ID).Error)

	_, _, err := NewAuthService(db).IssueSession(context.Background(), user)
	require.ErrorIs(t, err, ErrAuthSessionRevoked)
}

func TestAuthServiceDurableRefreshRejectsExpiredSession(t *testing.T) {
	viper.Set("jwt_secret", "durable-expiry-secret")
	defer viper.Set("jwt_secret", "")
	db := newDurableAuthSQLiteDB(t)
	tenantID := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO tenants (id,name,status,created_at,updated_at) VALUES (?,?,?,?,?)`, tenantID.String(), "expiry", "active", time.Now(), time.Now()).Error)
	user := insertDurableAuthUser(t, db, &tenantID)
	svc := NewAuthService(db)
	_, refresh, err := svc.IssueSession(context.Background(), user)
	require.NoError(t, err)
	require.NoError(t, db.Exec(`UPDATE auth_sessions SET refresh_expires_at = ? WHERE user_id = ?`, time.Now().Add(-time.Minute), user.ID).Error)

	_, _, err = svc.RefreshTokens(refresh)
	require.ErrorIs(t, err, ErrAuthSessionRevoked)
}

func TestAdminBlockTenantRevokesDurableSessions(t *testing.T) {
	viper.Set("jwt_secret", "durable-tenant-block-secret")
	defer viper.Set("jwt_secret", "")
	db := newDurableAuthSQLiteDB(t)
	tenantID := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO tenants (id,name,status,created_at,updated_at) VALUES (?,?,?,?,?)`, tenantID.String(), "blocked", "active", time.Now(), time.Now()).Error)
	user := insertDurableAuthUser(t, db, &tenantID)
	authSvc := NewAuthService(db)
	access, _, err := authSvc.IssueSession(context.Background(), user)
	require.NoError(t, err)
	sid, err := uuid.Parse(durableTokenClaims(t, access)["sid"].(string))
	require.NoError(t, err)
	adminSvc := NewAdminService(db)
	require.NoError(t, adminSvc.BlockTenant(tenantID))
	_, err = authSvc.ResolveAccessIdentity(context.Background(), user.ID, sid, 1)
	require.Error(t, err)
	require.NoError(t, adminSvc.UnblockTenant(tenantID))
	_, err = authSvc.ResolveAccessIdentity(context.Background(), user.ID, sid, 1)
	require.Error(t, err)
	user, err = authSvc.GetUserByID(user.ID)
	require.NoError(t, err)
	_, _, err = authSvc.IssueSession(context.Background(), user)
	require.NoError(t, err)
}

func TestUserServiceSecurityChangeRevokesDurableSessions(t *testing.T) {
	viper.Set("jwt_secret", "durable-security-change-secret")
	defer viper.Set("jwt_secret", "")
	db := newDurableAuthSQLiteDB(t)
	tenantID := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO tenants (id,name,status,created_at,updated_at) VALUES (?,?,?,?,?)`, tenantID.String(), "security", "active", time.Now(), time.Now()).Error)
	user := insertDurableAuthUser(t, db, &tenantID)
	user.Role = models.RoleDealer
	require.NoError(t, db.Exec(`UPDATE users SET role = ? WHERE id = ?`, user.Role, user.ID).Error)
	authSvc := NewAuthService(db)
	_, _, err := authSvc.IssueSession(context.Background(), user)
	require.NoError(t, err)
	var session models.AuthSession
	require.NoError(t, db.Where("user_id = ?", user.ID).First(&session).Error)

	userSvc := NewUserService(db)
	_, err = userSvc.UpdateEmployee(user.ID, user.ID, tenantID, models.UpdateEmployeeRequest{Role: models.RoleDealerManager}, string(models.RoleDealer))
	require.NoError(t, err)
	var version int64
	require.NoError(t, db.Raw(`SELECT auth_version FROM users WHERE id = ?`, user.ID).Scan(&version).Error)
	require.Equal(t, int64(2), version)
	_, err = authSvc.ResolveAccessIdentity(context.Background(), user.ID, session.ID, 1)
	require.Error(t, err)
}

func TestAuthMiddlewareDurableSession(t *testing.T) {
	viper.Set("jwt_secret", "durable-middleware-secret")
	defer viper.Set("jwt_secret", "")
	db := newDurableAuthSQLiteDB(t)
	tenantID := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO tenants (id,name,status,created_at,updated_at) VALUES (?,?,?,?,?)`, tenantID.String(), "middleware", "active", time.Now(), time.Now()).Error)
	user := insertDurableAuthUser(t, db, &tenantID)
	svc := NewAuthService(db)
	access, _, err := svc.IssueSession(context.Background(), user)
	require.NoError(t, err)
	claims := durableTokenClaims(t, access)
	sid, err := uuid.Parse(claims["sid"].(string))
	require.NoError(t, err)

	middleware.SetIdentityResolver(func(ctx context.Context, userID, sessionID string, authVersion int64) (string, string, bool) {
		uid, parseErr := uuid.Parse(userID)
		if parseErr != nil {
			return "", "", false
		}
		parsedSessionID, parseErr := uuid.Parse(sessionID)
		if parseErr != nil {
			return "", "", false
		}
		identity, resolveErr := svc.ResolveAccessIdentity(ctx, uid, parsedSessionID, authVersion)
		if resolveErr != nil {
			return "", "", false
		}
		tenant := ""
		if identity.TenantID != nil {
			tenant = identity.TenantID.String()
		}
		return string(identity.Role), tenant, true
	})
	defer middleware.SetIdentityResolver(nil)

	router := gin.New()
	router.Use(middleware.AuthMiddleware())
	router.GET("/private", func(c *gin.Context) { c.Status(http.StatusOK) })
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.Header.Set("Authorization", "Bearer "+access)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)

	require.NoError(t, svc.LogoutSession(context.Background(), user.ID, sid))
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusUnauthorized, response.Code)
}
