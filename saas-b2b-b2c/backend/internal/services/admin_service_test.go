package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"franchise-saas-backend/internal/mocks"
	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAdminService_GetAllTenants(t *testing.T) {
	t.Run("success returns list of tenants", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		tenants := []models.Tenant{
			{ID: uuid.New(), Name: "Tenant 1", Status: "active"},
			{ID: uuid.New(), Name: "Tenant 2", Status: "active"},
			{ID: uuid.New(), Name: "Tenant 3", Status: "suspended"},
		}
		tenantRepo.On("FindAll", mock.Anything).Return(tenants, nil)

		result, err := service.GetAllTenants(context.Background())

		require.NoError(t, err)
		require.Len(t, result, 3)
		tenantRepo.AssertExpectations(t)
	})

	t.Run("empty database returns empty array", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		tenantRepo.On("FindAll", mock.Anything).Return([]models.Tenant{}, nil)

		result, err := service.GetAllTenants(context.Background())

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result, 0)
		tenantRepo.AssertExpectations(t)
	})

	t.Run("repository error returns error", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		tenantRepo.On("FindAll", mock.Anything).Return(nil, errors.New("db error"))

		result, err := service.GetAllTenants(context.Background())

		require.Error(t, err)
		require.Nil(t, result)
		tenantRepo.AssertExpectations(t)
	})
}

func TestAdminService_CreateTenant(t *testing.T) {
	t.Run("success creates tenant", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		tenants := []models.Tenant{}
		tenantRepo.On("FindAll", mock.Anything).Return(tenants, nil)

		newTenant := &models.Tenant{
			ID:     uuid.New(),
			Name:   "New Tenant",
			Status: "active",
		}
		tenantRepo.On("CreateTenant", mock.Anything, mock.Anything).Return(newTenant, nil)

		req := &CreateTenantRequest{
			Name:        "New Tenant",
			LegalEntity: "LLC",
			INN:         "1234567890",
			MaxUsers:    10,
		}

		result, err := service.CreateTenant(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "New Tenant", result.Name)
		tenantRepo.AssertExpectations(t)
	})

	t.Run("duplicate name returns ErrAlreadyExists", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		existingTenant := models.Tenant{ID: uuid.New(), Name: "Existing Tenant"}
		tenants := []models.Tenant{existingTenant}
		tenantRepo.On("FindAll", mock.Anything).Return(tenants, nil)

		req := &CreateTenantRequest{
			Name: "Existing Tenant",
		}

		result, err := service.CreateTenant(context.Background(), req)

		require.Error(t, err)
		require.ErrorIs(t, err, ErrAlreadyExists)
		require.Nil(t, result)
		tenantRepo.AssertExpectations(t)
	})

	t.Run("repository error on create returns error", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		tenants := []models.Tenant{}
		tenantRepo.On("FindAll", mock.Anything).Return(tenants, nil)
		tenantRepo.On("CreateTenant", mock.Anything, mock.Anything).Return(nil, errors.New("create error"))

		req := &CreateTenantRequest{
			Name: "New Tenant",
		}

		result, err := service.CreateTenant(context.Background(), req)

		require.Error(t, err)
		require.Nil(t, result)
		tenantRepo.AssertExpectations(t)
	})
}

func TestAdminService_SuspendTenant(t *testing.T) {
	tenantID := uuid.New()

	t.Run("success suspends tenant", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		tenant := &models.Tenant{ID: tenantID, Name: "Test Tenant", Status: "active"}
		tenantRepo.On("FindByID", mock.Anything, tenantID).Return(tenant, nil)
		tenantRepo.On("UpdateTenant", mock.Anything, mock.Anything).Return(nil)

		err := service.SuspendTenant(context.Background(), tenantID)

		require.NoError(t, err)
		tenantRepo.AssertExpectations(t)
	})

	t.Run("non-existent tenant returns ErrNotFound", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		tenantRepo.On("FindByID", mock.Anything, tenantID).Return(nil, errors.New("not found"))

		err := service.SuspendTenant(context.Background(), tenantID)

		require.Error(t, err)
		require.ErrorIs(t, err, ErrNotFound)
		tenantRepo.AssertExpectations(t)
	})

	t.Run("tenant not found (nil) returns ErrNotFound", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		tenantRepo.On("FindByID", mock.Anything, tenantID).Return(nil, nil)

		err := service.SuspendTenant(context.Background(), tenantID)

		require.Error(t, err)
		require.ErrorIs(t, err, ErrNotFound)
		tenantRepo.AssertExpectations(t)
	})
}

func TestAdminService_DeleteTenant(t *testing.T) {
	tenantID := uuid.New()

	t.Run("success deletes tenant", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		tenant := &models.Tenant{ID: tenantID, Name: "Test Tenant", Status: "active"}
		tenantRepo.On("FindByID", mock.Anything, tenantID).Return(tenant, nil)
		tenantRepo.On("CountUsersByTenantID", mock.Anything, tenantID).Return(int64(0), nil)
		tenantRepo.On("UpdateTenant", mock.Anything, mock.Anything).Return(nil)

		err := service.DeleteTenant(context.Background(), tenantID)

		require.NoError(t, err)
		tenantRepo.AssertExpectations(t)
	})

	t.Run("tenant with active users returns ErrHasActiveUsers", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		tenant := &models.Tenant{ID: tenantID, Name: "Test Tenant", Status: "active"}
		tenantRepo.On("FindByID", mock.Anything, tenantID).Return(tenant, nil)
		tenantRepo.On("CountUsersByTenantID", mock.Anything, tenantID).Return(int64(5), nil)

		err := service.DeleteTenant(context.Background(), tenantID)

		require.Error(t, err)
		require.ErrorIs(t, err, ErrHasActiveUsers)
		tenantRepo.AssertExpectations(t)
	})

	t.Run("non-existent tenant returns ErrNotFound", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		tenantRepo.On("FindByID", mock.Anything, tenantID).Return(nil, errors.New("not found"))

		err := service.DeleteTenant(context.Background(), tenantID)

		require.Error(t, err)
		require.ErrorIs(t, err, ErrNotFound)
		tenantRepo.AssertExpectations(t)
	})
}

func TestAdminService_GetAuditLog(t *testing.T) {
	t.Run("success returns array with filters", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		logs := []models.AuditLog{
			{ID: uuid.New(), Action: "user_created", TenantID: uuid.New()},
			{ID: uuid.New(), Action: "tenant_updated", TenantID: uuid.New()},
		}
		tenantID := uuid.New()
		auditRepo.On("CountByFilters", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(2), nil)
		auditRepo.On("FindByFilters", mock.Anything, mock.Anything, mock.Anything, mock.Anything, 20, 0).Return(logs, nil)

		result, total, err := service.GetAuditLog(context.Background(), &tenantID, nil, nil, 20, 0)

		require.NoError(t, err)
		require.Len(t, result, 2)
		require.Equal(t, int64(2), total)
		auditRepo.AssertExpectations(t)
	})

	t.Run("pagination with limit and offset", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		logs := []models.AuditLog{
			{ID: uuid.New(), Action: "user_created"},
		}
		auditRepo.On("CountByFilters", mock.Anything, (*uuid.UUID)(nil), (*time.Time)(nil), (*time.Time)(nil)).Return(int64(100), nil)
		auditRepo.On("FindByFilters", mock.Anything, (*uuid.UUID)(nil), (*time.Time)(nil), (*time.Time)(nil), 10, 20).Return(logs, nil)

		result, total, err := service.GetAuditLog(context.Background(), nil, nil, nil, 10, 20)

		require.NoError(t, err)
		require.Len(t, result, 1)
		require.Equal(t, int64(100), total)
		auditRepo.AssertExpectations(t)
	})

	t.Run("default limit when 0", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		logs := []models.AuditLog{}
		auditRepo.On("CountByFilters", mock.Anything, (*uuid.UUID)(nil), (*time.Time)(nil), (*time.Time)(nil)).Return(int64(0), nil)
		auditRepo.On("FindByFilters", mock.Anything, (*uuid.UUID)(nil), (*time.Time)(nil), (*time.Time)(nil), 20, 0).Return(logs, nil)

		_, total, err := service.GetAuditLog(context.Background(), nil, nil, nil, 0, 0)

		require.NoError(t, err)
		require.Equal(t, int64(0), total)
		auditRepo.AssertExpectations(t)
	})
}

func TestAdminService_GetSystemStats(t *testing.T) {
	t.Run("success returns system statistics", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		tenants := []models.Tenant{
			{ID: uuid.New(), Name: "Tenant 1", Status: "active"},
			{ID: uuid.New(), Name: "Tenant 2", Status: "active"},
			{ID: uuid.New(), Name: "Tenant 3", Status: "churned"},
		}
		tenantRepo.On("FindAll", mock.Anything).Return(tenants, nil)

		result, err := service.GetSystemStats(context.Background())

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 3, result["total_tenants"])
		require.Equal(t, 2, result["active_tenants"])
		tenantRepo.AssertExpectations(t)
	})

	t.Run("returns 0 for active tenants when all churned", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		tenants := []models.Tenant{
			{ID: uuid.New(), Status: "churned"},
			{ID: uuid.New(), Status: "churned"},
		}
		tenantRepo.On("FindAll", mock.Anything).Return(tenants, nil)

		result, err := service.GetSystemStats(context.Background())

		require.NoError(t, err)
		require.Equal(t, 2, result["total_tenants"])
		require.Equal(t, 0, result["active_tenants"])
		tenantRepo.AssertExpectations(t)
	})

	t.Run("repository error returns error", func(t *testing.T) {
		tenantRepo := mocks.NewMockTenantRepository()
		auditRepo := mocks.NewMockAuditLogRepository()
		service := NewAdminServiceWithInterfaces(tenantRepo, auditRepo, nil)

		tenantRepo.On("FindAll", mock.Anything).Return(nil, errors.New("db error"))

		result, err := service.GetSystemStats(context.Background())

		require.Error(t, err)
		require.Nil(t, result)
		tenantRepo.AssertExpectations(t)
	})
}