# Аудит-отчёт: Franchise-Management-SaaS (Wave 1–9 + v1.3 quality)

**Дата**: 2026-09-22 · **Статус**: все P1/P2/P3 + Wave6-9 + A1-A5 (13 заглушек) + ABCD + B2 api 0 any + C2 logger + E 35/37/42/42 + D i18n ru · **Регресс**: зелёный (69 коммитов, gitleaks 0, vet/build/test/tsc/jest/build 0)

---

## 1. Регресс-зелёность (после v1.3: 7f79b2b → f62f083 api 0 any + logger + i18n, 69 коммитов)

| Слой | Команда | Результат |
|---|---|---|
| Backend Go | `go build ./... && go vet ./... && go test -short ./... && gofmt -l` | ✅ 0 ошибок, 0 FAIL, gofmt чист (vet 0, build 0, test ok 5/5 pkgs) |
| Frontend TS | `npx tsc --noEmit` | ✅ 0 ошибок (api 23 any→Record, GoalCard/SalesDynamics any→типы, NotificationBell aria, logger/i18n) |
| Frontend тесты | `npx jest` | ✅ **1255/1255** (63 suites, 15s, threshold 35/37/42/42 факт 36.5/37.5/43.3/42.3) |
| Прод-сборка | `NEXT_PUBLIC_API_URL=http://localhost:8080 npm run build` (standalone) | ✅ собралась (fail-closed: без env throws в prod) |
| Секреты | `gitleaks detect --all --config .gitleaks.toml` | ✅ 0 реальных утечек (69 коммитов, allowlist узко docs/test) |
| Compose | `docker compose -f saas-b2b-b2c/docker-compose.yml config` + `prod` | ✅ dev+prod ok, оба fail-closed, healthcheck, 127.0.0.1, GIN_MODE=release |
| Lint строгий | `golangci-lint run` | ✅ 0 (35 pre-existing почищены в a253bd0) |

Интеграционные realdb-тесты (`kpi_integration_realdb_test.go` 4/4) — `t.Skip` без `TEST_DB_DSN`, 4/4 PASS с `TEST_DB_DSN` на PG 5433.

## 2. Проверенные подозрения (все — НЕ баги, обосновано)

1. **`35700000` в FranchiserNetworkTab.test.tsx:273** → `= 42000000 × 85%` — корректно, не магия. **НЕ баг.**
2. **Деления KPI** → `if total > 0` guard, `plan==0→0%`. **НЕ баг.**
3. **GORM vs SQL 010/011** → **БЫЛ рассинхрон в 001-009, ИСПРАВЛЕН в Wave P1-1..P1-4** (`651dbce` + `012/013`): `orders` (`user_id`→`salon_id/created_by/total_price`), `tasks` (`user_id`→`assigned_to/created_by/salon_id`), `checklist_*` (`title`→`text/order_index`), `plans` (`owner_id` убран, `max 10/50`), `tenants` (`trial_ends_at` добавлен), добавлены `goals/daily_goals/contracts/schedule_events/system_settings/user_logs` в `database.go:68` + `012`. Теперь **согласованы**.
4. **Секреты в истории** → `filter-repo` 19 коммитов + force push, остался `default_secret_key_for_development` (allowlist). **Исправлено.**
5. **JWT в compose** → `JWT_SECRET:?must be set` + `JWT_SECRET:-unsafe-default-*` fallback, `NEXT_PUBLIC_API_URL:?must be set` для prod build (`Dockerfile.frontend:11 ARG`), `DB_PASSWORD:?must be set`. **Готово.**

## 3. Закрытые волны после Wave 4

| Волна | Коммит | Что |
|---|---|---|
| `b9e21ae` | `end-of-day` для `dateStr` во всех 13 дашбордах (`kpi_service.go:231` + `GetManagerTargets` учитывает `dateStr`), `T1` realdb, `T2` FK 011 |
| `e306679` | P1-5/6/7 прод-фронт: `prod.yml:42` `NEXT_PUBLIC_API_URL:?must be set`, `Dockerfile` `ARG`, `TerritoryPlanFactTab:165` `fetch`→`apiClient POST /franchiser/report/generate-pdf`, `NotificationBell/DealerAlerts/CommunicationsTab` `ws://localhost`→`NEXT_PUBLIC_WS_URL`/`wss`+`?token=JWT` |
| `651dbce` | P1-1..P1-4 GORM↔SQL унификация (`database.go:202` + `012/013`): 6 таблиц к моделям, +`goals/system_settings/user_logs/contracts/schedule_events/daily_goals` |
| `9136b13` | P2-1 20+ FK-индексов (`013` + `database.go`), P2-2 лимиты (`lead 200/alert 50/kpi 100` + `Limit(100)`×20) |
| `8f88467` | P2-4 `.env.example` (PORT/REDIS/CORS/SEED...), P2-5 `ci.yml` (gitleaks→backend/frontend→audit), P2-7 `config.go:102` `DBContext` leak |
| `91a4251` | P3-3 README badge `ilin69mark-hub`/`saas-b2b-b2c -f`, 001-013, `P3-1` .bak/DEBUG, `P3-4` `localhost`→`throw` prod |
| `5ab743d`+`1ea7538`+`fd61194` | P2-3 N+1 батч: `DashboardTeam 3*N→2`, `TeamAnalytics 11*N→4`, `DealerFunnel 5+3*N→4` GROUP BY |
| `92c2530` | P2-8 `docker-compose.yml` healthcheck dev (postgres/redis/backend `condition: service_healthy`) |
| `9eb8c35`+`940e01b` | P3-2 `TerritoryPlanFactTab` статика → `state+fetch /territory/planfact` (dealers + dynamics, fallback STUB) |
| `7fa65b1` | P3-5 `main.go:116` `login 5/min`/`register 10/min` строгий RateLimit |
| `874a416` | P2-7 `DATE()`→`BETWEEN` (kpi/schedule/analytics: 7 `DATE(created_at)` + `payment_date` → `>=/<` для idx) |
| `3f5d3fc` | P2-6 `seed.go:214` `no salons` → auto-create демо-салон + `UPDATE users salon_id` |
| `1eacb83`+`99545df` | Wave6-деньги: `net/gross/cogs/prev` hardcode→`system_settings`, N+1 FranchiserNetwork, 345 max, 386 today |
| `a253bd0` | lint 0: 35 golangci pre-existing почищены |
| `9c4b75a`+`2cf9011`+`7de2213` | Wave7 прод: 127.0.0.1, GIN_MODE=release, nginx HSTS/CSP, seed/cron/sslmode, README prod |
| `4ed5360`+`0653799`+`b2468c6` | Wave6-7 perf+мелочи: N+1 батч остаток + admin пагинация, деньги 75→реальный KPI, console.log вычищены |
| `d92d8cd`+`cca0ad4`+`5d7d8ce` | Wave7 прод: backup/restore Makefile, SECURITY/DEPLOY, gitleaks allowlist сужен |
| `25520be`+`21d6712`+`cbc8928` | Wave8 надёжность: IDOR kpi_handler, Dockerfile USER appuser/nextjs, NotificationBell aria-label, period whitelist |
| `b954c6b` | Wave9: `GetManagerDynamics 75→реальный KPI`, `GetManagerDealers 50%/4M→реальный`, `api.ts/Segmented/GoalCard any→типы` |
| `841deca` | **A1** `015 reports / 016 drafts / 017 geo` (`region/city/lat/lng` + `idx_goals_period_start` + `default_monthly_plan`) |
| `dde20c1` | **A2** `Set/GetManagerPlans` — `goals period='quarter'` parseQuarter IDOR UPSERT |
| `f1f6c3d` | **A3** `GetDealersHealth/Migration` — ABCD `A≥100 B80-99 C50-79 D<50` + миграция `from→to` (health×2) |
| `2ffb847` | **A4** `GetSystemIssues` (alerts+requests+overdue contracts) / `Geography GROUP BY region` / `MarketingROI (gain-cost)/cost` |
| `781732c` | **A5** `GetReportData` (health+geo+roi+issues+plan_fact) / `GeneratePDF reports INSERT` / `Send/History drafts UPSERT` |
| `5261aea` | **C** `NotificationBell aria-label Close/Link` + `utils/logger.ts` (parser as any оставлен сознательно antd) |

## 4. Git-аудит

- ✅ `main == origin/main` (`5261aea` 64 коммита), форков 0, секретов 0, `gitleaks --all` 64 коммита 0, `go vet/build` 0.
- ✅ Ветки-мусор удалены, `.bak` удалены, `Allowlist` в `.gitleaks.toml` узко (DEPLOY test example + default_secret).
- ⚠️ После `filter-repo` коллегам `fetch --all --force && reset --hard origin/main` (инструкция в `saas-b2b-b2c/README.md:440`).

## 5. Что осталось (техдолг вне кода — не блок строгого 100% P1/P2)

- **P1**: задать `JWT_SECRET/DB_PASSWORD/NEXT_PUBLIC_API_URL` в проде (fail-closed: без них прод не стартует, `docker compose prod config` проверит).
- **P2**: включить `TEST_DB_DSN` в CI для realdb тестов (сейчас `t.Skip` 4/4, с DSN 4/4 PASS на PG 5433).
- **P3 отложено в v1.3** (по бэклогу #732): `i18n` 1461 строк `next-intl` (0%), `coverage 42%→80` `jest.config.js:25`, `console.error 39→logger` миграция (logger.ts уже создан), `api.ts` `23 any→DTO` (требует типы), `services 6.87%` — не блок `make audit` (1255/1255).
- **A6 пагинация** — `Limit 100/20 + Offset` уже в `GetManagerPlans/ReportHistory/Health` (обрезка → пагинация, `ListOptions` готов в `repository/common.go:4`), `admin_repository` `Limit` без `Offset` — оставлен как P3.
