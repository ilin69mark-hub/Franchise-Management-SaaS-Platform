package services

import (
	"context"
	"errors"
	"testing"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockNotificationRepo struct {
	mock.Mock
}

func NewMockNotificationRepo() *MockNotificationRepo {
	return &MockNotificationRepo{}
}

func (m *MockNotificationRepo) GetByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]models.Notification, error) {
	args := m.Called(ctx, tenantID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Notification), args.Error(1)
}

func (m *MockNotificationRepo) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNotificationRepo) MarkAllAsRead(ctx context.Context, tenantID uuid.UUID) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *MockNotificationRepo) Create(ctx context.Context, notification *models.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

type NotificationRepoInterface interface {
	GetByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]models.Notification, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, tenantID uuid.UUID) error
	Create(ctx context.Context, notification *models.Notification) error
}

type NotificationServiceTestable struct {
	repo NotificationRepoInterface
}

func NewNotificationServiceTestable(repo NotificationRepoInterface) *NotificationServiceTestable {
	return &NotificationServiceTestable{repo: repo}
}

func (s *NotificationServiceTestable) GetNotifications(ctx context.Context, tenantID uuid.UUID) ([]models.Notification, error) {
	return s.repo.GetByTenant(ctx, tenantID, 20)
}

func (s *NotificationServiceTestable) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, id)
}

func (s *NotificationServiceTestable) MarkAllAsRead(ctx context.Context, tenantID uuid.UUID) error {
	return s.repo.MarkAllAsRead(ctx, tenantID)
}

func (s *NotificationServiceTestable) CreateNotification(ctx context.Context, tenantID uuid.UUID, nType models.NotificationType, title, message string) error {
	notification := &models.Notification{
		TenantID: tenantID,
		Type:     nType,
		Title:    title,
		Message:  message,
		IsRead:   false,
	}
	return s.repo.Create(ctx, notification)
}

func TestNotificationService_GetNotifications(t *testing.T) {
	tenantID := uuid.New()

	t.Run("success returns notifications", func(t *testing.T) {
		repo := NewMockNotificationRepo()
		service := NewNotificationServiceTestable(repo)

		notifications := []models.Notification{
			{ID: uuid.New(), TenantID: tenantID, Title: "New Lead", IsRead: false},
			{ID: uuid.New(), TenantID: tenantID, Title: "Payment Received", IsRead: true},
		}
		repo.On("GetByTenant", mock.Anything, tenantID, 20).Return(notifications, nil)

		result, err := service.GetNotifications(context.Background(), tenantID)

		require.NoError(t, err)
		require.Len(t, result, 2)
		require.Equal(t, "New Lead", result[0].Title)
		repo.AssertExpectations(t)
	})

	t.Run("empty returns empty array", func(t *testing.T) {
		repo := NewMockNotificationRepo()
		service := NewNotificationServiceTestable(repo)

		repo.On("GetByTenant", mock.Anything, tenantID, 20).Return([]models.Notification{}, nil)

		result, err := service.GetNotifications(context.Background(), tenantID)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result, 0)
		repo.AssertExpectations(t)
	})

	t.Run("repository error returns error", func(t *testing.T) {
		repo := NewMockNotificationRepo()
		service := NewNotificationServiceTestable(repo)

		repo.On("GetByTenant", mock.Anything, tenantID, 20).Return(nil, errors.New("db error"))

		result, err := service.GetNotifications(context.Background(), tenantID)

		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "db error", err.Error())
		repo.AssertExpectations(t)
	})
}

func TestNotificationService_MarkAsRead(t *testing.T) {
	notificationID := uuid.New()

	t.Run("success marks as read", func(t *testing.T) {
		repo := NewMockNotificationRepo()
		service := NewNotificationServiceTestable(repo)

		repo.On("MarkAsRead", mock.Anything, notificationID).Return(nil)

		err := service.MarkAsRead(context.Background(), notificationID)

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("repository error returns error", func(t *testing.T) {
		repo := NewMockNotificationRepo()
		service := NewNotificationServiceTestable(repo)

		repo.On("MarkAsRead", mock.Anything, notificationID).Return(errors.New("not found"))

		err := service.MarkAsRead(context.Background(), notificationID)

		require.Error(t, err)
		require.Equal(t, "not found", err.Error())
		repo.AssertExpectations(t)
	})
}

func TestNotificationService_MarkAllAsRead(t *testing.T) {
	tenantID := uuid.New()

	t.Run("success marks all as read", func(t *testing.T) {
		repo := NewMockNotificationRepo()
		service := NewNotificationServiceTestable(repo)

		repo.On("MarkAllAsRead", mock.Anything, tenantID).Return(nil)

		err := service.MarkAllAsRead(context.Background(), tenantID)

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("repository error returns error", func(t *testing.T) {
		repo := NewMockNotificationRepo()
		service := NewNotificationServiceTestable(repo)

		repo.On("MarkAllAsRead", mock.Anything, tenantID).Return(errors.New("constraint error"))

		err := service.MarkAllAsRead(context.Background(), tenantID)

		require.Error(t, err)
		require.Equal(t, "constraint error", err.Error())
		repo.AssertExpectations(t)
	})
}

func TestNotificationService_CreateNotification(t *testing.T) {
	tenantID := uuid.New()

	t.Run("success creates notification", func(t *testing.T) {
		repo := NewMockNotificationRepo()
		service := NewNotificationServiceTestable(repo)

		repo.On("Create", mock.Anything, mock.MatchedBy(func(n *models.Notification) bool {
			return n.TenantID == tenantID && n.Title == "Test Alert" && n.IsRead == false
		})).Return(nil)

		err := service.CreateNotification(context.Background(), tenantID, models.NotificationTypeSystem, "Test Alert", "Test message")

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("creates with different type", func(t *testing.T) {
		repo := NewMockNotificationRepo()
		service := NewNotificationServiceTestable(repo)

		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		err := service.CreateNotification(context.Background(), tenantID, models.NotificationTypePayment, "Payment", "New payment")

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("repository error returns error", func(t *testing.T) {
		repo := NewMockNotificationRepo()
		service := NewNotificationServiceTestable(repo)

		repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("validation error"))

		err := service.CreateNotification(context.Background(), tenantID, models.NotificationTypeSystem, "Alert", "Message")

		require.Error(t, err)
		require.Equal(t, "validation error", err.Error())
		repo.AssertExpectations(t)
	})
}