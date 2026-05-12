package repository

import (
	"context"
	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
)

type NotificationRepositoryInterface interface {
	Create(ctx context.Context, n *models.Notification) error
	GetByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]models.Notification, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, tenantID uuid.UUID) error
}

var _ NotificationRepositoryInterface = (*NotificationRepository)(nil)