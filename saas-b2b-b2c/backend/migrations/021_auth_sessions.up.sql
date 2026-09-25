ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_version BIGINT NOT NULL DEFAULT 1;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM users WHERE auth_version <= 0) THEN
        RAISE EXCEPTION 'users.auth_version must be positive';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'users_auth_version_check'
          AND conrelid = 'users'::regclass
    ) THEN
        ALTER TABLE users ADD CONSTRAINT users_auth_version_check CHECK (auth_version > 0);
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS auth_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    auth_version BIGINT NOT NULL DEFAULT 1,
    current_refresh_jti UUID NOT NULL UNIQUE,
    chain_started_at TIMESTAMPTZ NOT NULL,
    chain_expires_at TIMESTAMPTZ NOT NULL,
    refresh_expires_at TIMESTAMPTZ NOT NULL,
    last_token_issued_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    revoke_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT auth_sessions_auth_version_check CHECK (auth_version > 0),
    CONSTRAINT auth_sessions_refresh_expiry_check CHECK (refresh_expires_at <= chain_expires_at)
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'auth_sessions_auth_version_check'
          AND conrelid = 'auth_sessions'::regclass
    ) THEN
        ALTER TABLE auth_sessions ADD CONSTRAINT auth_sessions_auth_version_check CHECK (auth_version > 0);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'auth_sessions_refresh_expiry_check'
          AND conrelid = 'auth_sessions'::regclass
    ) THEN
        ALTER TABLE auth_sessions ADD CONSTRAINT auth_sessions_refresh_expiry_check CHECK (refresh_expires_at <= chain_expires_at);
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_auth_sessions_user_active ON auth_sessions(user_id) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_auth_sessions_refresh_expiry_active ON auth_sessions(refresh_expires_at) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_auth_sessions_revoked ON auth_sessions(revoked_at) WHERE revoked_at IS NOT NULL;
