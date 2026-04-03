-- Поля для Воронки и Health Score
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS trial_ends_at TIMESTAMP;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS converted_at TIMESTAMP;
