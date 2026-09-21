package services

import (
	"context"
	"testing"
	"time"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
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

func TestKPIService_GetDealerMarketingBudget_DefaultsWhenNoRow(t *testing.T) {
	db := newTestSQLiteDB(t)
	require.NoError(t, db.Exec(`
		CREATE TABLE marketing_budgets (
			id TEXT PRIMARY KEY, dealer_id TEXT, quarter TEXT,
			total_amount REAL, used_amount REAL)`).Error)

	svc := NewKPIService(db, nil, nil)
	resp, err := svc.GetDealerMarketingBudget(context.Background(), uuid.New(), "Q1-2026")
	require.NoError(t, err)
	require.Equal(t, 200000.0, resp.TotalAmount)
	require.Equal(t, 60000.0, resp.UsedAmount)
	require.Equal(t, 140000.0, resp.Remaining)
	require.Equal(t, 30, resp.UsagePercent)
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

func TestKPIService_GetDealerMarketingBudget_MissingTableDefaults(t *testing.T) {
	db := newTestSQLiteDB(t)

	svc := NewKPIService(db, nil, nil)
	resp, err := svc.GetDealerMarketingBudget(context.Background(), uuid.New(), "Q1-2026")
	require.NoError(t, err)
	require.Equal(t, 200000.0, resp.TotalAmount)
	require.Equal(t, 60000.0, resp.UsedAmount)
}
