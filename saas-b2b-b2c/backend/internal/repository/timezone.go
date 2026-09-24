package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantLocationForUser — S11: локация тенанта пользователя.
// Границы day/month считаются в зоне тенанта, а не в зоне сервера/UTC:
// иначе смена суток в дашборде приезжала на несколько часов позже.
// Любая ошибка (нет юзера, нет тенанта, кривая зона) = UTC, чтобы не 500-ить дашборд.
func TenantLocationForUser(ctx context.Context, db *gorm.DB, userID *uuid.UUID) *time.Location {
	if db == nil || userID == nil || *userID == uuid.Nil {
		return time.UTC
	}
	var raw sql.NullString
	row := db.WithContext(ctx).Table("users").Select("tenant_id").Where("id = ?", *userID).Row()
	if err := row.Scan(&raw); err != nil || !raw.Valid {
		return time.UTC
	}
	tenantID, err := uuid.Parse(raw.String)
	if err != nil || tenantID == uuid.Nil {
		return time.UTC
	}
	return TenantLocation(ctx, db, tenantID)
}

// TenantLocation — локация тенанта по IANA-имени; fallback UTC.
func TenantLocation(ctx context.Context, db *gorm.DB, tenantID uuid.UUID) *time.Location {
	if db == nil || tenantID == uuid.Nil {
		return time.UTC
	}
	var name string
	if err := db.WithContext(ctx).Table("tenants").Where("id = ?", tenantID).Pluck("timezone", &name).Error; err != nil {
		return time.UTC
	}
	return ResolveLocation(name)
}

// ResolveLocation — валидная IANA-зона или UTC.
func ResolveLocation(name string) *time.Location {
	name = strings.TrimSpace(name)
	if name == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}
