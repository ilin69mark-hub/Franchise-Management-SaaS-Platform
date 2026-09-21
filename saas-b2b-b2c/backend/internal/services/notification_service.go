package services

import (
	"context"
	"errors"
	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/google/uuid"
)

type NotificationService struct {
	repo repository.NotificationRepositoryInterface
}

func NewNotificationService(repo repository.NotificationRepositoryInterface) *NotificationService {
	return &NotificationService{repo: repo}
}

// notificationOwnerLookup - необязательная возможность репозитория достать
// уведомление по ID, чтобы сервис мог проверить владельца. Реализуется
// конкретным NotificationRepository; заглушки/тесты могут её не иметь.
type notificationOwnerLookup interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error)
}

func (s *NotificationService) GetNotifications(ctx context.Context, tenantID uuid.UUID) ([]models.Notification, error) {
	return s.repo.GetByTenant(ctx, tenantID, 20)
}

// MarkAsRead помечает уведомление прочитанным. Если передан userID, то
// операция разрешена только владельцу уведомления (иначе ErrForbidden).
func (s *NotificationService) MarkAsRead(ctx context.Context, id uuid.UUID, userID ...uuid.UUID) error {
	if len(userID) > 0 {
		uid := userID[0]
		if lookup, ok := s.repo.(notificationOwnerLookup); ok {
			n, err := lookup.GetByID(ctx, id)
			if err != nil {
				return err
			}
			if n == nil {
				return errors.New("notification not found")
			}
			if n.UserID == nil || *n.UserID != uid {
				return ErrForbidden
			}
		}
	}
	return s.repo.MarkAsRead(ctx, id)
}

// MarkAllAsRead помечает прочитанными уведомления указанного пользователя в
// рамках тенанта. Если userID не передан, сохраняется старое поведение
// (пометить всё в тенанте) для обратной совместимости.
func (s *NotificationService) MarkAllAsRead(ctx context.Context, tenantID uuid.UUID, userID ...uuid.UUID) error {
	if len(userID) > 0 {
		uid := userID[0]
		notifications, err := s.repo.GetByTenant(ctx, tenantID, 1000)
		if err != nil {
			return err
		}
		for i := range notifications {
			n := notifications[i]
			if n.UserID != nil && *n.UserID == uid && !n.IsRead {
				if err := s.repo.MarkAsRead(ctx, n.ID); err != nil {
					return err
				}
			}
		}
		return nil
	}
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
