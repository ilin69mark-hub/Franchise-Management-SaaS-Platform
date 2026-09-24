-- Migration: 019_users_email_lower
-- RE-AUDIT S4: case-insensitive email uniqueness.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower ON users(LOWER(email));
