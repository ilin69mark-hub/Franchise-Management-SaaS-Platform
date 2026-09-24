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
	sslmode := viper.GetString("db_sslmode")
	if sslmode == "" {
		sslmode = viper.GetString("DB_SSLMODE")
	}
	if sslmode == "" {
		if viper.GetString("GIN_MODE") == "release" {
			sslmode = "require"
		} else {
			sslmode = "disable"
		}
	}
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		viper.GetString("db_host"),
		viper.GetString("db_user"),
		viper.GetString("db_password"),
		viper.GetString("db_name"),
		viper.GetString("db_port"),
		sslmode,
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
		migrateAlerts,
		migrateDealerTasks,
		migrateDealerRequests,
		migrateMarketingBudgets,
		migrateDealerExpenses,
		migrateProducts,
		migrateLostSales,
		migratePromotions,
		migrateCategoryTurnover,
		migrateGoals,
		migrateSystemSettings,
		migrateUserLogs,
		migrateContracts,
		migrateScheduleEvents,
		migrateDailyGoals,
		migrateReports,
		migrateReportDrafts,
		migrateSalonsGeo,
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
			status VARCHAR(50) DEFAULT 'active',
			tenant_id UUID,
			salon_id UUID,
			managed_by UUID,
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			phone VARCHAR(50),
			display_name VARCHAR(255),
			position VARCHAR(255),
			bio TEXT,
			quote VARCHAR(255),
			avatar_url VARCHAR(500),
			user_status VARCHAR(50) DEFAULT 'online',
			available_for_questions BOOLEAN DEFAULT TRUE,
			achievements TEXT,
			contacts_telegram VARCHAR(100),
			contacts_phone VARCHAR(50),
			contacts_email_visible BOOLEAN DEFAULT TRUE,
			contacts_phone_visible BOOLEAN DEFAULT TRUE,
			contacts_whatsapp VARCHAR(50),
			contacts_working_hours VARCHAR(100),
			deleted_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active'`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id UUID`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS salon_id UUID`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS managed_by UUID`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS created_at TIMESTAMP`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS display_name VARCHAR(255)`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS position VARCHAR(255)`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS quote VARCHAR(255)`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500)`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS user_status VARCHAR(50) DEFAULT 'online'`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS available_for_questions BOOLEAN DEFAULT TRUE`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS achievements TEXT`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS contacts_telegram VARCHAR(100)`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS contacts_phone VARCHAR(50)`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS contacts_email_visible BOOLEAN DEFAULT TRUE`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS contacts_phone_visible BOOLEAN DEFAULT TRUE`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS contacts_whatsapp VARCHAR(50)`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS contacts_working_hours VARCHAR(100)`)
	return nil
}

func migrateTenants(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS tenants (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			status VARCHAR(50) DEFAULT 'active',
			plan_id UUID REFERENCES plans(id),
			legal_entity TEXT,
			inn VARCHAR(20),
			max_users INTEGER DEFAULT 10,
			paid_until TIMESTAMP,
			grace_period_days INTEGER DEFAULT 7,
			deleted_at TIMESTAMP,
			trial_ends_at TIMESTAMP,
			converted_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	// Добавить колонки если таблица уже существует
	db.Exec(`ALTER TABLE tenants ADD COLUMN IF NOT EXISTS legal_entity TEXT`)
	db.Exec(`ALTER TABLE tenants ADD COLUMN IF NOT EXISTS inn VARCHAR(20)`)
	db.Exec(`ALTER TABLE tenants ADD COLUMN IF NOT EXISTS max_users INTEGER DEFAULT 10`)
	db.Exec(`ALTER TABLE tenants ADD COLUMN IF NOT EXISTS paid_until TIMESTAMP`)
	db.Exec(`ALTER TABLE tenants ADD COLUMN IF NOT EXISTS grace_period_days INTEGER DEFAULT 7`)
	db.Exec(`ALTER TABLE tenants ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP`)
	db.Exec(`ALTER TABLE tenants ADD COLUMN IF NOT EXISTS trial_ends_at TIMESTAMP`)
	db.Exec(`ALTER TABLE tenants ADD COLUMN IF NOT EXISTS converted_at TIMESTAMP`)
	db.Exec(`ALTER TABLE tenants ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP`)
	db.Exec(`ALTER TABLE tenants ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP`)
	// Сделать plan_id nullable без constraint
	db.Exec(`ALTER TABLE tenants ALTER COLUMN plan_id DROP NOT NULL`)
	return nil
}

func migrateSalons(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS salons (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			dealer_id UUID,
			name VARCHAR(255) NOT NULL,
			address TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`ALTER TABLE salons ADD COLUMN IF NOT EXISTS dealer_id UUID`)
	db.Exec(`ALTER TABLE salons ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP`)
	return nil
}

func migrateOrders(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			salon_id UUID NOT NULL REFERENCES salons(id),
			created_by UUID NOT NULL REFERENCES users(id),
			status VARCHAR(50) DEFAULT 'new',
			total_price DECIMAL(10,2),
			description TEXT,
			total DECIMAL(10,2) DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	// Совместимость со старым GORM-скелетом (user_id/tenant_id) и 009 (total)
	db.Exec(`ALTER TABLE orders ADD COLUMN IF NOT EXISTS salon_id UUID`)
	db.Exec(`ALTER TABLE orders ADD COLUMN IF NOT EXISTS created_by UUID`)
	db.Exec(`ALTER TABLE orders ADD COLUMN IF NOT EXISTS total_price DECIMAL(10,2)`)
	db.Exec(`ALTER TABLE orders ADD COLUMN IF NOT EXISTS description TEXT`)
	db.Exec(`ALTER TABLE orders ADD COLUMN IF NOT EXISTS total DECIMAL(10,2) DEFAULT 0`)
	db.Exec(`ALTER TABLE orders ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP`)
	// Индексы как в 001
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_orders_salon ON orders(salon_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_orders_created_by ON orders(created_by)`)
	return nil
}

func migrateTasks(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title VARCHAR(255) NOT NULL,
			description TEXT,
			assigned_to UUID NOT NULL REFERENCES users(id),
			created_by UUID NOT NULL REFERENCES users(id),
			salon_id UUID REFERENCES salons(id),
			status VARCHAR(50) DEFAULT 'pending',
			due_date TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	// Обратная совместимость со старым GORM-скелетом
	db.Exec(`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS assigned_to UUID`)
	db.Exec(`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS created_by UUID`)
	db.Exec(`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS salon_id UUID`)
	db.Exec(`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS user_id UUID`)
	db.Exec(`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS tenant_id UUID`)
	db.Exec(`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_assigned_to ON tasks(assigned_to)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_created_by ON tasks(created_by)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_salon ON tasks(salon_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`)
	return nil
}

func migrateChecklists(db *gorm.DB) error {
	if err := db.Exec(`
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
	`).Error; err != nil {
		return err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_checklists_tenant ON checklists(tenant_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_checklists_assigned ON checklists(assigned_to)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_checklists_user ON checklists(user_id)`)
	return nil
}

func migratePlans(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS plans (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL,
			price DECIMAL(10,2) DEFAULT 0.0,
			max_salons INT DEFAULT 10,
			max_users INT DEFAULT 50,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`ALTER TABLE plans ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP`)
	db.Exec(`ALTER TABLE plans ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP`)
	// Убрать legacy owner_id если остался с древних GORM-миграций — не дропаем, но игнорируем
	return nil
}

func migrateNotifications(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS notifications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID,
			user_id UUID,
			type VARCHAR(50) DEFAULT 'info',
			title VARCHAR(255) NOT NULL,
			message TEXT,
			data TEXT,
			is_read BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_notifications_tenant ON notifications(tenant_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read)`)
	return nil
}

func migrateInvoices(db *gorm.DB) error {
	if err := db.Exec(`
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
	`).Error; err != nil {
		return err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_invoices_tenant ON invoices(tenant_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status)`)
	return nil
}

func migrateLeads(db *gorm.DB) error {
	if err := db.Exec(`
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
	`).Error; err != nil {
		return err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_leads_salon ON leads(salon_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_leads_manager ON leads(manager_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_leads_status ON leads(status)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_leads_created_at ON leads(created_at)`)
	return nil
}

func migrateLeadActivities(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS lead_activities (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			lead_id UUID NOT NULL,
			user_id UUID,
			salon_id UUID,
			type VARCHAR(50) DEFAULT 'note',
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`ALTER TABLE lead_activities ADD COLUMN IF NOT EXISTS salon_id UUID`)
	return nil
}

func migrateChecklistTemplates(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS checklist_templates (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID REFERENCES tenants(id),
			created_by UUID NOT NULL REFERENCES users(id),
			title VARCHAR(255) NOT NULL,
			description TEXT,
			type VARCHAR(50) DEFAULT 'daily',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`ALTER TABLE checklist_templates ADD COLUMN IF NOT EXISTS created_by UUID`)
	db.Exec(`ALTER TABLE checklist_templates ADD COLUMN IF NOT EXISTS type VARCHAR(50) DEFAULT 'daily'`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_checklist_templates_tenant ON checklist_templates(tenant_id)`)
	return nil
}

func migrateChecklistTemplateItems(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS checklist_template_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			template_id UUID NOT NULL REFERENCES checklist_templates(id) ON DELETE CASCADE,
			text VARCHAR(500) NOT NULL,
			order_index INT DEFAULT 0,
			validation_type VARCHAR(50) DEFAULT 'checkbox',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`ALTER TABLE checklist_template_items ADD COLUMN IF NOT EXISTS text VARCHAR(500)`)
	db.Exec(`ALTER TABLE checklist_template_items ADD COLUMN IF NOT EXISTS order_index INT DEFAULT 0`)
	db.Exec(`ALTER TABLE checklist_template_items ADD COLUMN IF NOT EXISTS validation_type VARCHAR(50) DEFAULT 'checkbox'`)
	db.Exec(`ALTER TABLE checklist_template_items ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP`)
	// Legacy GORM колонки — оставим для обратной совместимости
	db.Exec(`ALTER TABLE checklist_template_items ADD COLUMN IF NOT EXISTS title VARCHAR(255)`)
	db.Exec(`ALTER TABLE checklist_template_items ADD COLUMN IF NOT EXISTS description TEXT`)
	db.Exec(`ALTER TABLE checklist_template_items ADD COLUMN IF NOT EXISTS order_num INT DEFAULT 0`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_checklist_template_items_template ON checklist_template_items(template_id)`)
	return nil
}

func migrateAssignedChecklists(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS assigned_checklists (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			template_id UUID NOT NULL REFERENCES checklist_templates(id),
			salon_id UUID NOT NULL REFERENCES salons(id),
			assigned_to UUID NOT NULL REFERENCES users(id),
			due_date DATE,
			status VARCHAR(50) DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`ALTER TABLE assigned_checklists ADD COLUMN IF NOT EXISTS salon_id UUID`)
	db.Exec(`ALTER TABLE assigned_checklists ADD COLUMN IF NOT EXISTS assigned_to UUID`)
	db.Exec(`ALTER TABLE assigned_checklists ADD COLUMN IF NOT EXISTS due_date DATE`)
	// Legacy
	db.Exec(`ALTER TABLE assigned_checklists ADD COLUMN IF NOT EXISTS user_id UUID`)
	db.Exec(`ALTER TABLE assigned_checklists ADD COLUMN IF NOT EXISTS start_date TIMESTAMP`)
	db.Exec(`ALTER TABLE assigned_checklists ADD COLUMN IF NOT EXISTS end_date TIMESTAMP`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_assigned_checklists_user ON assigned_checklists(assigned_to)`)
	return nil
}

func migrateChecklistResponses(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS checklist_responses (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			assigned_checklist_id UUID NOT NULL REFERENCES assigned_checklists(id),
			item_id UUID NOT NULL REFERENCES checklist_template_items(id),
			value TEXT,
			completed BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`ALTER TABLE checklist_responses ADD COLUMN IF NOT EXISTS value TEXT`)
	db.Exec(`ALTER TABLE checklist_responses ADD COLUMN IF NOT EXISTS completed BOOLEAN DEFAULT FALSE`)
	// Legacy
	db.Exec(`ALTER TABLE checklist_responses ADD COLUMN IF NOT EXISTS user_id UUID`)
	db.Exec(`ALTER TABLE checklist_responses ADD COLUMN IF NOT EXISTS is_completed BOOLEAN DEFAULT FALSE`)
	db.Exec(`ALTER TABLE checklist_responses ADD COLUMN IF NOT EXISTS response_text TEXT`)
	return nil
}

func migrateAlerts(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS alerts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID,
			tenant_id UUID,
			type VARCHAR(50),
			category VARCHAR(50),
			priority VARCHAR(50),
			severity VARCHAR(20),
			title VARCHAR(255),
			message TEXT,
			description TEXT,
			link VARCHAR(500),
			data TEXT,
			is_read BOOLEAN DEFAULT FALSE,
			status VARCHAR(50) DEFAULT 'new',
			read_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_alerts_user ON alerts(user_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_alerts_tenant ON alerts(tenant_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status)`)
	return nil
}

func migrateDealerTasks(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS dealer_tasks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			dealer_id UUID,
			tenant_id UUID,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(50) DEFAULT 'pending',
			priority VARCHAR(50) DEFAULT 'normal',
			due_date TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_dealer_tasks_dealer ON dealer_tasks(dealer_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_dealer_tasks_tenant ON dealer_tasks(tenant_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_dealer_tasks_status ON dealer_tasks(status)`)
	return nil
}

func migrateDealerRequests(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS dealer_requests (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			dealer_id UUID,
			type VARCHAR(50),
			description TEXT,
			amount NUMERIC(12,2) DEFAULT 0,
			status VARCHAR(50) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_dealer_requests_dealer ON dealer_requests(dealer_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_dealer_requests_status ON dealer_requests(status)`)
	return nil
}

func migrateMarketingBudgets(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS marketing_budgets (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			dealer_id UUID,
			quarter VARCHAR(20),
			total_amount NUMERIC(12,2) DEFAULT 0,
			used_amount NUMERIC(12,2) DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_marketing_budgets_dealer ON marketing_budgets(dealer_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_marketing_budgets_quarter ON marketing_budgets(quarter)`)
	return nil
}

func migrateDealerExpenses(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS dealer_expenses (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			dealer_id UUID,
			category VARCHAR(50),
			amount NUMERIC(12,2) DEFAULT 0,
			period VARCHAR(20),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_dealer_expenses_dealer ON dealer_expenses(dealer_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_dealer_expenses_period ON dealer_expenses(period)`)
	return nil
}

func migrateProducts(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			salon_id UUID,
			name VARCHAR(255) NOT NULL,
			collection VARCHAR(150),
			category VARCHAR(100),
			price NUMERIC(12,2) DEFAULT 0,
			cost_price NUMERIC(12,2) DEFAULT 0,
			showroom_qty INT DEFAULT 0,
			warehouse_qty INT DEFAULT 0,
			turnover_days INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	return db.Exec(`CREATE INDEX IF NOT EXISTS idx_products_salon ON products(salon_id)`).Error
}

func migrateLostSales(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS lost_sales (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			salon_id UUID,
			reason VARCHAR(255),
			requests_count INT DEFAULT 0,
			lost_revenue NUMERIC(12,2) DEFAULT 0,
			period VARCHAR(20),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	return db.Exec(`CREATE INDEX IF NOT EXISTS idx_lost_sales_salon_period ON lost_sales(salon_id, period)`).Error
}

func migratePromotions(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS promotions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			salon_id UUID,
			name VARCHAR(255) NOT NULL,
			condition TEXT,
			discount_min INT DEFAULT 0,
			discount_max INT DEFAULT 0,
			start_date TIMESTAMP,
			end_date TIMESTAMP,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	return db.Exec(`CREATE INDEX IF NOT EXISTS idx_promotions_salon ON promotions(salon_id)`).Error
}

func migrateCategoryTurnover(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS category_turnover (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			salon_id UUID,
			category VARCHAR(100),
			avg_days INT DEFAULT 0,
			period VARCHAR(20),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	return db.Exec(`CREATE INDEX IF NOT EXISTS idx_category_turnover_salon_period ON category_turnover(salon_id, period)`).Error
}

func migrateGoals(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS goals (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			assigner_id UUID REFERENCES users(id),
			assignee_id UUID REFERENCES users(id),
			role VARCHAR(50) NOT NULL,
			sales_plan NUMERIC(15,2) DEFAULT 0,
			leads_plan INT DEFAULT 0,
			calls_plan INT DEFAULT 0,
			meetings_plan INT DEFAULT 0,
			target_date DATE NOT NULL,
			tenant_id UUID REFERENCES tenants(id),
			period VARCHAR(20) DEFAULT 'day',
			start_date DATE,
			end_date DATE,
			sales_fact NUMERIC,
			status VARCHAR(50) DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`ALTER TABLE goals ADD COLUMN IF NOT EXISTS period VARCHAR(20) DEFAULT 'day'`)
	db.Exec(`ALTER TABLE goals ADD COLUMN IF NOT EXISTS start_date DATE`)
	db.Exec(`ALTER TABLE goals ADD COLUMN IF NOT EXISTS end_date DATE`)
	db.Exec(`ALTER TABLE goals ADD COLUMN IF NOT EXISTS sales_fact NUMERIC`)
	db.Exec(`ALTER TABLE goals ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active'`)
	db.Exec(`ALTER TABLE goals ALTER COLUMN assigner_id DROP NOT NULL`)
	db.Exec(`ALTER TABLE goals ALTER COLUMN assignee_id DROP NOT NULL`)
	// RE-AUDIT: параллельный двойной POST давал дубликаты и 500.
	// Естественный ключ (получатель, период, даты): повтор упирается в 409.
	// NULLS NOT DISTINCT — сырые NULL в датах тоже конфликтуют, а не плодятся.
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_goals_assignee_period_dates ON goals(assignee_id, period, start_date, end_date) NULLS NOT DISTINCT`).Error; err != nil {
		return err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_goals_assignee_date ON goals(assignee_id, target_date)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_goals_assigner ON goals(assigner_id)`)
	return nil
}

func migrateSystemSettings(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS system_settings (
			key VARCHAR(255) PRIMARY KEY,
			value TEXT,
			description TEXT,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	// Дефолты из 010/014 (идемпотентно)
	db.Exec(`INSERT INTO system_settings (key, value, description) VALUES ('target_conversion', '30', 'Целевая конверсия в продажу, %') ON CONFLICT (key) DO NOTHING`)
	db.Exec(`INSERT INTO system_settings (key, value, description) VALUES ('target_extras_percent', '15', 'Целевая доля допов в выручке, %') ON CONFLICT (key) DO NOTHING`)
	db.Exec(`INSERT INTO system_settings (key, value, description) VALUES ('max_bonus', '50000', 'Максимальная премия менеджера, RUB') ON CONFLICT (key) DO NOTHING`)
	db.Exec(`INSERT INTO system_settings (key, value, description) VALUES ('marketing_spend_current_month', '0', 'Расходы на маркетинг за текущий месяц (RUB)') ON CONFLICT (key) DO NOTHING`)
	db.Exec(`INSERT INTO system_settings (key, value, description) VALUES ('avg_margin_percent', '32', 'Средняя маржа сети, %') ON CONFLICT (key) DO NOTHING`)
	db.Exec(`INSERT INTO system_settings (key, value, description) VALUES ('network_avg_conversion', '15', 'Средняя конверсия сети, %') ON CONFLICT (key) DO NOTHING`)
	db.Exec(`INSERT INTO system_settings (key, value, description) VALUES ('network_avg_check', '80000', 'Средний чек сети, RUB') ON CONFLICT (key) DO NOTHING`)
	db.Exec(`INSERT INTO system_settings (key, value, description) VALUES ('extras_rate', '0.1', 'Доля допов от выручки') ON CONFLICT (key) DO NOTHING`)
	db.Exec(`INSERT INTO system_settings (key, value, description) VALUES ('discount_default_percent', '5', 'Скидка по умолчанию, %') ON CONFLICT (key) DO NOTHING`)
	db.Exec(`INSERT INTO system_settings (key, value, description) VALUES ('net_profit_rate', '0.2', 'Доля чистой прибыли') ON CONFLICT (key) DO NOTHING`)
	db.Exec(`INSERT INTO system_settings (key, value, description) VALUES ('gross_margin_rate', '0.35', 'Доля маржинальной прибыли') ON CONFLICT (key) DO NOTHING`)
	db.Exec(`INSERT INTO system_settings (key, value, description) VALUES ('cogs_rate', '0.65', 'Доля себестоимости') ON CONFLICT (key) DO NOTHING`)
	db.Exec(`INSERT INTO system_settings (key, value, description) VALUES ('prev_month_factor', '0.9', 'Коэффициент прошлого месяца') ON CONFLICT (key) DO NOTHING`)
	return nil
}

func migrateUserLogs(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id),
			tenant_id UUID REFERENCES tenants(id),
			action VARCHAR(50) DEFAULT 'api_request',
			ip_address VARCHAR(45),
			user_agent TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_user_logs_created_at ON user_logs(created_at)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_user_logs_user_id ON user_logs(user_id)`)
	return nil
}

func migrateContracts(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS contracts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			salon_id UUID NOT NULL,
			lead_id UUID,
			manager_id UUID NOT NULL,
			client_name VARCHAR(255) NOT NULL,
			client_phone VARCHAR(50),
			client_email VARCHAR(255),
			total_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
			prepaid_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
			paid_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
			remain_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
			margin_percent NUMERIC(5,2) NOT NULL DEFAULT 0,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			payment_status VARCHAR(50) NOT NULL DEFAULT 'awaiting_payment',
			payment_date TIMESTAMP,
			deadline_date TIMESTAMP,
			products TEXT,
			description TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_contracts_salon_id ON contracts(salon_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_contracts_manager_id ON contracts(manager_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_contracts_lead ON contracts(lead_id)`)
	return nil
}

func migrateScheduleEvents(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schedule_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			salon_id UUID NOT NULL,
			user_id UUID NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			type VARCHAR(50) DEFAULT 'task',
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP,
			priority VARCHAR(50) DEFAULT 'normal',
			status VARCHAR(50) DEFAULT 'planned',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_schedule_events_salon ON schedule_events(salon_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_schedule_events_user ON schedule_events(user_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_schedule_events_start ON schedule_events(start_time)`)
	return nil
}

func migrateDailyGoals(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS daily_goals (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			salon_id UUID,
			user_id UUID,
			tenant_id UUID REFERENCES tenants(id),
			target_date DATE NOT NULL,
			sales_plan DECIMAL(12,2) DEFAULT 0,
			leads_plan INTEGER DEFAULT 0,
			calls_plan INTEGER DEFAULT 0,
			meetings_plan INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`ALTER TABLE daily_goals ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_daily_goals_tenant ON daily_goals(tenant_id)`)
	// PG15+: NULLS NOT DISTINCT чтобы (NULL, date) считалось дублем, иначе UNIQUE не работает для менеджеров без салона
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_daily_goals_salon_date ON daily_goals (salon_id, target_date) NULLS NOT DISTINCT`).Error; err != nil {
		db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_daily_goals_salon_date ON daily_goals (salon_id, target_date)`)
	}
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_daily_goals_user_date ON daily_goals (user_id, target_date) NULLS NOT DISTINCT`).Error; err != nil {
		db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_daily_goals_user_date ON daily_goals (user_id, target_date)`)
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_daily_goals_salon ON daily_goals(salon_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_daily_goals_user ON daily_goals(user_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_daily_goals_date ON daily_goals(target_date)`)
	return nil
}

func migrateReports(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS reports (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			franchiser_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			pdf_url TEXT,
			recipients JSONB,
			blocks JSONB,
			comment TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_reports_franchiser ON reports(franchiser_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_reports_created ON reports(created_at)`)
	return nil
}

func migrateReportDrafts(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS report_drafts (
			user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			data JSONB NOT NULL DEFAULT '{}',
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}
	return nil
}

func migrateSalonsGeo(db *gorm.DB) error {
	db.Exec(`ALTER TABLE salons ADD COLUMN IF NOT EXISTS region VARCHAR(100)`)
	db.Exec(`ALTER TABLE salons ADD COLUMN IF NOT EXISTS city VARCHAR(100)`)
	db.Exec(`ALTER TABLE salons ADD COLUMN IF NOT EXISTS lat DOUBLE PRECISION`)
	db.Exec(`ALTER TABLE salons ADD COLUMN IF NOT EXISTS lng DOUBLE PRECISION`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_salons_region ON salons(region)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_salons_city ON salons(city)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_goals_period_start ON goals(period, start_date)`)
	db.Exec(`INSERT INTO system_settings (key, value, description) VALUES ('default_monthly_plan', '4000000', 'Дефолтный месячный план дилера, RUB') ON CONFLICT (key) DO NOTHING`)
	return nil
}

func GetDB() *gorm.DB {
	return DB
}
