package services

import (
	"testing"

	"franchise-saas-backend/internal/mocks"
	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newUserQuotaSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := newTestSQLiteDB(t)
	statements := []string{
		`CREATE TABLE plans (id TEXT PRIMARY KEY, name TEXT NOT NULL, price REAL NOT NULL, max_salons INTEGER NOT NULL, max_users INTEGER NOT NULL, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE tenants (id TEXT PRIMARY KEY, name TEXT NOT NULL, plan_id TEXT, status TEXT, timezone TEXT, legal_entity TEXT, inn TEXT, max_users INTEGER DEFAULT 10, paid_until DATETIME, grace_period_days INTEGER DEFAULT 7, deleted_at DATETIME, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT NOT NULL, password_hash TEXT, role TEXT NOT NULL, status TEXT, auth_version INTEGER NOT NULL DEFAULT 1, tenant_id TEXT, salon_id TEXT, managed_by TEXT, first_name TEXT, last_name TEXT, phone TEXT, display_name TEXT, position TEXT, bio TEXT, quote TEXT, avatar_url TEXT, user_status TEXT, available_for_questions BOOLEAN DEFAULT 1, achievements TEXT, contacts_phone TEXT, contacts_telegram TEXT, contacts_whatsapp TEXT, contacts_email_visible BOOLEAN DEFAULT 1, contacts_phone_visible BOOLEAN DEFAULT 1, contacts_working_hours TEXT, deleted_at DATETIME, created_at DATETIME, updated_at DATETIME)`,
	}
	for _, statement := range statements {
		require.NoError(t, db.Exec(statement).Error)
	}
	return db
}

func TestEnforceUserQuota_MissingReferencedPlanFailsClosed(t *testing.T) {
	db := newUserQuotaSQLiteDB(t)
	tenantID := uuid.New()
	planID := uuid.New()
	tenant := models.Tenant{ID: tenantID, Name: "missing-plan", MaxUsers: 10, PlanID: &planID}
	require.NoError(t, db.Create(&tenant).Error)
	require.Error(t, enforceUserQuota(db, tenantID))
}

func TestUserService_CreateEmployee_QuotaFailureDoesNotInsert(t *testing.T) {
	db := newUserQuotaSQLiteDB(t)

	tenantID := uuid.New()
	tenant := models.Tenant{ID: tenantID, Name: "quota-tenant", MaxUsers: 1}
	require.NoError(t, db.Create(&tenant).Error)
	existing := models.User{
		ID:           uuid.New(),
		Email:        "existing-quota@example.com",
		PasswordHash: "hash",
		Role:         models.RoleDealer,
		TenantID:     &tenantID,
	}
	require.NoError(t, db.Create(&existing).Error)

	createCalls := 0
	require.NoError(t, db.Callback().Create().Before("gorm:begin_transaction").Register("user_quota_failure_create_probe", func(tx *gorm.DB) {
		if tx.Statement.Table == "users" {
			createCalls++
		}
	}))

	mockRepo := mocks.NewMockUserRepository()
	mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(nil).Maybe()
	service := NewUserServiceWithInterface(mockRepo, db)

	_, err := service.CreateEmployee(models.CreateEmployeeRequest{
		Email:    "blocked-quota@example.com",
		Password: "secure-password-12",
		Role:     models.RoleDealer,
	}, tenantID, uuid.New(), string(models.RoleFranchisor))

	require.Error(t, err)
	require.Equal(t, 0, createCalls)
	var count int64
	require.NoError(t, db.Model(&models.User{}).Where("tenant_id = ?", tenantID).Count(&count).Error)
	require.Equal(t, int64(1), count)
	mockRepo.AssertNotCalled(t, "CreateUser", mock.Anything, mock.Anything)
}

func TestUserService_CreateEmployee_DatabaseInsertUsesTransaction(t *testing.T) {
	db := newUserQuotaSQLiteDB(t)

	tenantID := uuid.New()
	tenant := models.Tenant{ID: tenantID, Name: "transaction-tenant", MaxUsers: 2}
	require.NoError(t, db.Create(&tenant).Error)

	createCalls := 0
	insertInTransaction := false
	require.NoError(t, db.Callback().Create().Before("gorm:begin_transaction").Register("user_quota_success_create_probe", func(tx *gorm.DB) {
		if tx.Statement.Table == "users" {
			createCalls++
			_, insertInTransaction = tx.Statement.ConnPool.(gorm.TxCommitter)
		}
	}))

	mockRepo := mocks.NewMockUserRepository()
	mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(nil).Maybe()
	service := NewUserServiceWithInterface(mockRepo, db)

	created, err := service.CreateEmployee(models.CreateEmployeeRequest{
		Email:    "transaction-quota@example.com",
		Password: "secure-password-12",
		Role:     models.RoleDealer,
	}, tenantID, uuid.New(), string(models.RoleFranchisor))

	require.NoError(t, err)
	require.NotNil(t, created)
	require.Equal(t, 1, createCalls)
	require.True(t, insertInTransaction)
	var count int64
	require.NoError(t, db.Model(&models.User{}).Where("tenant_id = ?", tenantID).Count(&count).Error)
	require.Equal(t, int64(1), count)
	mockRepo.AssertNotCalled(t, "CreateUser", mock.Anything, mock.Anything)
}
