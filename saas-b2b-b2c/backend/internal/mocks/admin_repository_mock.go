package mocks

import (
	"context"
	"time"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockTenantRepository struct {
	mock.Mock
}

func NewMockTenantRepository() *MockTenantRepository {
	return &MockTenantRepository{}
}

func (m *MockTenantRepository) CreateTenant(ctx context.Context, tenant *models.Tenant) (*models.Tenant, error) {
	args := m.Called(ctx, tenant)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantRepository) GetTenantByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantRepository) GetTenantBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantRepository) UpdateTenant(ctx context.Context, tenant *models.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTenantRepository) FindAll(ctx context.Context) ([]models.Tenant, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Tenant), args.Error(1)
}

func (m *MockTenantRepository) FindByStatus(ctx context.Context, status string) ([]models.Tenant, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Tenant), args.Error(1)
}

func (m *MockTenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantRepository) CountUsersByTenantID(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(int64), args.Error(1)
}

type MockAuditLogRepository struct {
	mock.Mock
}

func NewMockAuditLogRepository() *MockAuditLogRepository {
	return &MockAuditLogRepository{}
}

func (m *MockAuditLogRepository) FindByFilters(ctx context.Context, tenantID *uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]models.AuditLog, error) {
	args := m.Called(ctx, tenantID, startDate, endDate, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) CountByFilters(ctx context.Context, tenantID *uuid.UUID, startDate, endDate *time.Time) (int64, error) {
	args := m.Called(ctx, tenantID, startDate, endDate)
	return args.Get(0).(int64), args.Error(1)
}