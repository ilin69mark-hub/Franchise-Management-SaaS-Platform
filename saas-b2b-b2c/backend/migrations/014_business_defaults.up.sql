-- 014_business_defaults.up.sql — выносит хардкоды KPI в system_settings
INSERT INTO system_settings (key, value, description) VALUES
  ('avg_margin_percent', '32', 'Средняя маржа сети, % (fallback для P1 хардкода 32.0)'),
  ('network_avg_conversion', '15', 'Средняя конверсия сети, % (fallback 15.0)'),
  ('network_avg_check', '80000', 'Средний чек сети, RUB (fallback 80000)'),
  ('extras_rate', '0.1', 'Доля допов от выручки (fallback 0.1)'),
  ('discount_default_percent', '5', 'Скидка по умолчанию, % (fallback 5.0)'),
  ('net_profit_rate', '0.2', 'Доля чистой прибыли от выручки (fallback 0.2)'),
  ('gross_margin_rate', '0.35', 'Доля маржинальной прибыли (fallback 0.35)'),
  ('cogs_rate', '0.65', 'Доля себестоимости COGS (fallback 0.65)'),
  ('prev_month_factor', '0.9', 'Коэффициент прошлого месяца для прогноза (fallback 0.9)')
ON CONFLICT (key) DO NOTHING;
