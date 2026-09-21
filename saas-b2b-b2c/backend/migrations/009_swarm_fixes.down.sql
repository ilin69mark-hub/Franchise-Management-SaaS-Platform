-- =====================================================
-- 009_swarm_fixes.down.sql
-- Откат: убираем ТОЛЬКО то, что добавила 009_swarm_fixes.
-- =====================================================

ALTER TABLE orders DROP COLUMN IF EXISTS total;

ALTER TABLE users DROP COLUMN IF EXISTS contacts_working_hours;
ALTER TABLE users DROP COLUMN IF EXISTS contacts_whatsapp;
ALTER TABLE users DROP COLUMN IF EXISTS contacts_phone_visible;
ALTER TABLE users DROP COLUMN IF EXISTS contacts_email_visible;
ALTER TABLE users DROP COLUMN IF EXISTS contacts_phone;
ALTER TABLE users DROP COLUMN IF EXISTS contacts_telegram;
ALTER TABLE users DROP COLUMN IF EXISTS achievements;
ALTER TABLE users DROP COLUMN IF EXISTS available_for_questions;
ALTER TABLE users DROP COLUMN IF EXISTS user_status;
ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;
ALTER TABLE users DROP COLUMN IF EXISTS quote;
ALTER TABLE users DROP COLUMN IF EXISTS bio;
ALTER TABLE users DROP COLUMN IF EXISTS position;
ALTER TABLE users DROP COLUMN IF EXISTS display_name;
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE users DROP COLUMN IF EXISTS phone;
ALTER TABLE users DROP COLUMN IF EXISTS status;

ALTER TABLE goals DROP COLUMN IF EXISTS status;
ALTER TABLE goals DROP COLUMN IF EXISTS sales_fact;

DROP TABLE IF EXISTS dealer_expenses;
DROP TABLE IF EXISTS marketing_budgets;
DROP TABLE IF EXISTS dealer_requests;
DROP TABLE IF EXISTS dealer_tasks;
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS notifications;

ALTER TABLE lead_activities DROP COLUMN IF EXISTS salon_id;
ALTER TABLE salons DROP COLUMN IF EXISTS dealer_id;