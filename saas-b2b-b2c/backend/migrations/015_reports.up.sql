-- 015_reports.up.sql — хранение отчетов франчайзера
CREATE TABLE IF NOT EXISTS reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    franchiser_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pdf_url TEXT,
    recipients JSONB,
    blocks JSONB,
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_reports_franchiser ON reports(franchiser_id);
CREATE INDEX IF NOT EXISTS idx_reports_created ON reports(created_at);
