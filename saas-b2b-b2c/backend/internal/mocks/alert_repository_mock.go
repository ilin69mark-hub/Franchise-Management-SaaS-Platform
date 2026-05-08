package mocks

import (
	"context"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockAlertRepository struct {
	mock.Mock
}

func NewMockAlertRepository() *MockAlertRepository {
	return &MockAlertRepository{}
}

func (m *MockAlertRepository) CreateAlert(ctx context.Context, alert *models.Alert) (*models.Alert, error) {
	args := m.Called(ctx, alert)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetAlertByID(ctx context.Context, id uuid.UUID) (*models.Alert, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetAlertsByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.Alert, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetAlertsByUser(ctx context.Context, userID uuid.UUID) ([]models.Alert, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetUnreadAlerts(ctx context.Context, userID uuid.UUID) ([]models.Alert, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Alert), args.Error(1)
}

func (m *MockAlertRepository) MarkAsRead(ctx context.Context, alertID uuid.UUID) error {
	args := m.Called(ctx, alertID)
	return args.Error(0)
}

func (m *MockAlertRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAlertRepository) DeleteAlert(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}