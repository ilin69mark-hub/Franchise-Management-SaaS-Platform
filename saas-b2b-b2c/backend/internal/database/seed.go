package database

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// seedPassword возвращает пароль для сид-аккаунтов.
// Берётся из env SEED_PASSWORD; если не задан — генерируется
// случайный пароль, который один раз печатается в stderr.
func seedPassword() string {
	if pwd := os.Getenv("SEED_PASSWORD"); pwd != "" {
		return pwd
	}
	pwd := randomPassword()
	fmt.Fprintf(os.Stderr, "[seed] SEED_PASSWORD not set — generated password for seeded users: %s\n", pwd)
	return pwd
}

func randomPassword() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "P@ssw0rdSeeded!"
	}
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

// tableExists проверяет существование таблицы в текущей схеме.
func tableExists(db *gorm.DB, name string) bool {
	var count int64
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = ?", name).Scan(&count)
	return count > 0
}

func SeedUsers(db *gorm.DB) error {
	password := seedPassword()

	users := []struct {
		Email     string
		Role      string
		FirstName string
		LastName  string
	}{
		{Email: "admin@mail.ru", Role: "super_admin", FirstName: "Super", LastName: "Admin"},
		{Email: "fr@mail.ru", Role: "franchiser", FirstName: "Franchise", LastName: "Owner"},
		{Email: "manager1@1.ru", Role: "franchiser_manager", FirstName: "Александр", LastName: "Петров"},
		{Email: "dealer1@1.ru", Role: "dealer", FirstName: "Иван", LastName: "Смирнов"},
		{Email: "salon1@1.ru", Role: "salon_manager", FirstName: "Екатерина", LastName: "Сидорова"},
	}

	for _, u := range users {
		var count int64
		db.Table("users").Where("email = ?", u.Email).Count(&count)
		if count > 0 {
			log.Printf("User %s already exists, skipping", u.Email)
			continue
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		if u.Role == "franchiser" {
			var tenantCount int64
			db.Table("tenants").Count(&tenantCount)
			tenantUUID := "00000000-0000-0000-0000-000000000001"
			db.Exec(`INSERT INTO tenants (id, name, status) VALUES ($1, $2, 'active') ON CONFLICT DO NOTHING`,
				tenantUUID, u.FirstName+" "+u.LastName)
			db.Exec(`INSERT INTO users (id, email, password_hash, role, tenant_id, first_name, last_name)
				VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)`,
				u.Email, string(hashed), u.Role, tenantUUID, u.FirstName, u.LastName)
		} else {
			db.Exec(`INSERT INTO users (id, email, password_hash, role, tenant_id, first_name, last_name)
				VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)`,
				u.Email, string(hashed), u.Role, "00000000-0000-0000-0000-000000000001", u.FirstName, u.LastName)
		}

		log.Printf("User created: %s (%s)", u.Email, u.Role)
	}

	return nil
}

func SeedGoals(db *gorm.DB) error {
	today := time.Now().Format("2006-01-02")

	goals := []struct {
		SalesPlan  float64
		SalesFact  float64
		Role       string
		TargetDate string
	}{
		{SalesPlan: 15000000, SalesFact: 12000000, Role: "dealer", TargetDate: today},
		{SalesPlan: 8000000, SalesFact: 6500000, Role: "dealer", TargetDate: today},
		{SalesPlan: 1000000, SalesFact: 800000, Role: "salon_manager", TargetDate: today},
		{SalesPlan: 15000000, SalesFact: 12500000, Role: "franchiser_manager", TargetDate: today},
	}

	for _, g := range goals {
		var userID string
		db.Table("users").Where("role = ?", g.Role).Order("created_at").Limit(1).Select("id").Scan(&userID)
		if userID == "" {
			log.Printf("No user found for role %s, skipping goal", g.Role)
			continue
		}

		var count int64
		db.Table("goals").Where("assignee_id = ? AND target_date = ?", userID, g.TargetDate).Count(&count)
		if count > 0 {
			log.Printf("Goal already exists for %s on %s, skipping", g.Role, g.TargetDate)
			continue
		}

		// assigner_id подставляется как NULL: в фиксированной схеме
		// колонка nullable, сеялка не назначает цели от чужого лица.
		db.Exec(`INSERT INTO goals (id, assigner_id, assignee_id, sales_plan, sales_fact, role, target_date, status)
			VALUES (gen_random_uuid(), NULL, $1, $2, $3, $4, $5, 'active')`,
			userID, g.SalesPlan, g.SalesFact, g.Role, g.TargetDate)
		log.Printf("Goal created: %s plan=%v fact=%v", g.Role, g.SalesPlan, g.SalesFact)
	}

	return nil
}

func SeedChecklists(db *gorm.DB) error {
	if !tableExists(db, "checklists") {
		log.Printf("Warning: table 'checklists' does not exist, skipping SeedChecklists")
		return nil
	}

	checklists := []struct {
		Title       string
		Description string
		Role        string
	}{
		{"Проверить выкладку", "Проверить соответствие планограмме", "dealer"},
		{"Оформить витрину", "Оформить новую коллекцию", "dealer"},
		{"Проверить остатки", "Сверить остатки с базой", "salon_manager"},
		{"Отчёт по продажам", "Подготовить еженедельный отчёт", "dealer"},
	}

	for _, c := range checklists {
		var userID string
		db.Table("users").Where("role = ?", c.Role).Order("created_at").Limit(1).Select("id").Scan(&userID)
		if userID == "" {
			continue
		}

		db.Exec(`INSERT INTO checklists (id, user_id, assigned_to, title, description, status, created_at, updated_at)
			VALUES (gen_random_uuid(), $1, $1, $2, $3, 'pending', NOW(), NOW())`,
			userID, c.Title, c.Description)
		log.Printf("Checklist created: %s", c.Title)
	}

	return nil
}

func SeedAlerts(db *gorm.DB) error {
	if !tableExists(db, "alerts") {
		log.Printf("Warning: table 'alerts' does not exist, skipping SeedAlerts")
		return nil
	}

	alerts := []struct {
		Category    string
		Priority    string
		Title       string
		Description string
	}{
		{Category: "plan", Priority: "critical", Title: "План выполнен на 65%", Description: "Дилер Москва: план 15М, факт 9.7М"},
		{Category: "funnel", Priority: "warning", Title: "Падение конверсии", Description: "Конверсия ниже нормы на 15%"},
		{Category: "activity", Priority: "warning", Title: "Неактивный дилер", Description: "Не входил в систему 5 дней"},
		{Category: "stock", Priority: "critical", Title: "Дефицит товара", Description: "Кровать Прима - 0 шт."},
	}

	for _, a := range alerts {
		var userID string
		db.Table("users").Where("role = ?", "franchiser_manager").Limit(1).Select("id").Scan(&userID)
		if userID == "" {
			continue
		}

		db.Exec(`INSERT INTO alerts (id, user_id, category, priority, title, description, status, created_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, 'new', NOW())`,
			userID, a.Category, a.Priority, a.Title, a.Description)
		log.Printf("Alert created: %s", a.Title)
	}

	return nil
}

// SeedProductsAnalytics наполняет демо-данными продукты, упущенные
// продажи, акции и оборачиваемость по категориям. Данные привязаны
// к первому салону в базе (роль salon_manager), чтобы дашборды
// GetDashboardProducts / GetManagerTargets показывали живые значения,
// а не кодовые заглушки. Идемпотентно: по (salon_id, name).
func SeedProductsAnalytics(db *gorm.DB) error {
	if !tableExists(db, "products") {
		log.Printf("Warning: table 'products' does not exist, skipping SeedProductsAnalytics")
		return nil
	}

	// Первый салон в базе (создаётся GORM-миграцией/сидом салонов).
	var salonID string
	db.Table("salons").Order("created_at").Limit(1).Select("id").Scan(&salonID)
	if salonID == "" {
		log.Printf("Warning: no salons found, skipping SeedProductsAnalytics")
		return nil
	}

	products := []struct {
		Name       string
		Collection string
		Category   string
		Price      float64
		Cost       float64
		Showroom   int
		Warehouse  int
		Turnover   int
	}{
		{"Диван", "Основная", "Мебель", 120000, 70000, 2, 6, 35},
		{"Кресло", "Основная", "Мебель", 55000, 30000, 4, 12, 28},
		{"Кровать", "Спальня", "Мебель", 180000, 110000, 1, 4, 45},
		{"Шкаф", "Спальня", "Мебель", 150000, 90000, 2, 3, 120},
		{"Стол обеденный", "Кухня", "Мебель", 95000, 56000, 3, 8, 40},
		{"Тумба", "Кухня", "Мебель", 42000, 24000, 5, 10, 22},
		{"Светильник", "Декор", "Допы", 15000, 8000, 6, 20, 30},
		{"Подушка декоративная", "Декор", "Допы", 3500, 1500, 10, 40, 18},
		{"Доставка и подъём", "Услуги", "Услуги", 5000, 0, 0, 0, 0},
		{"Сборка мебели", "Услуги", "Услуги", 3000, 0, 0, 0, 0},
	}

	for _, p := range products {
		var count int64
		db.Table("products").Where("salon_id = ? AND name = ?", salonID, p.Name).Count(&count)
		if count > 0 {
			continue
		}
		db.Exec(`INSERT INTO products (id, salon_id, name, collection, category, price, cost_price, showroom_qty, warehouse_qty, turnover_days)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			salonID, p.Name, p.Collection, p.Category, p.Price, p.Cost, p.Showroom, p.Warehouse, p.Turnover)
		log.Printf("Product created: %s", p.Name)
	}

	// Упущенные продажи за текущий месяц
	if tableExists(db, "lost_sales") {
		period := time.Now().Format("2006-01")
		lost := []struct {
			Reason  string
			Count   int
			Revenue float64
		}{
			{"Нет в наличии", 3, 150000},
			{"Долгий срок производства", 2, 80000},
			{"Не устроила цена", 4, 200000},
			{"Не подошёл дизайн", 2, 120000},
		}
		for _, l := range lost {
			var count int64
			db.Table("lost_sales").Where("salon_id = ? AND reason = ? AND period = ?", salonID, l.Reason, period).Count(&count)
			if count > 0 {
				continue
			}
			db.Exec(`INSERT INTO lost_sales (id, salon_id, reason, requests_count, lost_revenue, period)
				VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)`,
				salonID, l.Reason, l.Count, l.Revenue, period)
			log.Printf("Lost sale created: %s", l.Reason)
		}
	}

	// Акции (для директив менеджера)
	if tableExists(db, "promotions") {
		now := time.Now()
		promos := []struct {
			Name      string
			Condition string
			Min, Max  int
			EndIn     int
		}{
			{"Летняя распродажа", "При покупке дивана - кресло в подарок", 10, 25, 5},
			{"Комплект со скидкой", "Мебель + услуги дизайнера", 15, 30, 14},
			{"Акция выходного дня", "Скидка 20% в субботу и воскресенье", 20, 20, 3},
		}
		for _, pr := range promos {
			var count int64
			db.Table("promotions").Where("salon_id = ? AND name = ?", salonID, pr.Name).Count(&count)
			if count > 0 {
				continue
			}
			db.Exec(`INSERT INTO promotions (id, salon_id, name, condition, discount_min, discount_max, start_date, end_date, is_active)
				VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, TRUE)`,
				salonID, pr.Name, pr.Condition, pr.Min, pr.Max, now, now.AddDate(0, 0, pr.EndIn))
			log.Printf("Promotion created: %s", pr.Name)
		}
	}

	// Оборачиваемость по категориям
	if tableExists(db, "category_turnover") {
		period := time.Now().Format("2006-01")
		cats := []struct {
			Category string
			Days     int
		}{
			{"Мебель", 90},
			{"Допы", 30},
			{"Услуги", 7},
		}
		for _, c := range cats {
			var count int64
			db.Table("category_turnover").Where("salon_id = ? AND category = ? AND period = ?", salonID, c.Category, period).Count(&count)
			if count > 0 {
				continue
			}
			db.Exec(`INSERT INTO category_turnover (id, salon_id, category, avg_days, period)
				VALUES (gen_random_uuid(), $1, $2, $3, $4)`,
				salonID, c.Category, c.Days, period)
			log.Printf("Category turnover created: %s", c.Category)
		}
	}

	return nil
}
