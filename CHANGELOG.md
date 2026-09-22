# Changelog

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
