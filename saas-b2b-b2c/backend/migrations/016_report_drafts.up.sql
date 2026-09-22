-- 016_report_drafts.up.sql — черновики отчетов на пользователя
CREATE TABLE IF NOT EXISTS report_drafts (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    data JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
