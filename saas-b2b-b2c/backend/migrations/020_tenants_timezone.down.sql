-- Migration: 020_tenants_timezone (rollback)
ALTER TABLE tenants DROP COLUMN IF EXISTS timezone;
