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

// GetByID возвращает уведомление по ID или (nil, nil), если не найдено.
// Используется для проверки владельца перед изменением статуса прочтения.
func (r *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	var n models.Notification
	err := r.db.WithContext(ctx).First(&n, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
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
