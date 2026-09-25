package models

import (
	"time"

	"github.com/google/uuid"
)

type AuthSession struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID            uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	AuthVersion       int64      `json:"auth_version" gorm:"not null;default:1"`
	CurrentRefreshJTI uuid.UUID  `json:"current_refresh_jti" gorm:"type:uuid;not null;uniqueIndex"`
	ChainStartedAt    time.Time  `json:"chain_started_at" gorm:"not null"`
	ChainExpiresAt    time.Time  `json:"chain_expires_at" gorm:"not null"`
	RefreshExpiresAt  time.Time  `json:"refresh_expires_at" gorm:"not null"`
	LastTokenIssuedAt time.Time  `json:"last_token_issued_at" gorm:"not null"`
	RevokedAt         *time.Time `json:"revoked_at"`
	RevokeReason      *string    `json:"revoke_reason"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (AuthSession) TableName() string {
	return "auth_sessions"
}
