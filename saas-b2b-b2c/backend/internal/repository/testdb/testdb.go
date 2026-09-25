package testdb

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
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
	db, err := openDSN(dsn)
	if err != nil {
		return nil, err
	}
	DB = db
	return db, nil
}

func openDSN(dsn string) (*gorm.DB, error) {
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
	return db, nil
}

func dsnWithSearchPath(dsn, schema string) string {
	if parsed, err := url.Parse(dsn); err == nil && parsed.Scheme != "" {
		query := parsed.Query()
		query.Set("search_path", schema)
		parsed.RawQuery = query.Encode()
		return parsed.String()
	}
	fields := strings.Fields(dsn)
	for i, field := range fields {
		if strings.HasPrefix(strings.ToLower(field), "search_path=") {
			fields[i] = "search_path=" + schema
			return strings.Join(fields, " ")
		}
	}
	return strings.TrimSpace(dsn) + " search_path=" + schema
}

func MustConnectIsolated(t TB) *gorm.DB {
	t.Helper()
	baseDSN := os.Getenv("TEST_DB_DSN")
	if baseDSN == "" {
		baseDSN = "host=localhost port=5433 user=postgres password=postgres dbname=franchise_test sslmode=disable"
	}
	adminDB, err := openDSN(baseDSN)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := adminDB.Exec(`CREATE SCHEMA "` + schema + `"`).Error; err != nil {
		t.Fatalf("failed to create isolated schema: %v", err)
	}
	db, err := openDSN(dsnWithSearchPath(baseDSN, schema))
	if err != nil {
		_ = adminDB.Exec(`DROP SCHEMA "` + schema + `" CASCADE`)
		t.Fatalf("failed to connect to isolated schema: %v", err)
	}
	if sqlDB, dbErr := db.DB(); dbErr == nil {
		sqlDB.SetMaxOpenConns(1)
	}
	if err := db.Exec(`SET search_path TO "` + schema + `"`).Error; err != nil {
		t.Fatalf("failed to set isolated search_path: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, dbErr := db.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
		_ = adminDB.Exec(`DROP SCHEMA "` + schema + `" CASCADE`)
		if sqlDB, dbErr := adminDB.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

func SetupSchema(db *gorm.DB) error {
	var existingUserID struct {
		IDType string
	}
	if err := db.Raw(`
		SELECT udt_name AS id_type
		FROM information_schema.columns
		WHERE table_schema = current_schema()
		  AND table_name = 'users'
		  AND column_name = 'id'
	`).Scan(&existingUserID).Error; err != nil {
		return err
	}
	userIDType := "UUID"
	switch existingUserID.IDType {
	case "text":
		userIDType = "TEXT"
	case "varchar":
		userIDType = "VARCHAR(255)"
	}
	var existingLeadID struct {
		IDType string
	}
	if err := db.Raw(`
		SELECT udt_name AS id_type
		FROM information_schema.columns
		WHERE table_schema = current_schema()
		  AND table_name = 'leads'
		  AND column_name = 'id'
	`).Scan(&existingLeadID).Error; err != nil {
		return err
	}
	leadIDType := "UUID"
	switch existingLeadID.IDType {
	case "text":
		leadIDType = "TEXT"
	case "varchar":
		leadIDType = "VARCHAR(255)"
	}
	schema := fmt.Sprintf(`
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
		contacts_email_visible BOOLEAN DEFAULT TRUE,
		contacts_phone_visible BOOLEAN DEFAULT TRUE,
		contacts_phone VARCHAR(50),
		contacts_telegram VARCHAR(100),
		contacts_working_hours VARCHAR(100),
		contacts_whatsapp VARCHAR(50),
		auth_version BIGINT NOT NULL DEFAULT 1,
		deleted_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		CONSTRAINT users_auth_version_check CHECK (auth_version > 0)
	);

	ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_version BIGINT NOT NULL DEFAULT 1;

	DO $$
	BEGIN
		IF NOT EXISTS (
			SELECT 1 FROM pg_constraint
			WHERE conname = 'users_auth_version_check'
			  AND conrelid = 'users'::regclass
		) THEN
			ALTER TABLE users ADD CONSTRAINT users_auth_version_check CHECK (auth_version > 0);
		END IF;
	END
	$$;

	CREATE TABLE IF NOT EXISTS tenants (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(255) NOT NULL,
		status VARCHAR(50) DEFAULT 'active',
		plan_id UUID,
		timezone VARCHAR(64) DEFAULT 'UTC',
		legal_entity TEXT,
		inn VARCHAR(20),
		max_users INTEGER DEFAULT 10,
		paid_until TIMESTAMPTZ,
		grace_period_days INTEGER DEFAULT 7,
		deleted_at TIMESTAMPTZ,
		trial_ends_at TIMESTAMPTZ,
		converted_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS auth_sessions (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		user_id %s NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		auth_version BIGINT NOT NULL DEFAULT 1,
		current_refresh_jti UUID NOT NULL UNIQUE,
		chain_started_at TIMESTAMPTZ NOT NULL,
		chain_expires_at TIMESTAMPTZ NOT NULL,
		refresh_expires_at TIMESTAMPTZ NOT NULL,
		last_token_issued_at TIMESTAMPTZ NOT NULL,
		revoked_at TIMESTAMPTZ,
		revoke_reason TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		CONSTRAINT auth_sessions_auth_version_check CHECK (auth_version > 0),
		CONSTRAINT auth_sessions_refresh_expiry_check CHECK (refresh_expires_at <= chain_expires_at)
	);

	CREATE INDEX IF NOT EXISTS idx_auth_sessions_user_active ON auth_sessions(user_id) WHERE revoked_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_auth_sessions_refresh_expiry_active ON auth_sessions(refresh_expires_at) WHERE revoked_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_auth_sessions_revoked ON auth_sessions(revoked_at) WHERE revoked_at IS NOT NULL;

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
		lead_id %s NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
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
	`, userIDType, leadIDType)

	return db.Exec(schema).Error
}

func Cleanup(db *gorm.DB) error {
	var tables []string
	if err := db.Raw(`
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = current_schema()
	`).Scan(&tables).Error; err != nil {
		return err
	}
	if len(tables) == 0 {
		return nil
	}
	quoted := make([]string, 0, len(tables))
	for _, table := range tables {
		quoted = append(quoted, `"`+strings.ReplaceAll(table, `"`, `""`)+`"`)
	}
	return db.Exec("TRUNCATE TABLE " + strings.Join(quoted, ", ") + " RESTART IDENTITY CASCADE").Error
}

func TruncateAll(db *gorm.DB) error {
	return Cleanup(db)
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

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}
	lockConn, err := sqlDB.Conn(ctx)
	if err != nil {
		t.Fatalf("failed to acquire test DB connection: %v", err)
	}
	const advisoryLockID int64 = 721602250613
	if _, err := lockConn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", advisoryLockID); err != nil {
		_ = lockConn.Close()
		t.Fatalf("failed to lock test DB: %v", err)
	}
	t.Cleanup(func() {
		_ = TruncateAll(db)
		_, _ = lockConn.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", advisoryLockID)
		_ = lockConn.Close()
	})

	if err := SetupSchema(db); err != nil {
		t.Fatalf("failed to setup schema: %v", err)
	}

	return db
}

type TB interface {
	Helper()
	Cleanup(func())
	Fatalf(format string, args ...interface{})
}
