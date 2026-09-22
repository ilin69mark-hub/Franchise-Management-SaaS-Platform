# Аудит-отчёт: Franchise-Management-SaaS (Wave 1–4 + P1-1..P3)

**Дата**: 2026-09-22 · **Статус**: все P1/P2 закрыты, P3-1/3/4 закрыты, P2-3 N+1 батч done · **Регресс**: зелёный (30 коммитов, gitleaks 0)

---

## 1. Регресс-зелёность (после Wave 5: b9e21ae → fd61194 + filter-repo)

| Слой | Команда | Результат |
|---|---|---|
| Backend Go | `go build ./... && go vet ./... && go test -short ./... && gofmt -l` | ✅ 0 ошибок, 0 FAIL, gofmt чист (vet 0, build 0, test ok 6/6 pkgs) |
| Frontend TS | `npx tsc --noEmit` | ✅ 0 ошибок |
| Frontend тесты | `npx jest` | ✅ **1255/1255** (63 suites, 17s) |
| Прод-сборка | `NEXT_PUBLIC_API_URL=http://localhost:8080 npm run build` (standalone) | ✅ собралась (fail-closed: без env throws в prod) |
| Секреты | `gitleaks detect --all --config .gitleaks.toml` | ✅ 0 реальных утечек (allowlist: default_secret_key_for_development, unsafe-default-*, CHANGE_ME_*) |
| Compose | `docker compose -f saas-b2b-b2c/docker-compose.yml config` + `prod` | ✅ dev ok, prod fail-closed (DB_PASSWORD/JWT_SECRET/NEXT_PUBLIC_API_URL :? must be set) |

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
| `b9e21ae` | `end-of-day` для `dateStr` во всех 13 дашбордах (`kpi_service.go:231` + `GetManagerTargets` теперь учитывает `dateStr`), `T1` realdb, `T2` FK 011 |
| `e306679` | P1-5/6/7 прод-фронт: `docker-compose.prod.yml:42` `NEXT_PUBLIC_API_URL:?must be set`, `Dockerfile.frontend` `ARG`, `TerritoryPlanFactTab:165` `fetch`→`apiClient POST /franchiser/report/generate-pdf`, `NotificationBell/DealerAlerts/CommunicationsTab` `ws://localhost`→`NEXT_PUBLIC_WS_URL`/`wss`+`?token=JWT` |
| `651dbce` | P1-1..P1-4 GORM↔SQL унификация (`database.go:202` + `012/013`): 6 таблиц приведено к моделям, добавлены `goals/system_settings/user_logs/contracts/schedule_events/daily_goals` |
| `9136b13` | P2-1 20+ FK-индексов (`013` + `database.go`: `leads/tasks/notifications/alerts/...`), P2-2 лимиты (`lead 200/alert 50/kpi 100` + `Limit(100)` в 20 местах `kpi_service.go`) |
| `8f88467` | P2-4 `.env.example` (PORT/REDIS/CORS/SEED_PASSWORD/JWT_EXPIRES), P2-5 `ci.yml` (gitleaks→backend/frontend→audit gate), P2-7 `config.go:102` `DBContext` leak (`Sleep→<-Done`) |
| `91a4251` | P3-3 README badge `ilin69mark-hub`/`saas-b2b-b2c/docker-compose -f`, миграции 001-013, `P3-1` .bak/DEBUG удалены, `P3-4` `localhost`→`throw` в prod (`axiosClient/services/*`) |
| `5ab743d`+`1ea7538`+`fd61194` | P2-3 N+1 батч: `GetDashboardTeam 3*N→2 GROUP BY`, `GetTeamAnalytics 11*N→4`, `GetDealerFunnel 5+3*N→4` |

## 4. Git-аудит

- ✅ `main == origin/main` (`fd61194` + `e306679` + `b9e21ae`...), форков 0, секретов 0, `gitleaks --all` 30 коммитов 0.
- ✅ Ветки-мусор удалены, `.bak` удалены, `Allowlist` в `.gitleaks.toml` покрывает `unsafe-default-*`.
- ⚠️ После `filter-repo` коллегам `fetch --all --force && reset --hard origin/main` (инструкция в `saas-b2b-b2c/README.md:440`).

## 5. Что осталось (не баги кода, техдолг)

- **P1**: задать `JWT_SECRET/DB_PASSWORD/NEXT_PUBLIC_API_URL` в проде (fail-closed).
- **P2**: включить `TEST_DB_DSN` в CI для realdb тестов (сейчас SKIP).
- **P3**: полный батч остальных KPI-методов по образцу `GetTeamAnalytics` (паттерн готов, `Limit(100)` уже mitigated) — по желанию.
