-- Откат 012: убираем только то что добавила 012, не трогая 001-011
DROP TABLE IF EXISTS checklists;
DROP TABLE IF EXISTS invoices;
-- колонки tenants/plans/salons/goals не дропаем — они используются приложением
