package services

import (
	"context"
	"testing"

	"franchise-saas-backend/internal/mocks"
	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupScopeTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY, tenant_id TEXT, role TEXT, managed_by TEXT, name TEXT,
		salon_id TEXT, status TEXT, deleted_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS salons (
		id TEXT PRIMARY KEY, name TEXT, tenant_id TEXT)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS dealer_requests (
		id TEXT PRIMARY KEY, dealer_id TEXT, type TEXT, description TEXT, status TEXT, created_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS contracts (
		id TEXT PRIMARY KEY, salon_id TEXT, client_name TEXT, status TEXT, deadline_date DATETIME)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS goals (
		id TEXT PRIMARY KEY, assignee_id TEXT, assigner_id TEXT, role TEXT, sales_plan REAL,
		leads_plan INTEGER, period TEXT, start_date DATETIME, end_date DATETIME, tenant_id TEXT)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS system_settings (key TEXT PRIMARY KEY, value REAL)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS alerts (
		id TEXT PRIMARY KEY, title TEXT, description TEXT, severity TEXT, status TEXT, created_at DATETIME, tenant_id TEXT)`).Error)
}

// healthSegments достаёт сегменты из ответа GetDealersHealth.
func healthSegments(t *testing.T, res map[string]interface{}) []map[string]interface{} {
	t.Helper()
	raw, ok := res["segments"].(map[string]interface{})
	require.True(t, ok, "ожидалась карта сегментов, получено %T", res["segments"])
	out := []map[string]interface{}{}
	for _, v := range raw {
		list, ok := v.([]map[string]interface{})
		require.True(t, ok, "неожиданный тип сегмента %T", v)
		out = append(out, list...)
	}
	return out
}

// C-03: health/migration отдают только дилеров своего tenant.
func TestGetDealersHealth_ScopedToCallerTenant(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupScopeTables(t, db)

	tenantA, tenantB := uuid.New(), uuid.New()
	caller := uuid.New()
	dealerA, dealerB := uuid.New(), uuid.New()

	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		caller.String(), tenantA.String(), string(models.RoleFranchisor)).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		dealerA.String(), tenantA.String(), string(models.RoleDealer)).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		dealerB.String(), tenantB.String(), string(models.RoleDealer)).Error)

	svc := NewKPIService(db, nil, nil)
	res, err := svc.GetDealersHealth(context.Background(), caller.String(), "month")
	require.NoError(t, err)

	seen := map[string]bool{}
	for _, e := range healthSegments(t, res) {
		seen[e["id"].(string)] = true
	}
	require.Contains(t, seen, dealerA.String(), "свой дилер виден")
	require.NotContains(t, seen, dealerB.String(), "чужий tenant не должен попадать в выдачу")
}

// C-03: system-issues не выдаёт алерты и заявки чужого tenant.
func TestGetSystemIssues_ScopedToCallerTenant(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupScopeTables(t, db)

	tenantA, tenantB := uuid.New(), uuid.New()
	caller, dealerA, dealerB := uuid.New(), uuid.New(), uuid.New()

	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		caller.String(), tenantA.String(), string(models.RoleFranchisor)).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		dealerA.String(), tenantA.String(), string(models.RoleDealer)).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, ?, ?)`,
		dealerB.String(), tenantB.String(), string(models.RoleDealer)).Error)
	require.NoError(t, db.Exec(`INSERT INTO alerts (id, title, description, severity, status, created_at, tenant_id) VALUES (?, 'A-alert', 'x', 'high', 'open', CURRENT_TIMESTAMP, ?)`,
		uuid.New().String(), tenantA.String()).Error)
	require.NoError(t, db.Exec(`INSERT INTO alerts (id, title, description, severity, status, created_at, tenant_id) VALUES (?, 'B-SECRET-alert', 'x', 'high', 'open', CURRENT_TIMESTAMP, ?)`,
		uuid.New().String(), tenantB.String()).Error)

	svc := NewKPIService(db, nil, nil)
	issues, err := svc.GetSystemIssues(context.Background(), caller.String(), "")
	require.NoError(t, err)
	for _, issue := range issues {
		require.NotEqual(t, "B-SECRET-alert", issue["title"], "REAUDIT-3: алерты чужого tenant не выдаются")
	}
}

// C-02: tenantless-вызывающий (nil tenant) получает пустую выборку, а не глобальную.
func TestGetDealersHealth_NilTenantFailsClosed(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupScopeTables(t, db)

	dealer := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, tenant_id, role) VALUES (?, NULL, ?)`,
		dealer.String(), string(models.RoleDealer)).Error)

	svc := NewKPIService(db, nil, nil)
	res, err := svc.GetDealersHealth(context.Background(), dealer.String(), "month")
	require.NoError(t, err)
	require.Empty(t, healthSegments(t, res), "nil tenant = fail-closed, не глобальная выдача")
}

// C-05: SendReport не позволяет менять чужой отчёт.
func TestSendReport_RejectsForeignReport(t *testing.T) {
	db := newTestSQLiteDB(t)
	require.NoError(t, db.Exec(`CREATE TABLE reports (
		id TEXT PRIMARY KEY, franchiser_id TEXT, pdf_url TEXT, recipients TEXT,
		created_at DATETIME, updated_at DATETIME)`).Error)

	ownReport, foreignReport := uuid.New(), uuid.New()
	owner, foreignOwner := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO reports (id, franchiser_id, pdf_url, recipients) VALUES (?, ?, '/a.pdf', '["victim@x"]')`,
		ownReport.String(), owner.String()).Error)
	require.NoError(t, db.Exec(`INSERT INTO reports (id, franchiser_id, pdf_url, recipients) VALUES (?, ?, '/b.pdf', '["victim@x"]')`,
		foreignReport.String(), foreignOwner.String()).Error)

	svc := NewKPIService(db, nil, nil)

	require.NoError(t, svc.SendReport(context.Background(), owner.String(), ownReport.String(), []string{"me@x"}))
	require.Error(t, svc.SendReport(context.Background(), owner.String(), foreignReport.String(), []string{"attacker@evil"}),
		"REAUDIT-3: чужой report_id менять нельзя")

	var recipients string
	require.NoError(t, db.Raw("SELECT recipients FROM reports WHERE id = ?", foreignReport.String()).Scan(&recipients).Error)
	require.Contains(t, recipients, "victim@x", "данные чужого отчёта не должны меняться")
}

// C-02: иерархия — дилер не управляет чужим сотрудником той же сети.
func TestCanManage_OnlyHierarchy(t *testing.T) {
	repo := mocks.NewMockUserRepository()
	svc := NewUserServiceWithInterface(repo, nil)

	boss, subordinate, peer := uuid.New(), uuid.New(), uuid.New()
	repo.On("GetUserByID", mock.Anything, boss).Return(&models.User{ID: boss, Role: models.RoleDealer}, nil)

	require.True(t, svc.canManage(boss, &models.User{ID: subordinate, ManagedBy: &boss, Role: models.RoleDealerManager}),
		"прямой подчинённый управляем")
	require.True(t, svc.canManage(boss, &models.User{ID: boss}), "себя тоже можно")
	require.False(t, svc.canManage(boss, &models.User{ID: peer, Role: models.RoleDealer}),
		"REAUDIT-3: равный по роли сотрудник не в иерархии")
	require.False(t, svc.canManage(boss, &models.User{ID: uuid.New(), Role: models.RoleSuperAdmin}),
		"super_admin вне иерархии")
}

// C-02: салон другого tenant нельзя изменить/удалить.
func TestUpdateDeleteSalon_RejectsForeignTenant(t *testing.T) {
	db := newTestSQLiteDB(t)
	require.NoError(t, db.Exec(`CREATE TABLE salons (
		id TEXT PRIMARY KEY, name TEXT, address TEXT, tenant_id TEXT, created_at DATETIME, updated_at DATETIME)`).Error)

	tenantA, tenantB := uuid.New(), uuid.New()
	salonID := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO salons (id, name, tenant_id) VALUES (?, 'victim', ?)`,
		salonID.String(), tenantB.String()).Error)

	svc := NewUserServiceWithInterface(nil, db)

	_, err := svc.UpdateSalon(context.Background(), salonID, tenantA, "hacked", "")
	require.Error(t, err, "REAUDIT-3: чужой салон нельзя изменить")

	err = svc.DeleteSalon(context.Background(), salonID, tenantA)
	require.Error(t, err, "REAUDIT-3: чужой салон нельзя удалить")

	var name string
	require.NoError(t, db.Raw("SELECT name FROM salons WHERE id = ?", salonID.String()).Scan(&name).Error)
	require.Equal(t, "victim", name)
}
