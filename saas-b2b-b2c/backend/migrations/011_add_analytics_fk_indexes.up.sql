-- =====================================================
-- 011_add_analytics_fk_indexes.up.sql
-- Индексы на FK-колонки аналитических таблиц (salon_id).
-- Расчёты KPI-дашбордов фильтруют по salon_id и period:
-- без этих индексов при росте объёмов будет seq-scan.
-- Идемпотентно: CREATE INDEX IF NOT EXISTS.
-- =====================================================

CREATE INDEX IF NOT EXISTS idx_products_salon             ON products (salon_id);
CREATE INDEX IF NOT EXISTS idx_lost_sales_salon_period     ON lost_sales (salon_id, period);
CREATE INDEX IF NOT EXISTS idx_promotions_salon            ON promotions (salon_id);
CREATE INDEX IF NOT EXISTS idx_category_turnover_salon_period ON category_turnover (salon_id, period);
