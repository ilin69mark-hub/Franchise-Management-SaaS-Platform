-- =====================================================
-- 001_full_schema.up.sql
-- Полный набор таблиц, включая goals.
-- -----------------------------------------------------
-- Выполнять **одноразово** в чистой базе.
-- =====================================================

-- -------------------------------------------------
-- 1. ОЧИСТКА (на всякий случай)
-- -------------------------------------------------
DROP TABLE IF EXISTS messages          CASCADE;
DROP TABLE IF EXISTS task_reports      CASCADE;
DROP TABLE IF EXISTS tasks             CASCADE;
DROP TABLE IF EXISTS orders            CASCADE;
DROP TABLE IF EXISTS salons            CASCADE;
DROP TABLE IF EXISTS users             CASCADE;
DROP TABLE IF EXISTS tenants           CASCADE;
DROP TABLE IF EXISTS plans             CASCADE;
DROP TABLE IF EXISTS goals             CASCADE;   -- если уже была

-- -------------------------------------------------
-- 2. ПЛАТФОРМЫ (SUPER‑ADMIN LEVEL)
-- -------------------------------------------------
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,               -- "Старт", "Бизнес"
    price DECIMAL(10,2) DEFAULT 0.0,
    max_salons INT DEFAULT 10,
    max_users  INT DEFAULT 50,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,               -- название сети
    plan_id UUID REFERENCES plans(id),
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- -------------------------------------------------
-- 3. ПОЛЬЗОВАТЕЛИ И ИЕРАРХИЯ
-- -------------------------------------------------
CREATE TABLE salons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,

    -- роли: super_admin, franchiser, franchiser_manager, dealer, salon_manager
    role VARCHAR(50) NOT NULL DEFAULT 'salon_manager',

    tenant_id UUID REFERENCES tenants(id),      -- NULL только у super_admin
    salon_id  UUID REFERENCES salons(id),       -- только у dealer & salon_manager
    managed_by UUID REFERENCES users(id),        -- иерархия

    first_name VARCHAR(100),
    last_name  VARCHAR(100),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- -------------------------------------------------
-- 4. БИЗНЕС‑ЛОГИКА (ЗАКАЗЫ, ЗАДАЧИ, ЧАТ)
-- -------------------------------------------------
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    salon_id UUID NOT NULL REFERENCES salons(id),
    created_by UUID NOT NULL REFERENCES users(id),

    status VARCHAR(50) DEFAULT 'new',
    total_price DECIMAL(10,2),
    description TEXT,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,

    assigned_to UUID NOT NULL REFERENCES users(id),
    created_by  UUID NOT NULL REFERENCES users(id),
    salon_id    UUID REFERENCES salons(id),

    status VARCHAR(50) DEFAULT 'pending',
    due_date TIMESTAMP,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE task_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id),
    user_id UUID NOT NULL REFERENCES users(id),

    comment TEXT,
    screenshot_url TEXT,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id   UUID NOT NULL REFERENCES users(id),
    receiver_id UUID NOT NULL REFERENCES users(id),

    content TEXT,
    attachment_url TEXT,
    read_at TIMESTAMP,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- -------------------------------------------------
-- 5. ЦЕЛИ (ПЛАНЫ) – таблица goals
-- -------------------------------------------------
CREATE TABLE goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    assigner_id UUID NOT NULL REFERENCES users(id),      -- кто создал цель
    assignee_id UUID NOT NULL REFERENCES users(id),      -- кому назначена

    role VARCHAR(50) NOT NULL,                           -- franchise_manager, dealer, dealer_manager, salon_manager

    sales_plan    NUMERIC(15,2) DEFAULT 0,
    leads_plan    INT            DEFAULT 0,
    calls_plan    INT            DEFAULT 0,
    meetings_plan INT            DEFAULT 0,

    target_date DATE NOT NULL,

    tenant_id UUID REFERENCES tenants(id),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- -------------------------------------------------
-- 6. ИНДЕКСЫ (ускоряют запросы)
-- -------------------------------------------------
CREATE INDEX idx_users_tenant      ON users (tenant_id);
CREATE INDEX idx_users_salon       ON users (salon_id);
CREATE INDEX idx_users_managed_by ON users (managed_by);

CREATE INDEX idx_orders_salon   ON orders (salon_id);
CREATE INDEX idx_orders_status  ON orders (status);

CREATE INDEX idx_goals_assignee_date ON goals (assignee_id, target_date);
CREATE INDEX idx_goals_assigner      ON goals (assigner_id);
CREATE INDEX idx_goals_role          ON goals (role);
CREATE INDEX idx_goals_tenant        ON goals (tenant_id);
