package database

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		viper.GetString("db_host"),
		viper.GetString("db_user"),
		viper.GetString("db_password"),
		viper.GetString("db_name"),
		viper.GetString("db_port"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	log.Println("Database connected")

	if err := runMigrations(DB); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database migrated")

	if err := SeedUsers(DB); err != nil {
		log.Printf("Warning: SeedUsers failed: %v", err)
	}

	return DB, nil
}

func runMigrations(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	migrations := []func(*gorm.DB) error{
		migrateUsers,
		migrateTenants,
		migrateSalons,
		migrateOrders,
		migrateTasks,
		migrateChecklists,
		migratePlans,
		migrateNotifications,
		migrateInvoices,
		migrateLeads,
		migrateLeadActivities,
		migrateChecklistTemplates,
		migrateChecklistTemplateItems,
		migrateAssignedChecklists,
		migrateChecklistResponses,
	}

	for _, m := range migrations {
		if err := m(db); err != nil {
			return err
		}
	}

	_ = sqlDB
	return nil
}

func migrateUsers(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) NOT NULL,
			password_hash TEXT,
			role VARCHAR(50) NOT NULL,
			tenant_id UUID,
			salon_id UUID,
			managed_by UUID,
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			phone VARCHAR(50),
			deleted_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id UUID`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS salon_id UUID`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS managed_by UUID`)

	return nil
}

func migrateTenants(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS tenants (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			status VARCHAR(50) DEFAULT 'active',
			plan_id UUID,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

func migrateSalons(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS salons (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			name VARCHAR(255) NOT NULL,
			address TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

func migrateOrders(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			tenant_id UUID,
			salon_id UUID,
			total DECIMAL(10,2) DEFAULT 0,
			status VARCHAR(50) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

func migrateTasks(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID,
			tenant_id UUID,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(50) DEFAULT 'pending',
			due_date TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

func migrateChecklists(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS checklists (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID,
			tenant_id UUID,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(50) DEFAULT 'pending',
			priority VARCHAR(50) DEFAULT 'normal',
			assigned_to UUID,
			start_date TIMESTAMP,
			end_date TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

func migratePlans(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS plans (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			owner_id UUID,
			name VARCHAR(255) NOT NULL,
			price DECIMAL(10,2) DEFAULT 0,
			max_salons INT DEFAULT 1,
			max_users INT DEFAULT 5,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

func migrateNotifications(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS notifications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			type VARCHAR(50) DEFAULT 'info',
			title VARCHAR(255) NOT NULL,
			message TEXT,
			is_read BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

func migrateInvoices(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS invoices (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			amount DECIMAL(10,2) DEFAULT 0,
			status VARCHAR(50) DEFAULT 'pending',
			due_date TIMESTAMP,
			paid_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

func migrateLeads(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS leads (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			salon_id UUID,
			manager_id UUID,
			full_name VARCHAR(255) NOT NULL,
			phone VARCHAR(50),
			email VARCHAR(255),
			interest_product TEXT,
			budget DECIMAL(12,2),
			status VARCHAR(50) DEFAULT 'new',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

func migrateLeadActivities(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS lead_activities (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			lead_id UUID NOT NULL,
			user_id UUID,
			type VARCHAR(50) DEFAULT 'note',
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

func migrateChecklistTemplates(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS checklist_templates (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

func migrateChecklistTemplateItems(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS checklist_template_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			template_id UUID NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			order_num INT DEFAULT 0
		)
	`).Error
}

func migrateAssignedChecklists(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS assigned_checklists (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			template_id UUID,
			status VARCHAR(50) DEFAULT 'pending',
			start_date TIMESTAMP,
			end_date TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

func migrateChecklistResponses(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS checklist_responses (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			assigned_checklist_id UUID NOT NULL,
			item_id UUID NOT NULL,
			user_id UUID,
			is_completed BOOLEAN DEFAULT FALSE,
			response_text TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

func GetDB() *gorm.DB {
	return DB
}