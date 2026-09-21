package services

import (
	"context"
	"testing"
	"time"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupProductsAnalyticsTables создаёт схему, нужную для
// GetDashboardProducts / GetManagerTargets / GetDealerProducts.
func setupProductsAnalyticsTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	mustExec := func(sql string) {
		t.Helper()
		require.NoError(t, db.Exec(sql).Error)
	}
	mustExec("CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT, password_hash TEXT, role TEXT, salon_id TEXT, created_at DATETIME, deleted_at DATETIME)")
	mustExec("CREATE TABLE leads (id TEXT PRIMARY KEY, salon_id TEXT, manager_id TEXT, full_name TEXT, interest_product TEXT, budget REAL, status TEXT, created_at DATETIME, updated_at DATETIME)")
	mustExec("CREATE TABLE products (id TEXT PRIMARY KEY, salon_id TEXT, name TEXT, collection TEXT, category TEXT, price REAL, cost_price REAL, showroom_qty INT, warehouse_qty INT, turnover_days INT)")
	mustExec("CREATE TABLE lost_sales (id TEXT PRIMARY KEY, salon_id TEXT, reason TEXT, requests_count INT, lost_revenue REAL, period TEXT)")
	mustExec("CREATE TABLE promotions (id TEXT PRIMARY KEY, salon_id TEXT, name TEXT, condition TEXT, discount_min INT, discount_max INT, start_date DATETIME, end_date DATETIME, is_active BOOLEAN)")
	mustExec("CREATE TABLE category_turnover (id TEXT PRIMARY KEY, salon_id TEXT, category TEXT, avg_days INT, period TEXT)")
	mustExec("CREATE TABLE goals (id TEXT PRIMARY KEY, assignee_id TEXT, sales_plan REAL, target_date DATETIME, period TEXT, role TEXT, status TEXT)")
	mustExec("CREATE TABLE system_settings (key TEXT PRIMARY KEY, value TEXT, description TEXT, updated_at DATETIME)")
}

func insertTestUser(t *testing.T, db *gorm.DB, id uuid.UUID, role string, salonID *uuid.UUID) {
	t.Helper()
	var salon any
	if salonID != nil {
		salon = salonID.String()
	} else {
		salon = nil
	}
	err := db.Exec("INSERT INTO users (id, email, password_hash, role, salon_id) VALUES (?, ?, '', ?, ?)",
		id.String(), id.String()+"@test.ru", role, salon).Error
	require.NoError(t, err)
}

func insertSalonProduct(t *testing.T, db *gorm.DB, salonID uuid.UUID, name, collection, category string, price, cost float64, showroom, warehouse, turnover int) {
	t.Helper()
	err := db.Exec("INSERT INTO products (id, salon_id, name, collection, category, price, cost_price, showroom_qty, warehouse_qty, turnover_days) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		uuid.New().String(), salonID.String(), name, collection, category, price, cost, showroom, warehouse, turnover).Error
	require.NoError(t, err)
}

func insertLeadWithProduct(t *testing.T, db *gorm.DB, salonID uuid.UUID, product string, budget float64, status string, createdAt time.Time) {
	t.Helper()
	err := db.Exec("INSERT INTO leads (id, salon_id, manager_id, full_name, interest_product, budget, status, created_at, updated_at) VALUES (?, ?, ?, 'test', ?, ?, ?, ?, ?)",
		uuid.New().String(), salonID.String(), uuid.New().String(), product, budget, status, createdAt, createdAt).Error
	require.NoError(t, err)
}

func TestKPIService_GetDashboardProducts_ReturnsRealData(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupProductsAnalyticsTables(t, db)

	salonID := uuid.New()
	userID := uuid.New()
	insertTestUser(t, db, userID, "salon_manager", &salonID)

	today := time.Now()
	period := today.Format("2006-01")

	// Проданный товар ИЗ каталога + товар вне каталога
	insertSalonProduct(t, db, salonID, "Диван", "Основная", "Мебель", 120000, 70000, 2, 6, 35)
	insertLeadWithProduct(t, db, salonID, "Диван", 120000, "sale", today)
	insertLeadWithProduct(t, db, salonID, "Что-то экзотическое", 50000, "paid", today)

	// Упущенные продажи
	require.NoError(t, db.Exec("INSERT INTO lost_sales (id, salon_id, reason, requests_count, lost_revenue, period) VALUES (?, ?, ?, ?, ?, ?)",
		uuid.New().String(), salonID.String(), "Нет в наличии", 3, 150000, period).Error)

	// Оборачиваемость
	require.NoError(t, db.Exec("INSERT INTO category_turnover (id, salon_id, category, avg_days, period) VALUES (?, ?, ?, ?, ?)",
		uuid.New().String(), salonID.String(), "Мебель", 120, period).Error)

	svc := NewKPIService(db, nil, nil)
	resp, err := svc.GetDashboardProducts(context.Background(), userID, "")
	require.NoError(t, err)
	require.Equal(t, 170000.0, resp.TotalRevenue)

	// Топ-товары: известен из каталога (маржа/категория реальные),
	// неизвестный - категория "Прочее", маржа 0
	require.Len(t, resp.TopProducts, 2)
	byName := map[string]models.TopProduct{}
	for _, tp := range resp.TopProducts {
		byName[tp.Name] = tp
	}
	divan, ok := byName["Диван"]
	require.True(t, ok)
	require.Equal(t, "Мебель", divan.Category)
	require.Equal(t, "Основная", divan.Collection)
	require.InDelta(t, 41.67, divan.Margin, 0.1)
	other, ok := byName["Что-то экзотическое"]
	require.True(t, ok)
	require.Equal(t, "Прочее", other.Category)
	require.Equal(t, 0.0, other.Margin)

	// Остатки из каталога
	require.Len(t, resp.StockItems, 1)
	require.Equal(t, "Диван", resp.StockItems[0].Name)
	require.Equal(t, 2, resp.StockItems[0].ShowroomQty)
	require.Equal(t, 6, resp.StockItems[0].WarehouseQty)
	require.Equal(t, 8*70000.0, resp.StockItems[0].TotalCost)

	// Упущенные продажи
	require.Len(t, resp.LostSales, 1)
	require.Equal(t, "Нет в наличии", resp.LostSales[0].Reason)
	require.Equal(t, 3, resp.LostSales[0].RequestsCount)
	require.Equal(t, 150000.0, resp.LostSales[0].LostRevenue)

	// Оборачиваемость
	require.Len(t, resp.CategoryTurnover, 1)
	require.Equal(t, "Мебель", resp.CategoryTurnover[0].Category)
	require.Equal(t, 120, resp.CategoryTurnover[0].AvgDays)
	require.True(t, resp.CategoryTurnover[0].IsSlowMoving)
}

func TestKPIService_GetDashboardProducts_NilSalonID(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupProductsAnalyticsTables(t, db)

	userID := uuid.New()
	insertTestUser(t, db, userID, "super_admin", nil)

	svc := NewKPIService(db, nil, nil)
	_, err := svc.GetDashboardProducts(context.Background(), userID, "")
	require.Error(t, err)
}

func TestKPIService_GetManagerTargets_RealBenchmarksAndCategories(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupProductsAnalyticsTables(t, db)

	salonID := uuid.New()
	userID := uuid.New()
	insertTestUser(t, db, userID, "salon_manager", &salonID)

	today := time.Now()
	firstOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())

	// План 1 000 000
	require.NoError(t, db.Exec("INSERT INTO goals (id, assignee_id, sales_plan, target_date, role, status) VALUES (?, ?, ?, ?, 'salon_manager', 'active')",
		uuid.New().String(), userID.String(), 1000000, today).Error)

	// Продажи: диван (Мебель 200k) + подушка (Допы 50k)
	insertSalonProduct(t, db, salonID, "Диван", "Основная", "Мебель", 120000, 70000, 1, 1, 30)
	insertSalonProduct(t, db, salonID, "Подушка декоративная", "Декор", "Допы", 3500, 1500, 5, 20, 18)
	insertLeadWithProduct(t, db, salonID, "Диван", 200000, "sale", today)
	insertLeadWithProduct(t, db, salonID, "Подушка декоративная", 50000, "sale", today)
	// Непроданный лид - влияет на конверсию
	insertLeadWithProduct(t, db, salonID, "Диван", 0, "new", today)

	// Бенчмарки из настроек
	require.NoError(t, db.Exec("INSERT INTO system_settings (key, value) VALUES ('target_conversion', '25'), ('target_extras_percent', '20'), ('max_bonus', '60000')").Error)

	// Акция
	endSoon := today.AddDate(0, 0, 3)
	require.NoError(t, db.Exec("INSERT INTO promotions (id, salon_id, name, condition, discount_min, discount_max, start_date, end_date, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		uuid.New().String(), salonID.String(), "Летняя распродажа", "Скидка", 10, 25, firstOfMonth, endSoon, true).Error)
	require.NoError(t, db.Exec("INSERT INTO promotions (id, salon_id, name, condition, discount_min, discount_max, start_date, end_date, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		uuid.New().String(), salonID.String(), "Неактивная", "Скидка", 5, 10, firstOfMonth, endSoon, false).Error)

	svc := NewKPIService(db, nil, nil)
	resp, err := svc.GetManagerTargets(context.Background(), userID, today.Format("2006-01-02"))
	require.NoError(t, err)
	require.True(t, resp.HasTargets)
	require.Equal(t, 1000000.0, resp.Plan.TotalAmount)
	require.Equal(t, 250000.0, resp.Plan.CurrentAmount)
	require.Equal(t, 25, resp.Plan.Percent)

	// Категории из реальных продаж, а не множители 0.7/0.2/0.1
	require.InDelta(t, 200000.0, resp.Plan.ByCategory["Мебель"], 0.01)
	require.InDelta(t, 50000.0, resp.Plan.ByCategory["Допы"], 0.01)
	require.Len(t, resp.Plan.ByCategory, 2)

	// Бенчмарки
	require.Equal(t, 25.0, resp.TargetConversion)
	require.Equal(t, 20.0, resp.TargetExtrasPercent)
	require.Equal(t, 60000.0, resp.MaxBonus)

	// Текущая конверсия: 2 из 3 лидов
	require.InDelta(t, 66.67, resp.CurrentConversion, 0.1)
	// Допы = не-мебель = 50k из 250k = 20%
	require.InDelta(t, 20.0, resp.CurrentExtrasPercent, 0.1)

	// Акции: только активная, истекающая <7 дней
	require.Len(t, resp.Promotions, 1)
	require.Equal(t, "Летняя распродажа", resp.Promotions[0].Name)
	require.True(t, resp.Promotions[0].IsExpiring)

	// Прогноз премии 25% → 20% от max = 12000
	require.InDelta(t, 60000.0*0.2, resp.BonusForecast, 0.01)
}

func TestKPIService_GetManagerTargets_NoPlan_ReturnsNoTargets(t *testing.T) {
	db := newTestSQLiteDB(t)
	setupProductsAnalyticsTables(t, db)

	salonID := uuid.New()
	userID := uuid.New()
	insertTestUser(t, db, userID, "salon_manager", &salonID)

	svc := NewKPIService(db, nil, nil)
	resp, err := svc.GetManagerTargets(context.Background(), userID, "")
	require.NoError(t, err)
	require.False(t, resp.HasTargets)
	require.Nil(t, resp.Plan)
	require.Empty(t, resp.Promotions)
}
