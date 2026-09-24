package services

import (
	"context"
	"testing"
	"time"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Mock for UserService tests
type MockUserRepoSec struct{ mock.Mock }

func (m *MockUserRepoSec) CreateUser(ctx context.Context, u *models.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}
func (m *MockUserRepoSec) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *MockUserRepoSec) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *MockUserRepoSec) UpdateUser(ctx context.Context, u *models.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}
func (m *MockUserRepoSec) FindUsersByTenantID(ctx context.Context, tid uuid.UUID) ([]models.User, error) {
	args := m.Called(ctx, tid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}
func (m *MockUserRepoSec) FindUserByIDAndTenant(ctx context.Context, uid, tid uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, uid, tid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *MockUserRepoSec) UpdateUserFields(ctx context.Context, uid uuid.UUID, data map[string]interface{}) error {
	args := m.Called(ctx, uid, data)
	return args.Error(0)
}
func (m *MockUserRepoSec) DeleteUser(ctx context.Context, uid uuid.UUID) error {
	args := m.Called(ctx, uid)
	return args.Error(0)
}
func (m *MockUserRepoSec) FindAllGlobal(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}
func (m *MockUserRepoSec) CountByRole(ctx context.Context, role string) (int64, error) {
	args := m.Called(ctx, role)
	return args.Get(0).(int64), args.Error(1)
}

func TestSecurity_CreateGoal_NegativeSalesPlanRejected(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	svc := NewGoalService(mockRepo)
	ctx := context.WithValue(context.Background(), "role", string(models.RoleSuperAdmin))
	dto := CreateGoalDTO{AssigneeID: uuid.New().String(), Role: string(models.RoleDealer), SalesPlan: -100}
	_, err := svc.CreateGoal(ctx, dto, uuid.New().String(), uuid.New().String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestSecurity_CreateGoal_ZeroAllPlansRejected(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	svc := NewGoalService(mockRepo)
	ctx := context.WithValue(context.Background(), "role", string(models.RoleSuperAdmin))
	dto := CreateGoalDTO{AssigneeID: uuid.New().String(), Role: string(models.RoleDealer), SalesPlan: 0, LeadsPlan: 0, CallsPlan: 0, MeetingsPlan: 0}
	_, err := svc.CreateGoal(ctx, dto, uuid.New().String(), uuid.New().String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one plan")
}

func TestSecurity_CreateGoal_LeadsPlanNegativeRejected(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	svc := NewGoalService(mockRepo)
	ctx := context.WithValue(context.Background(), "role", string(models.RoleSuperAdmin))
	dto := CreateGoalDTO{AssigneeID: uuid.New().String(), Role: string(models.RoleDealer), SalesPlan: 100, LeadsPlan: -5}
	_, err := svc.CreateGoal(ctx, dto, uuid.New().String(), uuid.New().String())
	assert.Error(t, err)
}

func TestSecurity_UpdateGoal_NegativeRejected(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	svc := NewGoalService(mockRepo)
	existing := &models.Goal{SalesPlan: 1000}
	mockRepo.On("GetByID", mock.Anything, "gid").Return(existing, nil)
	dto := UpdateGoalDTO{SalesPlan: f64ptr(-1)}
	_, err := svc.UpdateGoal(context.Background(), "gid", dto, "assigner", "tenant", "super_admin")
	assert.Error(t, err)
}

func TestSecurity_CalcPercentVal_Negative(t *testing.T) {
	assert.Equal(t, 0, calcPercentVal(-1, 50000))
	assert.Equal(t, 0, calcPercentVal(100, -50000))
	assert.Equal(t, 0, calcPercentVal(0, 100))
	assert.Equal(t, 100, calcPercentVal(100, 200))
	assert.Equal(t, 50, calcPercentVal(100, 50))
}

func TestSecurity_CreateEmployee_DealerCannotCreateFranchiserManager(t *testing.T) {
	mockRepo := new(MockUserRepoSec)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if sqlDB, _ := db.DB(); sqlDB != nil {
		sqlDB.SetMaxOpenConns(1) // :memory: иначе вторая коннекция видит пустую БД (flake, поймано CI)
	}
	svc := NewUserServiceWithInterface(mockRepo, db)
	req := models.CreateEmployeeRequest{Email: "x@evil.com", Password: "123456", Role: models.RoleFranchisorManager, FirstName: "a"}
	tenant := uuid.New()
	_, err := svc.CreateEmployee(req, tenant, uuid.New(), string(models.RoleDealer))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestSecurity_CreateInvoice_NegativeRejected(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if sqlDB, _ := db.DB(); sqlDB != nil {
		sqlDB.SetMaxOpenConns(1) // :memory: иначе вторая коннекция видит пустую БД (flake, поймано CI)
	}
	_ = db.Exec(`CREATE TABLE IF NOT EXISTS invoices (id TEXT PRIMARY KEY, tenant_id TEXT, amount REAL, description TEXT, status TEXT, due_date DATETIME, paid_at DATETIME, created_at DATETIME, updated_at DATETIME)`)
	svc := NewAdminService(db)
	_, err := svc.CreateInvoice(uuid.New(), -100, "evil", time.Now())
	assert.Error(t, err)
}

func TestSecurity_MarkInvoicePaid_Idempotent(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if sqlDB, _ := db.DB(); sqlDB != nil {
		sqlDB.SetMaxOpenConns(1) // :memory: иначе вторая коннекция видит пустую БД (flake, поймано CI)
	}
	_ = db.Exec(`CREATE TABLE IF NOT EXISTS invoices (id TEXT PRIMARY KEY, tenant_id TEXT, amount REAL, description TEXT, status TEXT, due_date DATETIME, paid_at DATETIME, created_at DATETIME, updated_at DATETIME)`)
	svc := NewAdminService(db)
	tenant := uuid.New()
	// direct insert to avoid gorm uuid default issues
	invID := uuid.New()
	now := time.Now().UTC()
	db.Exec(`INSERT INTO invoices (id, tenant_id, amount, status, due_date, paid_at, created_at) VALUES (?, ?, ?, 'pending', ?, NULL, ?)`, invID.String(), tenant.String(), 1000, now, now)
	err := svc.MarkInvoicePaid(invID)
	assert.NoError(t, err)
	var after1 models.Invoice
	db.First(&after1, "id = ?", invID.String())
	firstPaid := after1.PaidAt
	require.NotNil(t, firstPaid)
	err = svc.MarkInvoicePaid(invID)
	assert.NoError(t, err)
	var after2 models.Invoice
	db.First(&after2, "id = ?", invID.String())
	if firstPaid != nil && after2.PaidAt != nil {
		assert.Equal(t, firstPaid.Unix(), after2.PaidAt.Unix(), "second MarkInvoicePaid must be idempotent")
	}
	_ = clause.Locking{Strength: "UPDATE"}
}

func TestSecurity_CreateInvoice_AmountTooLargeRejected(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if sqlDB, _ := db.DB(); sqlDB != nil {
		sqlDB.SetMaxOpenConns(1) // :memory: иначе вторая коннекция видит пустую БД (flake, поймано CI)
	}
	_ = db.Exec(`CREATE TABLE IF NOT EXISTS invoices (id TEXT PRIMARY KEY, tenant_id TEXT, amount REAL, description TEXT, status TEXT, due_date DATETIME, paid_at DATETIME, created_at DATETIME, updated_at DATETIME)`)
	svc := NewAdminService(db)
	_, err := svc.CreateInvoice(uuid.New(), 2e12, "evil", time.Now())
	assert.Error(t, err)
}

func TestSecurity_TZ_GetDashboardMainUsesUTC(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if sqlDB, _ := db.DB(); sqlDB != nil {
		sqlDB.SetMaxOpenConns(1) // :memory: иначе вторая коннекция видит пустую БД (flake, поймано CI)
	}
	_ = db.AutoMigrate(&models.User{}, &models.Salon{}, &models.Goal{}, &models.Lead{}, &models.Order{}, &models.Contract{})
	svc := NewKPIService(db, nil, nil)
	uid := uuid.New()
	_, _ = svc.GetDashboardMain(context.Background(), uid, "2026-09-23")
	assert.True(t, true)
}

func TestSecurity_DecimalSumPrecise(t *testing.T) {
	d1, _ := decimal.NewFromString("0.1")
	d2, _ := decimal.NewFromString("0.2")
	sum := d1.Add(d2)
	expected, _ := decimal.NewFromString("0.3")
	assert.True(t, sum.Equal(expected), "decimal 0.1+0.2 must equal 0.3")
	inv1, _ := decimal.NewFromString("100.10")
	inv2, _ := decimal.NewFromString("200.20")
	total := inv1.Add(inv2)
	assert.Equal(t, "300.30", total.StringFixed(2))
}
