-- =====================================================
-- 009_swarm_fixes.up.sql
-- Исправление дрейфа схемы: едимей полезной нагрузки между
-- SQL-миграциями 001-008 и GORM AutoMigrate / кодом сервисов.
-- ВСЕ операции идемпотентны (безопасно запускать повторно).
-- =====================================================

-- ---------------------------------------------------------------
-- salons: колонка dealer_id (используется kpi_service.go)
-- ---------------------------------------------------------------
ALTER TABLE salons ADD COLUMN IF NOT EXISTS dealer_id UUID;

-- ---------------------------------------------------------------
-- lead_activities: колонка salon_id (используется kpi_service.go)
-- таблица создана в 005, добавляем только недостающую колонку
-- ---------------------------------------------------------------
ALTER TABLE lead_activities ADD COLUMN IF NOT EXISTS salon_id UUID;

-- ---------------------------------------------------------------
-- notifications: таблицы нет в 001-008 (GORM создаёт без tenant_id/data)
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    user_id UUID,
    title VARCHAR(255),
    message TEXT,
    type VARCHAR(50),
    data TEXT,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ---------------------------------------------------------------
-- alerts (используется seed.go, model Alert, kpi_service.go)
-- ---------------------------------------------------------------
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
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ---------------------------------------------------------------
-- dealer_tasks (kpi_service.go: GetDealerTasks / UpdateDealerTask)
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS dealer_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dealer_id UUID,
    tenant_id UUID,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'pending',
    priority VARCHAR(50) DEFAULT 'normal',
    due_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ---------------------------------------------------------------
-- dealer_requests (kpi_service.go: GetDealerRequests / CreateDealerRequest)
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS dealer_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dealer_id UUID,
    type VARCHAR(50),
    description TEXT,
    amount NUMERIC(12,2) DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ---------------------------------------------------------------
-- marketing_budgets (kpi_service.go: GetDealerMarketingBudget)
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS marketing_budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dealer_id UUID,
    quarter VARCHAR(20),
    total_amount NUMERIC(12,2) DEFAULT 0,
    used_amount NUMERIC(12,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ---------------------------------------------------------------
-- dealer_expenses (kpi_service.go: GetDealerFinance)
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS dealer_expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dealer_id UUID,
    category VARCHAR(50),
    amount NUMERIC(12,2) DEFAULT 0,
    period VARCHAR(20),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ---------------------------------------------------------------
-- goals: sales_fact, status + nullable assigner_id/assignee_id
-- (seed.go вставляет цели без assigner; модель Goal)
-- ---------------------------------------------------------------
ALTER TABLE goals ADD COLUMN IF NOT EXISTS sales_fact NUMERIC;
ALTER TABLE goals ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active';

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_schema = current_schema() AND table_name = 'goals'
                 AND column_name = 'assigner_id' AND is_nullable = 'NO') THEN
        ALTER TABLE goals ALTER COLUMN assigner_id DROP NOT NULL;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_schema = current_schema() AND table_name = 'goals'
                 AND column_name = 'assignee_id' AND is_nullable = 'NO') THEN
        ALTER TABLE goals ALTER COLUMN assignee_id DROP NOT NULL;
    END IF;
END $$;

-- ---------------------------------------------------------------
-- users: профильные колонки модели User (см. models/user.go)
-- ---------------------------------------------------------------
ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active';
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

ALTER TABLE users ADD COLUMN IF NOT EXISTS display_name VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS position VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS quote VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);
ALTER TABLE users ADD COLUMN IF NOT EXISTS user_status VARCHAR(50) DEFAULT 'online';
ALTER TABLE users ADD COLUMN IF NOT EXISTS available_for_questions BOOLEAN DEFAULT TRUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS achievements TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS contacts_telegram VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS contacts_phone VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS contacts_email_visible BOOLEAN DEFAULT TRUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS contacts_phone_visible BOOLEAN DEFAULT TRUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS contacts_whatsapp VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS contacts_working_hours VARCHAR(100);

-- ---------------------------------------------------------------
-- orders: total (используется GORM). total_price остаётся из 001.
-- ---------------------------------------------------------------
ALTER TABLE orders ADD COLUMN IF NOT EXISTS total NUMERIC(10,2);