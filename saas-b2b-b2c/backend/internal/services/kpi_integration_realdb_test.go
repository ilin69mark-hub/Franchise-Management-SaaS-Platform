package services

import (
	"context"
	"os"
	"testing"
	"time"

	"franchise-saas-backend/internal/repository/testdb"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// testDBConn возвращает соединение с реальным PostgreSQL для интеграционных
// realdb-тестов. Если TEST_DB_DSN не задан — тест пропускается (CI без БД
// остаётся зелёным), иначе выполняется на живой БД (docker-compose.test.yml,
// порт 5433, база franchise_test).
func testDBConn(t *testing.T) *gorm.DB {
	t.Helper()
	if os.Getenv("TEST_DB_DSN") == "" {
		t.Skip("TEST_DB_DSN не задан: интеграционный realdb-тест пропущен")
	}
	db, err := testdb.Connect()
	if err != nil {
		t.Fatalf("realdb: не удалось подключиться: %v", err)
	}
	return db
}

// setupRealDBSchema создаёт PG-совместимую схему аналитики (TIMESTAMP вместо
// DATETIME, DOUBLE PRECISION вместо REAL — иначе unit-схема из
// kpi_products_test.go не работает на живом PostgreSQL) и очищает таблицы,
// чтобы прогон был воспроизводим на уже использованной тестовой БД.
func setupRealDBSchema(t *testing.T, db *gorm.DB) {
	t.Helper()
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY, email TEXT, password_hash TEXT, role TEXT,
			salon_id TEXT, managed_by TEXT, created_at TIMESTAMP, deleted_at TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS leads (
			id TEXT PRIMARY KEY, salon_id TEXT, manager_id TEXT, full_name TEXT,
			interest_product TEXT, budget DOUBLE PRECISION, status TEXT,
			created_at TIMESTAMP, updated_at TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS products (
			id TEXT PRIMARY KEY, salon_id TEXT, name TEXT, collection TEXT,
			category TEXT, price DOUBLE PRECISION, cost_price DOUBLE PRECISION,
			showroom_qty INT, warehouse_qty INT, turnover_days INT)`,
		`CREATE TABLE IF NOT EXISTS lost_sales (
			id TEXT PRIMARY KEY, salon_id TEXT, reason TEXT, requests_count INT,
			lost_revenue DOUBLE PRECISION, period TEXT)`,
		`CREATE TABLE IF NOT EXISTS promotions (
			id TEXT PRIMARY KEY, salon_id TEXT, name TEXT, condition TEXT,
			discount_min INT, discount_max INT, start_date TIMESTAMP,
			end_date TIMESTAMP, is_active BOOLEAN)`,
		`CREATE TABLE IF NOT EXISTS category_turnover (
			id TEXT PRIMARY KEY, salon_id TEXT, category TEXT, avg_days INT,
			period TEXT)`,
		`CREATE TABLE IF NOT EXISTS goals (
			id TEXT PRIMARY KEY, assignee_id TEXT, sales_plan DOUBLE PRECISION,
			target_date TIMESTAMP, period TEXT, role TEXT, status TEXT)`,
		`CREATE TABLE IF NOT EXISTS system_settings (
			key TEXT PRIMARY KEY, value TEXT, description TEXT, updated_at TIMESTAMP)`,
	}
	for _, tbl := range []string{
		"users", "leads", "products", "lost_sales", "promotions",
		"category_turnover", "goals", "system_settings",
	} {
		stmts = append(stmts, `DELETE FROM `+tbl)
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error, "setupRealDBSchema: %s", stmt)
	}
}

// TestIntegrationDashboardProductsRealDB — realdb-аналог
// TestKPIService_GetDashboardProducts_ReturnsRealData на живом PostgreSQL:
// сверяет TotalRevenue=170000, маржу 41.67% и остатки против реальной БД.
func TestIntegrationDashboardProductsRealDB(t *testing.T) {
	db := testDBConn(t)
	setupRealDBSchema(t, db)

	salonID := uuid.New()
	userID := uuid.New()
	insertTestUser(t, db, userID, "salon_manager", &salonID)

	today := time.Now()
	period := today.Format("2006-01")

	insertSalonProduct(t, db, salonID, "Диван", "Основная", "Мебель", 120000, 70000, 2, 6, 35)
	insertLeadWithProduct(t, db, salonID, "Диван", 120000, "sale", today)
	insertLeadWithProduct(t, db, salonID, "Что-то экзотическое", 50000, "paid", today)

	require.NoError(t, db.Exec(
		"INSERT INTO lost_sales (id, salon_id, reason, requests_count, lost_revenue, period) VALUES (?, ?, ?, ?, ?, ?)",
		uuid.New().String(), salonID.String(), "Нет в наличии", 3, 150000, period).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO category_turnover (id, salon_id, category, avg_days, period) VALUES (?, ?, ?, ?, ?)",
		uuid.New().String(), salonID.String(), "Мебель", 120, period).Error)

	svc := NewKPIService(db, nil, nil)
	resp, err := svc.GetDashboardProducts(context.Background(), userID, "")
	require.NoError(t, err)
	require.Equal(t, 170000.0, resp.TotalRevenue)

	byName := map[string]struct {
		Category string
		Margin   float64
	}{}
	for _, tp := range resp.TopProducts {
		byName[tp.Name] = struct {
			Category string
			Margin   float64
		}{tp.Category, tp.Margin}
	}
	divan, ok := byName["Диван"]
	require.True(t, ok, "realdb: товар «Диван» не найден в TopProducts: %+v", resp.TopProducts)
	require.Equal(t, "Мебель", divan.Category)
	require.InDelta(t, 41.67, divan.Margin, 0.1)

	require.Len(t, resp.StockItems, 1)
	require.Equal(t, 8*70000.0, resp.StockItems[0].TotalCost)

	require.Len(t, resp.LostSales, 1)
	require.Equal(t, 150000.0, resp.LostSales[0].LostRevenue)

	require.Len(t, resp.CategoryTurnover, 1)
	require.Equal(t, 120, resp.CategoryTurnover[0].AvgDays)
	require.True(t, resp.CategoryTurnover[0].IsSlowMoving)
}

// TestIntegrationManagerTargetsRealDB — realdb-аналог
// TestKPIService_GetManagerTargets_RealBenchmarksAndCategories на живом
// PostgreSQL: план 1 000 000, конверсия 66.67%, MaxBonus 60000 из
// system_settings, BonusForecast 12000.
func TestIntegrationManagerTargetsRealDB(t *testing.T) {
	db := testDBConn(t)
	setupRealDBSchema(t, db)

	salonID := uuid.New()
	userID := uuid.New()
	insertTestUser(t, db, userID, "salon_manager", &salonID)

	today := time.Now()
	firstOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())

	require.NoError(t, db.Exec(
		"INSERT INTO goals (id, assignee_id, sales_plan, target_date, role, status) VALUES (?, ?, ?, ?, 'salon_manager', 'active')",
		uuid.New().String(), userID.String(), 1000000, today).Error)

	insertSalonProduct(t, db, salonID, "Диван", "Основная", "Мебель", 120000, 70000, 1, 1, 30)
	insertSalonProduct(t, db, salonID, "Подушка декоративная", "Декор", "Допы", 3500, 1500, 5, 20, 18)
	insertLeadWithProduct(t, db, salonID, "Диван", 200000, "sale", today)
	insertLeadWithProduct(t, db, salonID, "Подушка декоративная", 50000, "sale", today)
	insertLeadWithProduct(t, db, salonID, "Диван", 0, "new", today)

	require.NoError(t, db.Exec(
		"INSERT INTO system_settings (key, value) VALUES ('target_conversion', '25'), ('target_extras_percent', '20'), ('max_bonus', '60000')").Error)

	endSoon := today.AddDate(0, 0, 3)
	require.NoError(t, db.Exec(
		"INSERT INTO promotions (id, salon_id, name, condition, discount_min, discount_max, start_date, end_date, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		uuid.New().String(), salonID.String(), "Летняя распродажа", "Скидка", 10, 25, firstOfMonth, endSoon, true).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO promotions (id, salon_id, name, condition, discount_min, discount_max, start_date, end_date, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		uuid.New().String(), salonID.String(), "Неактивная", "Скидка", 5, 10, firstOfMonth, endSoon, false).Error)

	svc := NewKPIService(db, nil, nil)
	resp, err := svc.GetManagerTargets(context.Background(), userID, today.Format("2006-01-02"))
	require.NoError(t, err)
	require.True(t, resp.HasTargets, "realdb: план не утверждён на живой БД")
	require.Equal(t, 1000000.0, resp.Plan.TotalAmount)
	require.Equal(t, 250000.0, resp.Plan.CurrentAmount)
	require.Equal(t, 25, resp.Plan.Percent)

	require.InDelta(t, 200000.0, resp.Plan.ByCategory["Мебель"], 0.01)
	require.InDelta(t, 50000.0, resp.Plan.ByCategory["Допы"], 0.01)
	require.Len(t, resp.Plan.ByCategory, 2)

	require.Equal(t, 25.0, resp.TargetConversion)
	require.Equal(t, 20.0, resp.TargetExtrasPercent)
	require.Equal(t, 60000.0, resp.MaxBonus)

	require.InDelta(t, 66.67, resp.CurrentConversion, 0.1)
	require.InDelta(t, 20.0, resp.CurrentExtrasPercent, 0.1)

	require.Len(t, resp.Promotions, 1)
	require.Equal(t, "Летняя распродажа", resp.Promotions[0].Name)
	require.True(t, resp.Promotions[0].IsExpiring)

	require.InDelta(t, 60000.0*0.2, resp.BonusForecast, 0.01)
}

// TestIntegrationDealerProductsRealDB — realdb-прогон GetDealerProducts на
// живом PostgreSQL: дилер видит салоны своих менеджеров (managed_by), выручка
// 150000 из лидов sale+paid текущего месяца.
func TestIntegrationDealerProductsRealDB(t *testing.T) {
	db := testDBConn(t)
	setupRealDBSchema(t, db)

	dealerID := uuid.New()
	managerID := uuid.New()
	salonID := uuid.New()

	require.NoError(t, db.Exec(
		"INSERT INTO users (id, email, password_hash, role, salon_id) VALUES (?, ?, '', 'dealer', NULL)",
		dealerID.String(), dealerID.String()+"@test.ru").Error)
	require.NoError(t, db.Exec(
		"INSERT INTO users (id, email, password_hash, role, salon_id, managed_by) VALUES (?, ?, '', 'salon_manager', ?, ?)",
		managerID.String(), managerID.String()+"@test.ru", salonID.String(), dealerID.String()).Error)

	today := time.Now()
	insertLeadWithProduct(t, db, salonID, "Диван", 100000, "sale", today)
	insertLeadWithProduct(t, db, salonID, "Кресло", 50000, "paid", today)
	insertLeadWithProduct(t, db, salonID, "Стол", 77777, "new", today)

	svc := NewKPIService(db, nil, nil)
	// dateStr="" → targetDate=time.Now(), как в approved unit-тестах: сегодняшние
	// лиды (созданные после 00:00) попадают в BETWEEN firstOfMonth..now.
	resp, err := svc.GetDealerProducts(context.Background(), dealerID, "")
	require.NoError(t, err)
	require.Equal(t, 150000.0, resp.TotalRevenue, "realdb: выручка дилера не сошлась (new-лид не должен считаться)")
	require.Len(t, resp.TopProducts, 2)
}

// TestIntegrationRealDBSchemaSetupIdempotent — доказывает, что realdb-прогон
// идемпотентен: повторный setup не падает на живой уже использованной БД и
// очищает данные, а количество строк после вставки детерминировано.
func TestIntegrationRealDBSchemaSetupIdempotent(t *testing.T) {
	db := testDBConn(t)
	setupRealDBSchema(t, db)
	setupRealDBSchema(t, db)

	salonID := uuid.New()
	insertSalonProduct(t, db, salonID, "Диван", "Основная", "Мебель", 120000, 70000, 2, 6, 35)

	var n int64
	require.NoError(t, db.Raw("SELECT COUNT(*) FROM products").Scan(&n).Error)
	require.Equal(t, int64(1), n, "realdb: после setup+insert в products ровно 1 строка")

	setupRealDBSchema(t, db)
	require.NoError(t, db.Raw("SELECT COUNT(*) FROM products").Scan(&n).Error)
	require.Equal(t, int64(0), n, "realdb: повторный setup очищает products")
}
