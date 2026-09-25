package repository

import (
	"context"
	"time"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxAuthSessionCleanupLimit = 1000

type AuthSessionRepository struct {
	db *gorm.DB
}

func NewAuthSessionRepository(db *gorm.DB) *AuthSessionRepository {
	return &AuthSessionRepository{db: db}
}

func (r *AuthSessionRepository) Create(ctx context.Context, session *models.AuthSession) error {
	if session == nil {
		return gorm.ErrInvalidData
	}
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	if session.AuthVersion <= 0 {
		return gorm.ErrInvalidData
	}
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *AuthSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AuthSession, error) {
	var session models.AuthSession
	if err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *AuthSessionRepository) GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (*models.AuthSession, error) {
	var session models.AuthSession
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id.String(), userID.String()).
		First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *AuthSessionRepository) Revoke(ctx context.Context, id, userID uuid.UUID, reason string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.AuthSession{}).
		Where("id = ? AND user_id = ? AND revoked_at IS NULL", id.String(), userID.String()).
		Updates(authSessionRevokeValues(now, reason)).Error
}

func (r *AuthSessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID, reason string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&models.AuthSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID.String()).
		Updates(authSessionRevokeValues(now, reason))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}

	var exists int
	if err := r.db.WithContext(ctx).Raw("SELECT 1 FROM users WHERE id = ? LIMIT 1", userID.String()).Scan(&exists).Error; err != nil {
		return err
	}
	if exists == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *AuthSessionRepository) IncrementAuthVersion(ctx context.Context, userID uuid.UUID) error {
	return incrementAuthVersion(r.db.WithContext(ctx), userID)
}

func (r *AuthSessionRepository) IncrementAuthVersionAndRevoke(ctx context.Context, userID uuid.UUID, reason string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := incrementAuthVersion(tx, userID); err != nil {
			return err
		}
		return tx.Model(&models.AuthSession{}).
			Where("user_id = ? AND revoked_at IS NULL", userID.String()).
			Updates(authSessionRevokeValues(time.Now().UTC(), reason)).Error
	})
}

func (r *AuthSessionRepository) RotateRefreshJTI(ctx context.Context, sessionID, oldJTI, newJTI uuid.UUID) (bool, error) {
	if oldJTI == uuid.Nil || newJTI == uuid.Nil {
		return false, gorm.ErrInvalidData
	}
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&models.AuthSession{}).
		Where("id = ? AND current_refresh_jti = ? AND revoked_at IS NULL", sessionID.String(), oldJTI.String()).
		Updates(map[string]interface{}{
			"current_refresh_jti":  newJTI,
			"last_token_issued_at": now,
			"updated_at":           now,
		})
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 1 {
		return true, nil
	}
	var session models.AuthSession
	if err := r.db.WithContext(ctx).Where("id = ?", sessionID.String()).First(&session).Error; err != nil {
		return false, err
	}
	return false, nil
}

func (r *AuthSessionRepository) Cleanup(ctx context.Context, now time.Time, limit int) (int64, error) {
	if limit <= 0 || limit > maxAuthSessionCleanupLimit {
		limit = maxAuthSessionCleanupLimit
	}

	var result *gorm.DB
	if r.db.Name() == "postgres" {
		result = r.db.WithContext(ctx).Exec(`
			DELETE FROM auth_sessions
			WHERE id IN (
				SELECT id
				FROM auth_sessions
				WHERE revoked_at IS NOT NULL OR refresh_expires_at <= ?
				ORDER BY refresh_expires_at ASC, id ASC
				LIMIT ?
				FOR UPDATE SKIP LOCKED
			)
		`, now, limit)
	} else {
		result = r.db.WithContext(ctx).Exec(`
			DELETE FROM auth_sessions
			WHERE id IN (
				SELECT id
				FROM auth_sessions
				WHERE revoked_at IS NOT NULL OR refresh_expires_at <= ?
				ORDER BY refresh_expires_at ASC, id ASC
				LIMIT ?
			)
		`, now, limit)
	}
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func authSessionRevokeValues(now time.Time, reason string) map[string]interface{} {
	values := map[string]interface{}{
		"revoked_at": now,
		"updated_at": now,
	}
	if reason == "" {
		values["revoke_reason"] = nil
	} else {
		values["revoke_reason"] = reason
	}
	return values
}

func incrementAuthVersion(db *gorm.DB, userID uuid.UUID) error {
	result := db.Exec(
		"UPDATE users SET auth_version = auth_version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		userID.String(),
	)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
