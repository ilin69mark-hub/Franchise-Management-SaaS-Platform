-- Migration: 018_goals_unique
-- RE-AUDIT: естественный ключ целей против дублей при параллельных POST.
CREATE UNIQUE INDEX IF NOT EXISTS idx_goals_assignee_period_dates
    ON goals(assignee_id, period, start_date, end_date) NULLS NOT DISTINCT;
