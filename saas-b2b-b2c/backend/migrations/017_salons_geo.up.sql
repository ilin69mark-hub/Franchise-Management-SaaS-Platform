-- 017_salons_geo.up.sql — география салонов для GetDealersGeography
ALTER TABLE salons ADD COLUMN IF NOT EXISTS region VARCHAR(100);
ALTER TABLE salons ADD COLUMN IF NOT EXISTS city VARCHAR(100);
ALTER TABLE salons ADD COLUMN IF NOT EXISTS lat DOUBLE PRECISION;
ALTER TABLE salons ADD COLUMN IF NOT EXISTS lng DOUBLE PRECISION;
CREATE INDEX IF NOT EXISTS idx_salons_region ON salons(region);
CREATE INDEX IF NOT EXISTS idx_salons_city ON salons(city);
CREATE INDEX IF NOT EXISTS idx_goals_period_start ON goals(period, start_date);
INSERT INTO system_settings (key, value, description) VALUES ('default_monthly_plan', '4000000', 'Дефолтный месячный план дилера, RUB') ON CONFLICT (key) DO NOTHING;
