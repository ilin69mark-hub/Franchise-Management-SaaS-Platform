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
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	conn, err := db.DB()
	require.NoError(t, err)
	conn.SetMaxOpenConns(1)
	return db
}

func setupDashboardStatsTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	mustExec := func(sql string) {
		t.Helper()
		require.NoError(t, db.Exec(sql).Error)
	}
	mustExec("CREATE TABLE leads (id TEXT PRIMARY KEY, salon_id TEXT, manager_id TEXT, status TEXT, budget REAL, created_at DATETIME, updated_at DATETIME)")
	mustExec("CREATE TABLE lead_activities (id TEXT PRIMARY KEY, lead_id TEXT, user_id TEXT, type TEXT, description TEXT, created_at DATETIME)")
}

func insertTestLead(t *testing.T, db *gorm.DB, id, salonID, managerID uuid.UUID, status string, budget float64, createdAt time.Time) {
	t.Helper()
	err := db.Exec("INSERT INTO leads (id, salon_id, manager_id, status, budget, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id.String(), salonID.String(), managerID.String(), status, budget, createdAt, createdAt).Error
	require.NoError(t, err)
}

func insertTestActivity(t *testing.T, db *gorm.DB, leadID, userID uuid.UUID, activityType string, createdAt time.Time) {
	t.Helper()
	err := db.Exec("INSERT INTO lead_activities (id, lead_id, user_id, type, description, created_at) VALUES (?, ?, ?, ?, '', ?)",
		uuid.New().String(), leadID.String(), userID.String(), activityType, createdAt).Error
	require.NoError(t, err)
}

func TestKPIService_GetDashboardStats_BuildsCallsMeetingQueriesIndependently(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupDashboardStatsTables(t, db)

	today := time.Now()
	salonID := uuid.New()
	otherSalonID := uuid.New()
	mgrID := uuid.New()
	userID := uuid.New()

	leadInSalon := uuid.New()
	leadOtherSalon := uuid.New()

	insertTestLead(t, db, leadInSalon, salonID, mgrID, "sale", 100, today)
	insertTestLead(t, db, leadOtherSalon, otherSalonID, mgrID, "new", 50, today)

	insertTestActivity(t, db, leadInSalon, mgrID, "call", today)
	insertTestActivity(t, db, leadInSalon, mgrID, "call", today)
	insertTestActivity(t, db, leadInSalon, mgrID, "meeting", today)
	insertTestActivity(t, db, leadOtherSalon, mgrID, "call", today)

	goal := &models.DailyGoal{SalesPlan: 100, LeadsPlan: 10, CallsPlan: 10, MeetingsPlan: 10, TargetDate: today}
	mockKpi := new(MockKPIRepositoryForTest)
	mockKpi.On("GetGoal", mock.Anything, (*uuid.UUID)(nil), &salonID, mock.Anything).Return(goal, nil)
	mockKpi.On("GetGoal", mock.Anything, &mgrID, (*uuid.UUID)(nil), mock.Anything).Return(goal, nil)

	svc := NewKPIService(db, mockKpi, nil)

	t.Run("salon scoped via EXISTS subquery", func(t *testing.T) {
		resp, err := svc.GetDashboardStats(context.Background(), userID, salonID, false)
		require.NoError(t, err)
		require.Equal(t, 2.0, resp.Calls.Fact)
		require.Equal(t, 1.0, resp.Meetings.Fact)
	})

	t.Run("manager scoped by user_id", func(t *testing.T) {
		resp, err := svc.GetDashboardStats(context.Background(), mgrID, salonID, true)
		require.NoError(t, err)
		require.Equal(t, 3.0, resp.Calls.Fact)
		require.Equal(t, 1.0, resp.Meetings.Fact)
	})
}

func TestKPIService_GetDealerRequests_MissingTableReturnsEmpty(t *testing.T) {
	db := newTestSQLiteDB(t)
	svc := NewKPIService(db, nil, nil)

	resp, err := svc.GetDealerRequests(context.Background(), uuid.New(), "")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Requests, 0)
	require.Equal(t, 0, resp.Total)
}

func TestKPIService_GetDealerRequests_ReadsRows(t *testing.T) {
	db := newTestSQLiteDB(t)
	require.NoError(t, db.Exec(`
		CREATE TABLE dealer_requests (
			id TEXT PRIMARY KEY, dealer_id TEXT, type TEXT, description TEXT,
			amount REAL, status TEXT, created_at DATETIME, updated_at DATETIME)`).Error)

	dealerID := uuid.New()
	now := time.Now()
	require.NoError(t, db.Exec(
		`INSERT INTO dealer_requests (id, dealer_id, type, description, amount, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), dealerID.String(), "promo", "desc", 1234.5, "pending", now, now).Error)

	svc := NewKPIService(db, nil, nil)
	resp, err := svc.GetDealerRequests(context.Background(), dealerID, "")
	require.NoError(t, err)
	require.Len(t, resp.Requests, 1)
	require.Equal(t, dealerID, resp.Requests[0].DealerID)
	require.Equal(t, "promo", resp.Requests[0].Type)
	require.Equal(t, 1234.5, resp.Requests[0].Amount)
	require.Equal(t, "pending", resp.Requests[0].Status)
	require.Equal(t, 1, resp.Pending)
}

func TestKPIService_GetDealerRequests_RealErrorPropagates(t *testing.T) {
	db := newTestSQLiteDB(t)
	require.NoError(t, db.Exec("CREATE TABLE dealer_requests (id TEXT PRIMARY KEY)").Error)

	svc := NewKPIService(db, nil, nil)
	_, err := svc.GetDealerRequests(context.Background(), uuid.New(), "")
	require.Error(t, err)
}

func TestKPIService_GetDealerMarketingBudget_NoRowReturnsZeros(t *testing.T) {
	db := newTestSQLiteDB(t)
	require.NoError(t, db.Exec(`
		CREATE TABLE marketing_budgets (
			id TEXT PRIMARY KEY, dealer_id TEXT, quarter TEXT,
			total_amount REAL, used_amount REAL)`).Error)

	svc := NewKPIService(db, nil, nil)
	resp, err := svc.GetDealerMarketingBudget(context.Background(), uuid.New(), "Q1-2026")
	require.NoError(t, err)
	require.Equal(t, 0.0, resp.TotalAmount)
	require.Equal(t, 0.0, resp.UsedAmount)
	require.Equal(t, 0.0, resp.Remaining)
	require.Equal(t, 0, resp.UsagePercent)
}

func TestKPIService_GetDealerMarketingBudget_ReadsRow(t *testing.T) {
	db := newTestSQLiteDB(t)
	require.NoError(t, db.Exec(`
		CREATE TABLE marketing_budgets (
			id TEXT PRIMARY KEY, dealer_id TEXT, quarter TEXT,
			total_amount REAL, used_amount REAL)`).Error)

	dealerID := uuid.New()
	require.NoError(t, db.Exec(
		`INSERT INTO marketing_budgets (id, dealer_id, quarter, total_amount, used_amount) VALUES (?, ?, ?, ?, ?)`,
		uuid.New().String(), dealerID.String(), "Q1-2026", 500000, 100000).Error)

	svc := NewKPIService(db, nil, nil)
	resp, err := svc.GetDealerMarketingBudget(context.Background(), dealerID, "Q1-2026")
	require.NoError(t, err)
	require.Equal(t, 500000.0, resp.TotalAmount)
	require.Equal(t, 100000.0, resp.UsedAmount)
	require.Equal(t, 20, resp.UsagePercent)
}

func TestKPIService_GetDealerMarketingBudget_MissingTableReturnsError(t *testing.T) {
	db := newTestSQLiteDB(t)

	svc := NewKPIService(db, nil, nil)
	_, err := svc.GetDealerMarketingBudget(context.Background(), uuid.New(), "Q1-2026")
	require.Error(t, err)
}

func setupUsersTableForScopeTest(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE users (
		id TEXT PRIMARY KEY, tenant_id TEXT, role TEXT, managed_by TEXT,
		first_name TEXT, last_name TEXT, salon_id TEXT, deleted_at DATETIME)`).Error)
}

func setupDealerRequestsTableForScopeTest(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE dealer_requests (
		id TEXT PRIMARY KEY, dealer_id TEXT, type TEXT, description TEXT,
		amount REAL, status TEXT, created_at DATETIME, updated_at DATETIME)`).Error)
}

func TestKPIService_GetManagerDealers_CrossTenantForbidden(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupUsersTableForScopeTest(t, db)

	caller, tenantA, tenantB, mgrB := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		caller.String(), tenantA.String(), string(models.RoleFranchisor)).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		mgrB.String(), tenantB.String(), string(models.RoleFranchisorManager)).Error)

	svc := NewKPIService(db, nil, nil)
	_, err := svc.GetManagerDealers(context.Background(), caller.String(), mgrB.String())
	require.Error(t, err)
	require.Contains(t, err.Error(), "forbidden")
}

func TestKPIService_GetManagerDealers_SameTenantFranchiserAllowed(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupUsersTableForScopeTest(t, db)

	caller, tenantA, mgrA := uuid.New(), uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		caller.String(), tenantA.String(), string(models.RoleFranchisor)).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		mgrA.String(), tenantA.String(), string(models.RoleFranchisorManager)).Error)

	svc := NewKPIService(db, nil, nil)
	resp, err := svc.GetManagerDealers(context.Background(), caller.String(), mgrA.String())
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestKPIService_GetManagerDealers_PeerDealerDenied(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupUsersTableForScopeTest(t, db)

	// Дилер из той же сети запрашивает чужого менеджера — не руководитель, не franchiser.
	peer, tenantA, mgrA := uuid.New(), uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		peer.String(), tenantA.String(), string(models.RoleDealer)).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		mgrA.String(), tenantA.String(), string(models.RoleFranchisorManager)).Error)

	svc := NewKPIService(db, nil, nil)
	_, err := svc.GetManagerDealers(context.Background(), peer.String(), mgrA.String())
	require.Error(t, err)
	require.Contains(t, err.Error(), "forbidden")
}

func TestKPIService_GetFranchiserRequests_TenantIsolation(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupUsersTableForScopeTest(t, db)
	setupDealerRequestsTableForScopeTest(t, db)
	now := time.Now()

	tenantA, tenantB := uuid.New(), uuid.New()
	callerA, dealerA, dealerB := uuid.New(), uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		callerA.String(), tenantA.String(), string(models.RoleFranchisor)).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role, first_name) VALUES (?, ?, ?, ?)`,
		dealerA.String(), tenantA.String(), string(models.RoleDealer), "DealerA").Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role, first_name) VALUES (?, ?, ?, ?)`,
		dealerB.String(), tenantB.String(), string(models.RoleDealer), "DealerB").Error)
	require.NoError(t, db.Exec(
		`INSERT INTO dealer_requests (id, dealer_id, type, description, amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), dealerA.String(), "promo", "a", 100.0, "pending", now, now).Error)
	require.NoError(t, db.Exec(
		`INSERT INTO dealer_requests (id, dealer_id, type, description, amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), dealerB.String(), "promo", "b", 200.0, "pending", now, now).Error)

	svc := NewKPIService(db, nil, nil)
	resp, err := svc.GetFranchiserRequests(context.Background(), callerA, "all")
	require.NoError(t, err)
	require.Len(t, resp.Requests, 1, "franchiser сети A не должен видеть запросы сети B")
	require.Equal(t, dealerA, resp.Requests[0].DealerID)
}

func setupNotificationsTableForAssignTest(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE notifications (
		id TEXT PRIMARY KEY, tenant_id TEXT, user_id TEXT, type TEXT,
		title TEXT, message TEXT, is_read INTEGER, data TEXT, created_at DATETIME)`).Error)
}

func TestKPIService_AssignAlert_SameTenantAssigns(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupUsersTableForScopeTest(t, db)
	setupNotificationsTableForAssignTest(t, db)

	tenantA := uuid.New()
	caller, target, alert := uuid.New(), uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		caller.String(), tenantA.String(), string(models.RoleFranchisor)).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		target.String(), tenantA.String(), string(models.RoleDealer)).Error)
	require.NoError(t, db.Exec(`INSERT INTO notifications (id, tenant_id, type, title) VALUES (?, ?, ?, ?)`,
		alert.String(), tenantA.String(), "info", "t").Error)

	svc := NewKPIService(db, nil, nil)
	require.NoError(t, svc.AssignAlert(context.Background(), caller, alert, target))

	var got string
	require.NoError(t, db.Table("notifications").Select("user_id").Where("id = ?", alert.String()).Scan(&got).Error)
	require.Equal(t, target.String(), got)
}

func TestKPIService_AssignAlert_CrossTenantAlertForbidden(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupUsersTableForScopeTest(t, db)
	setupNotificationsTableForAssignTest(t, db)

	tenantA, tenantB := uuid.New(), uuid.New()
	caller, target, alert := uuid.New(), uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		caller.String(), tenantA.String(), string(models.RoleFranchisor)).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		target.String(), tenantA.String(), string(models.RoleDealer)).Error)
	require.NoError(t, db.Exec(`INSERT INTO notifications (id, tenant_id, type, title) VALUES (?, ?, ?, ?)`,
		alert.String(), tenantB.String(), "info", "t").Error)

	svc := NewKPIService(db, nil, nil)
	err := svc.AssignAlert(context.Background(), caller, alert, target)
	require.Error(t, err)
	require.Contains(t, err.Error(), "forbidden")
}

func TestKPIService_AssignAlert_CrossTenantTargetForbidden(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupUsersTableForScopeTest(t, db)
	setupNotificationsTableForAssignTest(t, db)

	tenantA, tenantB := uuid.New(), uuid.New()
	caller, target, alert := uuid.New(), uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		caller.String(), tenantA.String(), string(models.RoleFranchisor)).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		target.String(), tenantB.String(), string(models.RoleDealer)).Error)
	require.NoError(t, db.Exec(`INSERT INTO notifications (id, tenant_id, type, title) VALUES (?, ?, ?, ?)`,
		alert.String(), tenantA.String(), "info", "t").Error)

	svc := NewKPIService(db, nil, nil)
	err := svc.AssignAlert(context.Background(), caller, alert, target)
	require.Error(t, err)
	require.Contains(t, err.Error(), "forbidden")
}

func setupGoalsTableForPlansTest(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE goals (
		id TEXT PRIMARY KEY, assigner_id TEXT, assignee_id TEXT, role TEXT,
		sales_plan REAL, leads_plan INTEGER, calls_plan INTEGER, meetings_plan INTEGER,
		period TEXT, start_date DATETIME, end_date DATETIME, target_date DATETIME,
		tenant_id TEXT, created_at DATETIME, updated_at DATETIME)`).Error)
}

func TestKPIService_SetManagerPlans_CrossTenantSkipped(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupUsersTableForScopeTest(t, db)
	setupGoalsTableForPlansTest(t, db)

	tenantA, tenantB := uuid.New(), uuid.New()
	franchiser := uuid.New()
	mgrA, mgrB := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		franchiser.String(), tenantA.String(), string(models.RoleFranchisor)).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role, managed_by) VALUES (?, ?, ?, ?)`,
		mgrA.String(), tenantA.String(), string(models.RoleFranchisorManager), franchiser.String()).Error)
	// Связь есть, но сеть чужая (повреждённая иерархия): тоже пропуск.
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role, managed_by) VALUES (?, ?, ?, ?)`,
		mgrB.String(), tenantB.String(), string(models.RoleFranchisorManager), franchiser.String()).Error)

	svc := NewKPIService(db, nil, nil)
	plans := []struct {
		ManagerID     string  `json:"manager_id"`
		PlanAmount    float64 `json:"plan_amount"`
		TargetDealers int     `json:"target_dealers"`
	}{
		{ManagerID: mgrA.String(), PlanAmount: 1000, TargetDealers: 5},
		{ManagerID: mgrB.String(), PlanAmount: 2000, TargetDealers: 5},
	}
	err := svc.SetManagerPlans(context.Background(), franchiser.String(), "2026-Q1", plans)
	require.Error(t, err, "частичный батч — не молчаливый успех")
	require.Contains(t, err.Error(), "skipped 1")

	// REAUDIT-4: батч атомарен — если хоть один план невалиден, не пишется НИЧЕГО
	// (раньше валидная часть применялась, а хендлер отвечал ошибкой).
	var n int64
	require.NoError(t, db.Table("goals").Count(&n).Error)
	require.Equal(t, int64(0), n, "частичное применение запрещено")
}

func TestKPIService_GetFranchiserDealers_TenantIsolation(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupUsersTableForScopeTest(t, db)

	tenantA, tenantB := uuid.New(), uuid.New()
	franchiser := uuid.New()
	dealerA, dealerB := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		franchiser.String(), tenantA.String(), string(models.RoleFranchisor)).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role, managed_by) VALUES (?, ?, ?, ?)`,
		dealerA.String(), tenantA.String(), string(models.RoleDealer), franchiser.String()).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role, managed_by) VALUES (?, ?, ?, ?)`,
		dealerB.String(), tenantB.String(), string(models.RoleDealer), franchiser.String()).Error)

	svc := NewKPIService(db, nil, nil)
	resp, err := svc.GetFranchiserDealers(context.Background(), franchiser, "all")
	require.NoError(t, err)
	require.Len(t, resp.Dealers, 1, "чужой сети не видно даже при битой связи")
	require.Equal(t, dealerA, resp.Dealers[0].ID)
}

// S11: границы суток считаются в зоне тенанта, не в UTC/зоне сервера.
func TestResolveLocation_ValidAndInvalid(t *testing.T) {
	require.Equal(t, "Asia/Almaty", repository.ResolveLocation("Asia/Almaty").String())
	require.Equal(t, "UTC", repository.ResolveLocation("Not/AZone").String())
	require.Equal(t, "UTC", repository.ResolveLocation("").String())
}

func TestKPIService_NowForUsesTenantTimezone(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupUsersTableForScopeTest(t, db)
	require.NoError(t, db.Exec(`CREATE TABLE tenants (id TEXT PRIMARY KEY, timezone TEXT)`).Error)

	tenantID, userID := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO tenants (id, timezone) VALUES (?, ?)`,
		tenantID.String(), "Asia/Almaty").Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		userID.String(), tenantID.String(), string(models.RoleDealer)).Error)

	svc := NewKPIService(db, nil, nil)
	now := svc.nowFor(context.Background(), userID)
	require.Equal(t, "Asia/Almaty", now.Location().String())

	utcNow := time.Now().UTC()
	assert.WithinDuration(t, utcNow, now, 5*time.Second, "момент времени тот же, только зона тенанта")
}
