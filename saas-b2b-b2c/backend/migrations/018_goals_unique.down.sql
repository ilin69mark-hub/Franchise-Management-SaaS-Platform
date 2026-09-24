-- Migration: 018_goals_unique (rollback)
DROP INDEX IF EXISTS idx_goals_assignee_period_dates;
