package services

import (
	"context"
	"testing"
	"time"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type AdminMockTenantRepo struct {
	mock.Mock
}

func (m *AdminMockTenantRepo) FindAll(ctx context.Context) ([]models.Tenant, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Tenant), args.Error(1)
}

func (m *AdminMockTenantRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *AdminMockTenantRepo) CreateTenant(ctx context.Context, tenant *models.Tenant) (*models.Tenant, error) {
	args := m.Called(ctx, tenant)
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *AdminMockTenantRepo) UpdateTenant(ctx context.Context, tenant *models.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *AdminMockTenantRepo) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *AdminMockTenantRepo) GetTenantByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *AdminMockTenantRepo) GetTenantBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *AdminMockTenantRepo) FindByStatus(ctx context.Context, status string) ([]models.Tenant, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]models.Tenant), args.Error(1)
}

func (m *AdminMockTenantRepo) CountUsersByTenantID(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(int64), args.Error(1)
}

type AdminMockAuditLogRepo struct {
	mock.Mock
}

func (m *AdminMockAuditLogRepo) CountByFilters(ctx context.Context, tenantID *uuid.UUID, startDate, endDate *time.Time) (int64, error) {
	args := m.Called(ctx, tenantID, startDate, endDate)
	return args.Get(0).(int64), args.Error(1)
}

func (m *AdminMockAuditLogRepo) FindByFilters(ctx context.Context, tenantID *uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]models.AuditLog, error) {
	args := m.Called(ctx, tenantID, startDate, endDate, limit, offset)
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

func TestAdminService_GetAuditLog_Success(t *testing.T) {
	mockTenant := &AdminMockTenantRepo{}
	mockAudit := &AdminMockAuditLogRepo{}

	tenantID := uuid.New()
	ctx := context.Background()

	mockAudit.On("CountByFilters", ctx, mock.Anything, mock.Anything, mock.Anything).Return(int64(2), nil)
	mockAudit.On("FindByFilters", ctx, mock.Anything, mock.Anything, mock.Anything, 20, 0).Return([]models.AuditLog{
		{ID: uuid.New(), Action: "login"},
		{ID: uuid.New(), Action: "view_dashboard"},
	}, nil)

	svc := NewAdminServiceWithInterfaces(mockTenant, mockAudit, nil)
	logs, total, err := svc.GetAuditLog(ctx, &tenantID, nil, nil, 20, 0)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, logs, 2)
	mockAudit.AssertExpectations(t)
}

func TestAdminService_GetAllTenants_Success(t *testing.T) {
	mockTenant := &AdminMockTenantRepo{}
	mockAudit := &AdminMockAuditLogRepo{}

	ctx := context.Background()
	tenantID := uuid.New()

	mockTenant.On("FindAll", ctx).Return([]models.Tenant{
		{ID: tenantID, Name: "Test Salon", Status: "active"},
	}, nil)

	svc := NewAdminServiceWithInterfaces(mockTenant, mockAudit, nil)
	tenants, err := svc.GetAllTenants(ctx)

	assert.NoError(t, err)
	assert.Len(t, tenants, 1)
	assert.Equal(t, "Test Salon", tenants[0]["name"])
	mockTenant.AssertExpectations(t)
}

func TestAdminService_CreateTenant_Duplicate(t *testing.T) {
	mockTenant := &AdminMockTenantRepo{}
	mockAudit := &AdminMockAuditLogRepo{}

	ctx := context.Background()

	mockTenant.On("FindAll", ctx).Return([]models.Tenant{
		{Name: "Existing"},
	}, nil)

	svc := NewAdminServiceWithInterfaces(mockTenant, mockAudit, nil)
	req := &CreateTenantRequest{Name: "Existing"}

	_, err := svc.CreateTenant(ctx, req)
	assert.Equal(t, ErrAlreadyExists, err)
	mockTenant.AssertExpectations(t)
}

func TestAdminService_SuspendTenant_NotFound(t *testing.T) {
	mockTenant := &AdminMockTenantRepo{}
	mockAudit := &AdminMockAuditLogRepo{}

	ctx := context.Background()
	tenantID := uuid.New()

	mockTenant.On("FindByID", ctx, tenantID).Return(nil, nil)

	svc := NewAdminServiceWithInterfaces(mockTenant, mockAudit, nil)
	err := svc.SuspendTenant(ctx, tenantID)

	assert.Equal(t, ErrNotFound, err)
	mockTenant.AssertExpectations(t)
}

func TestAdminService_DeleteTenant_HasUsers(t *testing.T) {
	mockTenant := &AdminMockTenantRepo{}
	mockAudit := &AdminMockAuditLogRepo{}

	ctx := context.Background()
	tenantID := uuid.New()
	tenant := &models.Tenant{ID: tenantID, Name: "Test"}

	mockTenant.On("FindByID", ctx, tenantID).Return(tenant, nil)
	mockTenant.On("CountUsersByTenantID", ctx, tenantID).Return(int64(5), nil)

	svc := NewAdminServiceWithInterfaces(mockTenant, mockAudit, nil)
	err := svc.DeleteTenant(ctx, tenantID)

	assert.Equal(t, ErrHasActiveUsers, err)
	mockTenant.AssertExpectations(t)
}

func TestAdminService_GetSystemStats_Success(t *testing.T) {
	mockTenant := &AdminMockTenantRepo{}
	mockAudit := &AdminMockAuditLogRepo{}

	ctx := context.Background()
	tenantID := uuid.New()

	mockTenant.On("FindAll", ctx).Return([]models.Tenant{
		{ID: tenantID, Name: "Active Tenant", Status: "active"},
		{ID: uuid.New(), Name: "Blocked Tenant", Status: "blocked"},
	}, nil)

	svc := NewAdminServiceWithInterfaces(mockTenant, mockAudit, nil)
	stats, err := svc.GetSystemStats(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 2, stats["total_tenants"])
	assert.Equal(t, 1, stats["active_tenants"])
	mockTenant.AssertExpectations(t)
}

func TestAdminService_DeleteTenant_Success(t *testing.T) {
	mockTenant := &AdminMockTenantRepo{}
	mockAudit := &AdminMockAuditLogRepo{}

	ctx := context.Background()
	tenantID := uuid.New()
	tenant := &models.Tenant{ID: tenantID, Name: "Test", Status: "active"}

	mockTenant.On("FindByID", ctx, tenantID).Return(tenant, nil)
	mockTenant.On("CountUsersByTenantID", ctx, tenantID).Return(int64(0), nil)
	mockTenant.On("UpdateTenant", ctx, tenant).Return(nil)

	svc := NewAdminServiceWithInterfaces(mockTenant, mockAudit, nil)
	err := svc.DeleteTenant(ctx, tenantID)

	assert.NoError(t, err)
	assert.Equal(t, "churned", tenant.Status)
	mockTenant.AssertExpectations(t)
}

func TestAdminService_DeleteTenant_NotFound(t *testing.T) {
	mockTenant := &AdminMockTenantRepo{}
	mockAudit := &AdminMockAuditLogRepo{}

	ctx := context.Background()
	tenantID := uuid.New()

	mockTenant.On("FindByID", ctx, tenantID).Return(nil, gorm.ErrRecordNotFound)

	svc := NewAdminServiceWithInterfaces(mockTenant, mockAudit, nil)
	err := svc.DeleteTenant(ctx, tenantID)

	assert.Equal(t, ErrNotFound, err)
	mockTenant.AssertExpectations(t)
}

var _ repository.TenantRepositoryInterface = (*AdminMockTenantRepo)(nil)
var _ repository.AuditLogRepositoryInterface = (*AdminMockAuditLogRepo)(nil)

func TestAdminService_CreateTenantRequest(t *testing.T) {
	mockTenant := &AdminMockTenantRepo{}
	mockAudit := &AdminMockAuditLogRepo{}

	ctx := context.Background()

	mockTenant.On("FindAll", ctx).Return([]models.Tenant{}, nil)
	mockTenant.On("CreateTenant", ctx, mock.Anything).Return(&models.Tenant{
		ID:   uuid.New(),
		Name: "New Tenant",
	}, nil)

	svc := NewAdminServiceWithInterfaces(mockTenant, mockAudit, nil)
	req := &CreateTenantRequest{
		Name:        "New Tenant",
		LegalEntity: "OOO New",
		INN:         "1234567890",
		MaxUsers:    50,
	}

	tenant, err := svc.CreateTenant(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Equal(t, "New Tenant", tenant.Name)
	mockTenant.AssertExpectations(t)
}

func TestAdminService_GetAuditLog_WithPagination(t *testing.T) {
	mockTenant := &AdminMockTenantRepo{}
	mockAudit := &AdminMockAuditLogRepo{}

	ctx := context.Background()
	tenantID := uuid.New()

	mockAudit.On("CountByFilters", ctx, mock.Anything, mock.Anything, mock.Anything).Return(int64(100), nil)
	mockAudit.On("FindByFilters", ctx, mock.Anything, mock.Anything, mock.Anything, 10, 20).Return([]models.AuditLog{
		{ID: uuid.New(), Action: "test1"},
		{ID: uuid.New(), Action: "test2"},
	}, nil)

	svc := NewAdminServiceWithInterfaces(mockTenant, mockAudit, nil)
	logs, total, err := svc.GetAuditLog(ctx, &tenantID, nil, nil, 10, 20)

	assert.NoError(t, err)
	assert.Equal(t, int64(100), total)
	assert.Len(t, logs, 2)
}

func TestAdminService_GetAuditLog_DefaultLimit(t *testing.T) {
	mockTenant := &AdminMockTenantRepo{}
	mockAudit := &AdminMockAuditLogRepo{}

	ctx := context.Background()
	tenantID := uuid.New()

	mockAudit.On("CountByFilters", ctx, mock.Anything, mock.Anything, mock.Anything).Return(int64(5), nil)
	mockAudit.On("FindByFilters", ctx, mock.Anything, mock.Anything, mock.Anything, 20, 0).Return([]models.AuditLog{
		{ID: uuid.New()},
	}, nil)

	svc := NewAdminServiceWithInterfaces(mockTenant, mockAudit, nil)
	_, total, err := svc.GetAuditLog(ctx, &tenantID, nil, nil, 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
}