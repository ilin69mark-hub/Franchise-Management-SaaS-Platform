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
