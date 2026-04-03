-- =====================================================
-- ЭТАП 1: ОЧИСТКА И ПОДГОТОВКА
-- =====================================================

-- Удаляем старые таблицы, если есть, чтобы начать чисто
DROP TABLE IF EXISTS order_documents CASCADE;
DROP TABLE IF EXISTS task_reports CASCADE;
DROP TABLE IF EXISTS tasks CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS salons CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS plans CASCADE;
DROP TABLE IF EXISTS tenants CASCADE;

-- =====================================================
-- ЭТАП 2: ТАБЛИЦЫ ПЛАТФОРМЫ (SUPER ADMIN LEVEL)
-- =====================================================

-- 1. Тарифные планы (для SaaS)
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL, -- например "Старт", "Бизнес"
    price DECIMAL(10, 2) DEFAULT 0.0,
    max_salons INT DEFAULT 10,
    max_users INT DEFAULT 50,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Сети Франчайзи (Tenants)
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL, -- Название сети (например, "Цветочный Дом")
    plan_id UUID REFERENCES plans(id), -- Тариф
    status VARCHAR(50) DEFAULT 'active', -- active, suspended, deleted
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- ЭТАП 3: ПОЛЬЗОВАТЕЛИ И ИЕРАРХИЯ
-- =====================================================

-- 3. Салоны (Точки продаж)
CREATE TABLE salons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id), -- К какой сети принадлежит
    name VARCHAR(255) NOT NULL, -- Название салона
    address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. Пользователи (Единая таблица для всех ролей)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    
    -- Роли: super_admin, franchiser, franchiser_manager, dealer, salon_manager
    role VARCHAR(50) NOT NULL DEFAULT 'salon_manager',
    
    -- Принадлежность к сети (NULL только у super_admin)
    tenant_id UUID REFERENCES tenants(id),
    
    -- Принадлежность к салону (Только для Dealer и Salon Manager)
    salon_id UUID REFERENCES salons(id),
    
    -- ИЕРАРХИЯ: ID начальника (кто поставил задачу / кому подчиняется)
    -- У Dealer это ID Franchiser Manager. У Salon Manager это ID Dealer.
    managed_by UUID REFERENCES users(id),
    
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- ЭТАП 4: БИЗНЕС-ЛОГИКА (ЗАДАЧИ И ЗАКАЗЫ)
-- =====================================================

-- 5. Заказы
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    salon_id UUID NOT NULL REFERENCES salons(id), -- Откуда заказ
    created_by UUID NOT NULL REFERENCES users(id), -- Кто создал (Дилер или Менеджер)
    
    -- Статусы: new, waiting_payment, in_progress, clarification, rejected, completed
    status VARCHAR(50) DEFAULT 'new',
    
    total_price DECIMAL(10, 2),
    description TEXT, -- Что заказали (пока текстом, потом можно сделать items)
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 6. Задачи (бывшие чек-листы, теперь универсальные задачи)
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    
    assigned_to UUID NOT NULL REFERENCES users(id), -- Кому поставлена (Менеджеру)
    created_by UUID NOT NULL REFERENCES users(id),  -- Кто поставил (Дилер/Франчайзер)
    salon_id UUID REFERENCES salons(id), -- Касается какого салона
    
    status VARCHAR(50) DEFAULT 'pending', -- pending, in_progress, review, completed, rejected
    due_date TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 7. Отчеты по задачам (Скриншоты и комментарии)
CREATE TABLE task_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id),
    user_id UUID NOT NULL REFERENCES users(id),
    
    comment TEXT,
    screenshot_url TEXT, -- Ссылка на файл в S3/Local Storage
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 8. Внутренние сообщения (Чат)
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id UUID NOT NULL REFERENCES users(id),
    receiver_id UUID NOT NULL REFERENCES users(id),
    
    content TEXT,
    attachment_url TEXT, -- Файл (счет, фото)
    read_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- ЭТАП 5: ИНДЕКСЫ (Для скорости)
-- =====================================================

CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_salon ON users(salon_id);
CREATE INDEX idx_users_managed_by ON users(managed_by);
CREATE INDEX idx_orders_salon ON orders(salon_id);
CREATE INDEX idx_orders_status ON orders(status);