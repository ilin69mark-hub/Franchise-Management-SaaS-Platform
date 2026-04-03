-- Таблица для логирования действий (DAU/MAU)
CREATE TABLE IF NOT EXISTS user_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    tenant_id UUID REFERENCES tenants(id),
    action VARCHAR(50) DEFAULT 'api_request',
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для скорости
CREATE INDEX IF NOT EXISTS idx_user_logs_created_at ON user_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_user_logs_user_id ON user_logs(user_id);

-- Поля для Воронки (Trial -> Paid)
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS trial_ends_at TIMESTAMP;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS converted_at TIMESTAMP;
