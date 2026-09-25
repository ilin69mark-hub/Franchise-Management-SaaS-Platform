DROP TABLE IF EXISTS auth_sessions;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_auth_version_check;
ALTER TABLE users DROP COLUMN IF EXISTS auth_version;
