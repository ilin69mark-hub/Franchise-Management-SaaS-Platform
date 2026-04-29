package database

import (
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedUsers(db *gorm.DB) error {
	users := []struct {
		Email     string
		Password  string
		Role      string
		FirstName string
		LastName  string
	}{
		{Email: "admin@mail.ru", Password: "qwerty", Role: "super_admin", FirstName: "Super", LastName: "Admin"},
		{Email: "fr@mail.ru", Password: "qwerty", Role: "franchiser", FirstName: "Franchise", LastName: "Owner"},
		{Email: "manager1@1.ru", Password: "qwerty", Role: "franchiser_manager", FirstName: "Александр", LastName: "Петров"},
		{Email: "dealer1@1.ru", Password: "qwerty", Role: "dealer", FirstName: "Иван", LastName: "Смирнов"},
		{Email: "salon1@1.ru", Password: "qwerty", Role: "salon_manager", FirstName: "Екатерина", LastName: "Сидорова"},
	}

	for _, u := range users {
		var count int64
		db.Table("users").Where("email = ?", u.Email).Count(&count)
		if count > 0 {
			log.Printf("User %s already exists, skipping", u.Email)
			continue
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
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
		SalesPlan    float64
		SalesFact   float64
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

		db.Exec(`INSERT INTO goals (id, assignee_id, sales_plan, sales_fact, role, target_date, status)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, 'active')`,
			userID, g.SalesPlan, g.SalesFact, g.Role, g.TargetDate)
		log.Printf("Goal created: %s plan=%v fact=%v", g.Role, g.SalesPlan, g.SalesFact)
	}

	return nil
}

func SeedChecklists(db *gorm.DB) error {
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

		db.Exec(`INSERT INTO checklists (id, assignee_id, title, description, status, created_at, updated_at)
			VALUES (gen_random_uuid(), $1, $2, $3, 'pending', NOW(), NOW())`,
			userID, c.Title, c.Description)
		log.Printf("Checklist created: %s", c.Title)
	}

	return nil
}

func SeedAlerts(db *gorm.DB) error {
	alerts := []struct {
		Category   string
		Priority   string
		Title      string
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