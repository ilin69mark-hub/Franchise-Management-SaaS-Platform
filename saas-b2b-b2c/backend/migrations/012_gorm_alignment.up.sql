-- =====================================================
-- 012_gorm_alignment.up.sql
-- Выравнивает SQL-инициализацию (docker-entrypoint) с GORM-миграциями.
-- До этого GORM создавал часть таблиц (invoices, checklists, tenants-дополнения),
-- которых не было в 001-011, поэтому чистая SQL-база без запуска backend
-- оставалась неполной. Все операции идемпотентны.
-- =====================================================

-- GORM-таблицы, которых нет в 001-011
CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    amount DECIMAL(10,2) DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending',
    due_date TIMESTAMP,
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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
);

-- Tenants: GORM-расширения (legal_entity/inn/paid_until) отсутствуют в 001-002
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS legal_entity TEXT;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS inn VARCHAR(20);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS max_users INTEGER DEFAULT 10;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS paid_until TIMESTAMP;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS grace_period_days INTEGER DEFAULT 7;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Plans: GORM добавил updated_at/deleted_at, SQL 001 имел только created_at
ALTER TABLE plans ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE plans ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

-- Salons: GORM/009 dealer_id уже покрыты, но страхуем чистый SQL
ALTER TABLE salons ADD COLUMN IF NOT EXISTS dealer_id UUID;
ALTER TABLE salons ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Goals: на случай если 001 не отработал (например, DROP в тестах)
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
);
ALTER TABLE goals ADD COLUMN IF NOT EXISTS period VARCHAR(20) DEFAULT 'day';
ALTER TABLE goals ADD COLUMN IF NOT EXISTS start_date DATE;
ALTER TABLE goals ADD COLUMN IF NOT EXISTS end_date DATE;
