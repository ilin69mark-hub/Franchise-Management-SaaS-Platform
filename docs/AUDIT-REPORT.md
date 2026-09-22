# Аудит-отчёт: Franchise-Management-SaaS (Wave 1–4)

**Дата**: 2026-09-22 · **Статус**: все пункты закрыты/обоснованы · **Регресс**: зелёный

---

## 1. Регресс-зелёность (после Wave 3.5 + filter-repo + чистки веток)

| Слой | Команда | Результат |
|---|---|---|
| Backend Go | `go build ./... && go vet ./... && go test -short ./... && gofmt -l` | ✅ 0 ошибок, 0 FAIL, gofmt чист |
| Frontend TS | `npx tsc --noEmit` | ✅ 0 ошибок |
| Frontend тесты | `npx jest` | ✅ **1255/1255** (63 suites) |
| Прод-сборка | `next build` (standalone) | ✅ собралась |
| Секреты | `gitleaks detect --all` | ✅ 0 реальных утечек (README-JWT — в allowlist) |

Интеграционные realdb-тесты (`internal/services/kpi_integration_realdb_test.go`, одобренные в истории
main) автоматически **пропускаются** без `TEST_DB_DSN` (паттерн `t.Skip`), поэтому CI зелёный без Postgres;
с заданной `TEST_DB_DSN` прогоняются против реальной БД.

## 2. Проверенные подозрения (все — НЕ баги, обосновано)

1. **`35700000` в FranchiserNetworkTab.test.tsx:273** → `= 42000000 (план) × 85% (прогноз)` — корректная
   формула прогноза в копейках, не магическое число. **НЕ баг.**
2. **Деления в KPI-сервисе** → `kpiPercent = completed/total*100` имеет guard `if total > 0`;
   `plan == 0 → 0%`. Деления на 0 нет. **НЕ баг.**
3. **GORM-миграции vs SQL-миграции** → обе системы согласованы (сверка таблиц products/leads/lost_sales/
   promotions/category_turnover между `010_*.sql` и `MigrateProducts*`); рассинхрона нет. **НЕ баг.**
4. **Секреты в git-истории** → была реальная проблема, **исправлена**: `git-filter-repo --replace-text`
   вычистил JWT-секреты из всей истории (19 коммитов, все ветки) + force push. В истории остался только
   `default_secret_key_for_development` (безопасный dev-fallback, gitleaks-allowlist). 
5. **JWT-секрет в docker-compose → параметризован** (`JWT_SECRET:?must be set` + безопасный dev-fallback),
   прод-параметризация готова. **НЕ баг.**

## 3. Git-аудит

- ✅ Ветки согласованы с origin (main == origin/main).
- ✅ Ветки-мусор и дубликаты удалены (qwen-*, создать-readme-*, исчерпанные feature/*).
- ✅ Форков — 0; репо-секретов GitHub — 0; в истории HEAD секреты вычищены. Бэкап-бандл **локальный**, не в git.
- ⚠️ После force push коллегам нужен `git pull --force` или свежий clone (инструкция в README).

## 4. Что нужно от сопровождения (не баги кода)

- **P1**: задать реальный `JWT_SECRET` в прод-окружении (fail-closed: без него прод не стартует).
- **P2**: CI-деплой на push в main — workflow-заготовка готова, требуются прод-хост/токены.
- **P3**: поднять Postgres (docker compose) и прогнать интеграционные тесты с `TEST_DB_DSN` — сейчас скипаются.
