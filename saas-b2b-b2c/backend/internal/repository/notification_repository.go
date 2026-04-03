package repository

import (
	"context"
	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *models.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *NotificationRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at desc").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.Notification{}).Where("id = ?", id).Update("is_read", true).Error
}

func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("tenant_id = ?", tenantID).
		Update("is_read", true).Error
}
