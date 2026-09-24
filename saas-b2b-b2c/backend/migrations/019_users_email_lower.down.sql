-- Migration: 019_users_email_lower (rollback)
DROP INDEX IF EXISTS idx_users_email_lower;
