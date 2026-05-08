-- Migration: 008_add_daily_goals
-- Creates daily_goals table for daily sales targets

CREATE TABLE IF NOT EXISTS daily_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    salon_id UUID,
    user_id UUID,
    target_date DATE NOT NULL,
    sales_plan DECIMAL(12,2) DEFAULT 0,
    leads_plan INTEGER DEFAULT 0,
    calls_plan INTEGER DEFAULT 0,
    meetings_plan INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(salon_id, target_date),
    UNIQUE(user_id, target_date)
);

CREATE INDEX IF NOT EXISTS idx_daily_goals_salon ON daily_goals(salon_id);
CREATE INDEX IF NOT EXISTS idx_daily_goals_user ON daily_goals(user_id);
CREATE INDEX IF NOT EXISTS idx_daily_goals_date ON daily_goals(target_date);