# Changelog

## v1.3-prod (2026-09-22) — Quality 100% Δ (api 0 any + logger + coverage + i18n ru)
- **Frontend B2/C2:** `services/api.ts 23 any→Record<string,unknown>` (tsc 0), `store 39 console.error→logger.error` (19 файлов, utils/logger.ts), `NotificationBell aria` уже в v1.2.
- **Quality E/D:** `jest.config.js 30/35/40/40→35/37/42/42` (факт 36.5/37.5/43.3/42.3, 1255/1255), `locales/ru.json 10 ключей + utils/i18n.ts t()` ru-only заглушка (1461 строк → поэтапно).
- **Аудит:** 69 коммитов, `AUDIT-REPORT` v1.3, `make audit` зелёный. Остаток `90 :any` в `TenantsSection/SalonFunnel` sorter/render — P3 низкая, не блок.

## v1.2-prod (2026-09-22) — Strict 100% (13 заглушек + ABCD + reports)
- **DB:** `015 reports`, `016 report_drafts JSONB`, `017 salons_geo region/city/lat/lng + idx_goals_period_start + default_monthly_plan`, GORM `migrateReports/Drafts/Geo`.
- **KPI A2-A5:** `Set/GetManagerPlans` (goals quarter UPSERT IDOR), `GetDealersHealth/Migration` ABCD `A≥100 B80-99 C50-79 D<50` (GROUP BY plan/fact), `GetSystemIssues` (alerts+requests+overdue), `Geography GROUP BY region`, `MarketingROI (gain-cost)/cost`, `GetReportData/GeneratePDF/Send/History/Drafts` (reports/report_drafts).
- **Frontend C:** `NotificationBell aria-label Close/Link + aria-hidden`, `utils/logger.ts` (Sentry-ready), `parser as any` оставлен сознательно (antd InputNumber).
- **Аудит:** 64 коммита, `AUDIT-REPORT` v1.2, `make audit` 1255/1255 зелёный. P3 `i18n/coverage/console` отложено в v1.3 (бэклог #732).

## v1.1-prod (2026-09-22) — Production 100% (Wave6-9)
- **KPI:** `GetManagerDynamics 75→реальный KPI (plans vs leads по месяцам)`, `GetManagerDealers 50%/4M→реальный percent/plan (goals+salon leads)`, `system_settings` fallback.
- **Frontend типы:** `api.ts as any→typed`, `GoalCard/SalesDynamicsChart/Segmented any→типы`, `TerritoryPlanFactTab RawDealer/HistoryRow`, `register.tsx as any→typed`, `NotificationBell aria-label`.
- **Аудит:** 57 коммитов, `AUDIT-REPORT` обновлён, `make audit` зелёный 1255/1255.
- Предыдущий релиз v1.0-prod остаётся базой; дельта Wave9 — последняя миля качества до 100% прода.

## v1.0-prod (2026-09-22) — Production ready
- **DB:** GORM↔SQL унификация (`012/013/014`), 20+ FK-индексов, `daily_goals` NULLS NOT DISTINCT, `sslmode` env, `seed` auto-create салон.
- **KPI:** `end-of-day` для `dateStr` (13 мест), `Limit(100)` в 20 местах, N+1 батч `DashboardTeam/TeamAnalytics/DealerFunnel/FranchiserNetwork`, `DATE→BETWEEN`, хардкоды → `system_settings` (014).
- **Frontend:** `NEXT_PUBLIC_API_URL:?` fail-closed, `ws://localhost`→`wss+token`, `TerritoryPlanFactTab` статика→fetch, `console.log`→`debug`, `logout` dispatch, `register` role.
- **Infra:** `docker-compose.prod` fail-closed (`DB_PASSWORD/JWT_SECRET/NEXT_PUBLIC_API_URL`), `127.0.0.1` для DB/Redis, healthcheck, `nginx` HSTS/CSP + `frontend` health, `GIN_MODE=release`, `cron WithSeconds`.
- **CI/Sec:** `ci.yml` gate (gitleaks→backend/frontend→audit), `gitleaks` allowlist сужен, `CORS Vary`, `RateLimit login 5/min`, `DBContext` leak fix, `golangci-lint` 0.
- **Docs:** `README prod` (`make prod-up`), `SECURITY.md`, `DEPLOY.md`, `AUDIT-REPORT` 38+ коммитов, `make backup/restore`.

## v1.0 (2026-09-22) — Audit Wave 1-4
- `filter-repo` секреты, `Wave 1` RBAC/IDOR, `Wave 2` стабы, `Wave 3` real data `products`.
