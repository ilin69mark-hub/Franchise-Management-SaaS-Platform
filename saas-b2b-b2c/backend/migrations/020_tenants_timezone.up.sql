-- Migration: 020_tenants_timezone
-- S11: per-tenant IANA timezone for day/month boundaries.
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS timezone VARCHAR(64) DEFAULT 'UTC';
