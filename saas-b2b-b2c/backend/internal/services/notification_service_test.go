package services

import (
	"context"
	"errors"
	"testing"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockNotificationRepo struct {
	mock.Mock
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

func TestNotificationService_GetNotifications_Success(t *testing.T) {
	mockRepo := new(MockNotificationRepo)
	service := NewNotificationService(mockRepo)

	tenantID := uuid.New()
	expected := []models.Notification{
		{ID: uuid.New(), Title: "Alert 1"},
		{ID: uuid.New(), Title: "Alert 2"},
	}

	mockRepo.On("GetByTenant", mock.Anything, tenantID, 20).Return(expected, nil)

	result, err := service.GetNotifications(context.Background(), tenantID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestNotificationService_GetNotifications_Empty(t *testing.T) {
	mockRepo := new(MockNotificationRepo)
	service := NewNotificationService(mockRepo)

	tenantID := uuid.New()

	mockRepo.On("GetByTenant", mock.Anything, tenantID, 20).Return([]models.Notification{}, nil)

	result, err := service.GetNotifications(context.Background(), tenantID)

	assert.NoError(t, err)
	assert.Len(t, result, 0)
	mockRepo.AssertExpectations(t)
}

func TestNotificationService_GetNotifications_Error(t *testing.T) {
	mockRepo := new(MockNotificationRepo)
	service := NewNotificationService(mockRepo)

	tenantID := uuid.New()

	mockRepo.On("GetByTenant", mock.Anything, tenantID, 20).Return(nil, errors.New("db error"))

	result, err := service.GetNotifications(context.Background(), tenantID)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestNotificationService_MarkAsRead_Success(t *testing.T) {
	mockRepo := new(MockNotificationRepo)
	service := NewNotificationService(mockRepo)

	notificationID := uuid.New()

	mockRepo.On("MarkAsRead", mock.Anything, notificationID).Return(nil)

	err := service.MarkAsRead(context.Background(), notificationID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestNotificationService_MarkAsRead_Error(t *testing.T) {
	mockRepo := new(MockNotificationRepo)
	service := NewNotificationService(mockRepo)

	notificationID := uuid.New()

	mockRepo.On("MarkAsRead", mock.Anything, notificationID).Return(errors.New("db error"))

	err := service.MarkAsRead(context.Background(), notificationID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestNotificationService_MarkAllAsRead_Success(t *testing.T) {
	mockRepo := new(MockNotificationRepo)
	service := NewNotificationService(mockRepo)

	tenantID := uuid.New()

	mockRepo.On("MarkAllAsRead", mock.Anything, tenantID).Return(nil)

	err := service.MarkAllAsRead(context.Background(), tenantID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestNotificationService_MarkAllAsRead_Error(t *testing.T) {
	mockRepo := new(MockNotificationRepo)
	service := NewNotificationService(mockRepo)

	tenantID := uuid.New()

	mockRepo.On("MarkAllAsRead", mock.Anything, tenantID).Return(errors.New("db error"))

	err := service.MarkAllAsRead(context.Background(), tenantID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestNotificationService_CreateNotification_Success(t *testing.T) {
	mockRepo := new(MockNotificationRepo)
	service := NewNotificationService(mockRepo)

	tenantID := uuid.New()

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(n *models.Notification) bool {
		return n.TenantID == tenantID && n.IsRead == false
	})).Return(nil)

	err := service.CreateNotification(context.Background(), tenantID, models.NotificationTypeSystem, "Test", "Message")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestNotificationService_CreateNotification_Error(t *testing.T) {
	mockRepo := new(MockNotificationRepo)
	service := NewNotificationService(mockRepo)

	tenantID := uuid.New()

	mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	err := service.CreateNotification(context.Background(), tenantID, models.NotificationTypeSystem, "Test", "Message")

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}