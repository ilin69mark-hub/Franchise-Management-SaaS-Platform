-- =====================================================
-- 010_products_analytics.up.sql
-- Таблицы для честных данных дашборда товаров и директив
-- менеджера вместо кодовых заглушек (см. kpi_service.go:
-- GetDashboardProducts / GetManagerTargets / GetDealerProducts).
-- ВСЕ операции идемпотентны.
-- =====================================================

-- ---------------------------------------------------------------
-- products: каталог товаров салона (остатки, коллекция, категория,
-- маржа = (price - cost_price) / price)
-- ---------------------------------------------------------------
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
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ---------------------------------------------------------------
-- lost_sales: упущенные продажи по причинам отказа
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS lost_sales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    salon_id UUID,
    reason VARCHAR(255),
    requests_count INT DEFAULT 0,
    lost_revenue NUMERIC(12,2) DEFAULT 0,
    period VARCHAR(20),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ---------------------------------------------------------------
-- promotions: акции салона (директивы менеджеру)
-- ---------------------------------------------------------------
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
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ---------------------------------------------------------------
-- category_turnover: средняя оборачиваемость по категориям
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS category_turnover (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    salon_id UUID,
    category VARCHAR(100),
    avg_days INT DEFAULT 0,
    period VARCHAR(20),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Бенчмарки директив: читаются сервисом, дефолты для новых БД.
INSERT INTO system_settings (key, value, description)
VALUES ('target_conversion', '30', 'Целевая конверсия в продажу, %')
ON CONFLICT (key) DO NOTHING;

INSERT INTO system_settings (key, value, description)
VALUES ('target_extras_percent', '15', 'Целевая доля допов в выручке, %')
ON CONFLICT (key) DO NOTHING;

INSERT INTO system_settings (key, value, description)
VALUES ('max_bonus', '50000', 'Максимальная премия менеджера, RUB')
ON CONFLICT (key) DO NOTHING;