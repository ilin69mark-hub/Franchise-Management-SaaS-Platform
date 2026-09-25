package repository

import (
	"context"
	"testing"
	"time"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newAuthSessionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	require.NoError(t, db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			auth_version INTEGER NOT NULL DEFAULT 1,
			deleted_at DATETIME,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE auth_sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			auth_version INTEGER NOT NULL DEFAULT 1 CHECK (auth_version > 0),
			current_refresh_jti TEXT NOT NULL UNIQUE,
			chain_started_at DATETIME NOT NULL,
			chain_expires_at DATETIME NOT NULL,
			refresh_expires_at DATETIME NOT NULL,
			last_token_issued_at DATETIME NOT NULL,
			revoked_at DATETIME,
			revoke_reason TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
			CHECK (refresh_expires_at <= chain_expires_at)
		);
		CREATE INDEX idx_auth_sessions_user_active ON auth_sessions(user_id) WHERE revoked_at IS NULL;
		CREATE INDEX idx_auth_sessions_refresh_expiry_active ON auth_sessions(refresh_expires_at) WHERE revoked_at IS NULL;
		CREATE INDEX idx_auth_sessions_revoked ON auth_sessions(revoked_at) WHERE revoked_at IS NOT NULL;
	`).Error)
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}

func createAuthSessionTestUser(t *testing.T, db *gorm.DB, userID uuid.UUID) {
	t.Helper()
	require.NoError(t, db.Exec("INSERT INTO users (id, auth_version) VALUES (?, ?)", userID.String(), int64(1)).Error)
}

func authSessionTestFixture(userID uuid.UUID, now time.Time) *models.AuthSession {
	return &models.AuthSession{
		ID:                uuid.New(),
		UserID:            userID,
		AuthVersion:       1,
		CurrentRefreshJTI: uuid.New(),
		ChainStartedAt:    now,
		ChainExpiresAt:    now.Add(24 * time.Hour),
		RefreshExpiresAt:  now.Add(time.Hour),
		LastTokenIssuedAt: now,
	}
}

func TestAuthSessionRepositoryCreateAndGetSQLite(t *testing.T) {
	db := newAuthSessionTestDB(t)
	repo := NewAuthSessionRepository(db)
	ctx := context.Background()
	userID := uuid.New()
	createAuthSessionTestUser(t, db, userID)
	now := time.Now().UTC().Truncate(time.Second)
	session := authSessionTestFixture(userID, now)
	session.ID = uuid.Nil

	require.NoError(t, repo.Create(ctx, session))
	assert.NotEqual(t, uuid.Nil, session.ID)

	got, err := repo.GetByID(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, got.ID)
	assert.Equal(t, userID, got.UserID)
	assert.Equal(t, session.CurrentRefreshJTI, got.CurrentRefreshJTI)
	assert.WithinDuration(t, now, got.ChainStartedAt, time.Second)

	got, err = repo.GetByIDForUser(ctx, session.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, got.ID)

	_, err = repo.GetByIDForUser(ctx, session.ID, uuid.New())
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestAuthSessionRepositoryRotateRefreshJTISQLite(t *testing.T) {
	db := newAuthSessionTestDB(t)
	repo := NewAuthSessionRepository(db)
	ctx := context.Background()
	userID := uuid.New()
	createAuthSessionTestUser(t, db, userID)
	session := authSessionTestFixture(userID, time.Now().UTC().Truncate(time.Second))
	require.NoError(t, repo.Create(ctx, session))

	oldJTI := session.CurrentRefreshJTI
	newJTI := uuid.New()
	rotated, err := repo.RotateRefreshJTI(ctx, session.ID, oldJTI, newJTI)
	require.NoError(t, err)
	assert.True(t, rotated)

	rotated, err = repo.RotateRefreshJTI(ctx, session.ID, oldJTI, uuid.New())
	require.NoError(t, err)
	assert.False(t, rotated)

	got, err := repo.GetByID(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, newJTI, got.CurrentRefreshJTI)
	assert.WithinDuration(t, time.Now(), got.LastTokenIssuedAt, time.Second)
}

func TestAuthSessionRepositoryRevokeSQLite(t *testing.T) {
	db := newAuthSessionTestDB(t)
	repo := NewAuthSessionRepository(db)
	ctx := context.Background()
	userID := uuid.New()
	createAuthSessionTestUser(t, db, userID)
	session := authSessionTestFixture(userID, time.Now().UTC().Truncate(time.Second))
	require.NoError(t, repo.Create(ctx, session))

	require.NoError(t, repo.Revoke(ctx, session.ID, uuid.New(), "logout"))

	require.NoError(t, repo.Revoke(ctx, session.ID, userID, "logout"))
	got, err := repo.GetByID(ctx, session.ID)
	require.NoError(t, err)
	require.NotNil(t, got.RevokedAt)
	require.NotNil(t, got.RevokeReason)
	assert.Equal(t, "logout", *got.RevokeReason)
	firstRevokedAt := *got.RevokedAt

	require.NoError(t, repo.Revoke(ctx, session.ID, userID, "second"))
	got, err = repo.GetByID(ctx, session.ID)
	require.NoError(t, err)
	require.NotNil(t, got.RevokedAt)
	assert.Equal(t, firstRevokedAt, *got.RevokedAt)
	assert.Equal(t, "logout", *got.RevokeReason)
}

func TestAuthSessionRepositoryRevokeAllForUserSQLite(t *testing.T) {
	db := newAuthSessionTestDB(t)
	repo := NewAuthSessionRepository(db)
	ctx := context.Background()
	userID := uuid.New()
	otherUserID := uuid.New()
	createAuthSessionTestUser(t, db, userID)
	createAuthSessionTestUser(t, db, otherUserID)
	now := time.Now().UTC().Truncate(time.Second)
	first := authSessionTestFixture(userID, now)
	second := authSessionTestFixture(userID, now)
	other := authSessionTestFixture(otherUserID, now)
	require.NoError(t, repo.Create(ctx, first))
	require.NoError(t, repo.Create(ctx, second))
	require.NoError(t, repo.Create(ctx, other))

	require.NoError(t, repo.RevokeAllForUser(ctx, userID, "password_changed"))
	require.NoError(t, repo.RevokeAllForUser(ctx, userID, "password_changed_again"))
	assert.ErrorIs(t, repo.RevokeAllForUser(ctx, uuid.New(), "missing"), gorm.ErrRecordNotFound)

	firstGot, err := repo.GetByID(ctx, first.ID)
	require.NoError(t, err)
	assert.NotNil(t, firstGot.RevokedAt)
	secondGot, err := repo.GetByID(ctx, second.ID)
	require.NoError(t, err)
	assert.NotNil(t, secondGot.RevokedAt)
	otherGot, err := repo.GetByID(ctx, other.ID)
	require.NoError(t, err)
	assert.Nil(t, otherGot.RevokedAt)
}

func TestAuthSessionRepositoryIncrementAuthVersionSQLite(t *testing.T) {
	db := newAuthSessionTestDB(t)
	repo := NewAuthSessionRepository(db)
	ctx := context.Background()
	userID := uuid.New()
	createAuthSessionTestUser(t, db, userID)

	require.NoError(t, repo.IncrementAuthVersion(ctx, userID))
	var version int64
	require.NoError(t, db.Raw("SELECT auth_version FROM users WHERE id = ?", userID.String()).Scan(&version).Error)
	assert.Equal(t, int64(2), version)
	assert.ErrorIs(t, repo.IncrementAuthVersion(ctx, uuid.New()), gorm.ErrRecordNotFound)

	now := time.Now().UTC().Truncate(time.Second)
	first := authSessionTestFixture(userID, now)
	second := authSessionTestFixture(userID, now)
	require.NoError(t, repo.Create(ctx, first))
	require.NoError(t, repo.Create(ctx, second))
	require.NoError(t, repo.IncrementAuthVersionAndRevoke(ctx, userID, "role_changed"))
	require.NoError(t, db.Raw("SELECT auth_version FROM users WHERE id = ?", userID.String()).Scan(&version).Error)
	assert.Equal(t, int64(3), version)
	firstGot, err := repo.GetByID(ctx, first.ID)
	require.NoError(t, err)
	assert.NotNil(t, firstGot.RevokedAt)
	secondGot, err := repo.GetByID(ctx, second.ID)
	require.NoError(t, err)
	assert.NotNil(t, secondGot.RevokedAt)
}

func TestAuthSessionRepositoryCleanupSQLite(t *testing.T) {
	db := newAuthSessionTestDB(t)
	repo := NewAuthSessionRepository(db)
	ctx := context.Background()
	userID := uuid.New()
	createAuthSessionTestUser(t, db, userID)
	now := time.Now().UTC().Truncate(time.Second)

	expired := make([]*models.AuthSession, 3)
	for i := range expired {
		expired[i] = authSessionTestFixture(userID, now.Add(-2*time.Hour))
		expired[i].RefreshExpiresAt = now.Add(-time.Hour)
		expired[i].ChainExpiresAt = now.Add(time.Hour)
		require.NoError(t, repo.Create(ctx, expired[i]))
	}
	active := authSessionTestFixture(userID, now)
	require.NoError(t, repo.Create(ctx, active))
	revoked := authSessionTestFixture(userID, now)
	revoked.RevokedAt = timePtr(now)
	revoked.RevokeReason = stringPtr("logout")
	require.NoError(t, repo.Create(ctx, revoked))

	deleted, err := repo.Cleanup(ctx, now, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(2), deleted)
	var count int64
	require.NoError(t, db.Model(&models.AuthSession{}).Where("refresh_expires_at <= ?", now).Count(&count).Error)
	assert.Equal(t, int64(1), count)

	deleted, err = repo.Cleanup(ctx, now, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), deleted)
	_, err = repo.GetByID(ctx, active.ID)
	assert.NoError(t, err)
	_, err = repo.GetByID(ctx, revoked.ID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func timePtr(value time.Time) *time.Time {
	return &value
}

func stringPtr(value string) *string {
	return &value
}
