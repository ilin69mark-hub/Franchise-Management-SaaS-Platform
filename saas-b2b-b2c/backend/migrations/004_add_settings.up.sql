-- Таблица для хранения глобальных настроек и метрик (например, Marketing Spend)
CREATE TABLE IF NOT EXISTS system_settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Добавим дефолтное значение для маркетинга
INSERT INTO system_settings (key, value, description) 
VALUES ('marketing_spend_current_month', '0', 'Расходы на маркетинг за текущий месяц (RUB)') 
ON CONFLICT (key) DO NOTHING;
