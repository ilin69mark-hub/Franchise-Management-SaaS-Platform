-- =====================================================
-- 011_add_analytics_fk_indexes.down.sql
-- Откат: убираем ТОЛЬКО то, что добавила 011 (индексы аналитики).
-- =====================================================

DROP INDEX IF EXISTS idx_products_salon;
DROP INDEX IF EXISTS idx_lost_sales_salon_period;
DROP INDEX IF EXISTS idx_promotions_salon;
DROP INDEX IF EXISTS idx_category_turnover_salon_period;
