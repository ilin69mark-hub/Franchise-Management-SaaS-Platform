package services

import (
	"context"
	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/google/uuid"
)

type NotificationService struct {
	repo *repository.NotificationRepository
}

func NewNotificationService(repo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) GetNotifications(ctx context.Context, tenantID uuid.UUID) ([]models.Notification, error) {
	return s.repo.GetByTenant(ctx, tenantID, 20)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, id)
}

func (s *NotificationService) MarkAllAsRead(ctx context.Context, tenantID uuid.UUID) error {
	return s.repo.MarkAllAsRead(ctx, tenantID)
}

// Внутренний метод для создания уведомлений
func (s *NotificationService) CreateNotification(ctx context.Context, tenantID uuid.UUID, nType models.NotificationType, title, message string) error {
	notification := &models.Notification{
		TenantID: tenantID,
		Type:     nType,
		Title:    title,
		Message:  message,
		IsRead:   false,
	}
	return s.repo.Create(ctx, notification)
}
