package testdb

import (
	"context"
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB *gorm.DB
)

func Connect() (*gorm.DB, error) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5433 user=postgres password=postgres dbname=franchise_test sslmode=disable"
	}

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	return db, nil
}

func SetupSchema(db *gorm.DB) error {
	schema := `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

	CREATE TABLE IF NOT EXISTS leads (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		salon_id UUID NOT NULL,
		manager_id UUID NOT NULL,
		full_name TEXT NOT NULL,
		phone TEXT,
		email TEXT,
		interest_product TEXT,
		budget DECIMAL,
		status TEXT DEFAULT 'new',
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS lead_activities (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		lead_id UUID NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
		user_id UUID NOT NULL,
		type TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS notifications (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		tenant_id UUID,
		user_id UUID,
		type TEXT,
		title TEXT,
		message TEXT,
		is_read BOOLEAN DEFAULT false,
		data TEXT,
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS plans (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name TEXT NOT NULL,
		price DECIMAL NOT NULL,
		max_salons INTEGER NOT NULL,
		max_users INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		deleted_at TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS goals (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		assigner_id UUID NOT NULL,
		assignee_id UUID NOT NULL,
		role TEXT,
		tenant_id UUID,
		sales_plan DECIMAL,
		leads_plan INTEGER,
		calls_plan INTEGER,
		meetings_plan INTEGER,
		period VARCHAR(20) DEFAULT 'day',
		start_date DATE,
		end_date DATE,
		target_date TIMESTAMP,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);
	`

	return db.Exec(schema).Error
}

func Cleanup(db *gorm.DB) error {
	tables := []string{
		"lead_activities",
		"leads",
		"notifications",
		"plans",
		"goals",
	}

	for _, table := range tables {
		if err := db.Exec("DELETE FROM " + table).Error; err != nil {
			return err
		}
	}
	return nil
}

func TruncateAll(db *gorm.DB) error {
	return db.Exec(`
		TRUNCATE TABLE lead_activities, leads, notifications, plans, goals CASCADE
	`).Error
}

func MustConnect(t TB) *gorm.DB {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var db *gorm.DB
	var err error

	for i := 0; i < 30; i++ {
		db, err = Connect()
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for database: %v", ctx.Err())
		default:
		}
		time.Sleep(500 * time.Millisecond)
	}

	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if err := SetupSchema(db); err != nil {
		t.Fatalf("failed to setup schema: %v", err)
	}

	t.Cleanup(func() {
		TruncateAll(db)
	})

	return db
}

type TB interface {
	Helper()
	Cleanup(func())
	Fatalf(format string, args ...interface{})
}