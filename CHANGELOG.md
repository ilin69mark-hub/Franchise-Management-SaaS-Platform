# Changelog

## v1.7 (2026-09-23) — GoalCard 100% + GoalList/GoalFormModal (26 тестов)
- **Tests:** `GoalCard.test.tsx` 6 тестов (Spin/404/error/data — файл 100%), `GoalList.test.tsx` 11 тестов (canSeeGoal-фильтр, period label, модалка create/edit, delete ok/err через mock `antd` message), `GoalFormModal.test.tsx` 9 тестов (заголовки, ok-disabled без ролей, employees dropdown, onCancel). `global 42.99/37.14/38.14/44.03→44.74/38.87/39.33/45.9`, `70 suites 1294/1294`.
- **Пороги:** `jest.config.js` `35/37/42/42→37/38/44/43`, `make audit` зелёный, коммит `5f04b2c` push в main.

## v1.6 (2026-09-23) — B3 src 0 any (последние 9 any + parser as any)
- **Frontend B3:** `AuditSection` `ColumnsType<AdminAction|ActiveSession|UserLogin>` (интерфейсы export из auditStore), `TechHealthSection` `useState<ErrorLog|null>` + `ColumnsType<ErrorLog>` (`ErrorLog` export из techHealthStore), `ChecklistBoard` `handleFinish ChecklistFormValues` (dayjs RangePicker tuple, `Checklist` тип +`recurrence`), `CommunicationsTab` `WsMessage` discriminated union (Task/Request), `ExpenseFormTab` `handleSave ExpenseFormValues`, `TerritoryFunnelTab` `point Record<string,string|number>`, `TerritoryBenchmarkTab` `CustomTooltip DealerScatter` + `shape props:unknown→cast`. `InputNumber parser as any→Number(...)||0` (`min={0 as number}`).
- **Аудит:** `0 :any/as any` в `src` (кроме `__tests__/setupTests`), `tsc 0`, `jest 1268/1268`, `lint` чисто, коммит `d3958ca` push в main. Остаток `any` — только в тестах (state mocks, сознательно).

## v1.5-prod (2026-09-22) — B3 27 any + utils 97% (logger/i18n 8 tests, 1263/1263)
- **Tests:** `utils/logger.test.ts + i18n.test.ts 8 tests` `utils 37→97%` `global 42.3→42.57%` (35/37/42/42 факт 37.06/37.82/43.61), `65 suites 1263/1263` `audit` зелёный. `ExpenseFormTab/Billing/Activity` B3 уже `27 any (<30)`.
- **Docs:** `any 141→27` (<30 достигнут), `B3` 8 файлов уже `v1.4`.

## v1.4-prod (2026-09-22) — B3 ColumnsType<Tenant> DTO (TenantsSection/SalonFunnel/Billing)
- **Frontend B3:** `TenantsSection` `ColumnsType<Tenant>` (`Tenant` export store, sorter/onFilter `Tenant`, `handleCreateTenant` typed, `onPressEnter KeyboardEvent`), `SalonFunnelTab` `user User, Lead, HotDeal/FreshLead ColumnsType, catch unknown`, `BillingSection` `handleCreateInvoice/Payment` typed, `onPressEnter KeyboardEvent` — `18 any→0` в 3 файлах, `tsc 0`.
- **Аудит:** 72 коммита, `make audit` 1255/1255 зелёный. Остаток `BillingSection render, ActivitySection` — низкая.

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
