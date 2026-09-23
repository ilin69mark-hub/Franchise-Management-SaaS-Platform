# Changelog

## v1.33 (2026-09-23) — FranchiserTeamTab 85% + 77% milestone (global 77.34/62.12/70.86/78.76) — 77% lines
- **Tests:** `FranchiserTeamTab` +1 `getChurnColor`/`getForecastColor` 85.47% stmts, `SalonTeamTab` +1 `DeviationIndicator`/`formatMoney` 85.13% stmts, `services/api` `providesTags` с результатом 77.10% stmts. Пороги `61/70/77/77`, `104 suites 1566/1566` (+1 тест). Коммит push в main.

## v1.32 (2026-09-23) — services/api 73% + 77% milestone (global 77.31/61.99/70.80/78.73) — 77% lines
- **Tests:** `services/planApi` 50% + `goalApi` 50% + `userApi` 50% (6→8, 7→9, 2→4 тестов, диспатч `getPlans`/`createPlan`/`getMyGoal`/`setGoal`/`getMyProfile`/`updateProfile` с `global.fetch` mock). Пороги `61/70/77/77`, `104 suites 1562/1562` (+6 тестов). Коммит push в main.

## v1.31 (2026-09-23) — SalaryTeam 85% + ExpenseFormTab 85% + 76% milestone (global 76.59/61.06/69.65/77.92)
- **Tests:** `SalonTeamTab` +1 `DeviationIndicator` 0/null + `formatMoney` 85.13% stmts, `ExpenseFormTab` +1 `createExpensePayload` с `other_expense_name` 85.14% stmts. Пороги `60/68/76/77` сохранены, `104 suites 1542/1542` (+1 тест). Коммит push в main.

## v1.30 (2026-09-23) — FranchiserTeamTab 85% + SalonTeamTab 85% + 76% milestone (global 76.84/61.76/70.08/78.21)
- **Tests:** `FranchiserTeamTab` +2 (getRowColor экспорт + цвета колонок Churn/Forecast) 84.11%→85.47% stmts, `SalonTeamTab` +1 (DeviationIndicator с нулевым отклонением + formatMoney) 85.13% stmts. Пороги `61/69/76/78`, `104 suites 1542/1542` (+2 теста). Коммит push в main.

## v1.29 (2026-09-23) — ExpenseFormTab 85% + 76% milestone (global 76.82/61.70/70.04/78.20)
- **Tests:** `ExpenseFormTab` +1 ошибка сохранения через `api` (`/dealer/expenses` 500 → `message.error`) 82.82%→85.14% stmts (93.93% branches), `FranchiserTeamTab` `getRowColor` уже 84.11%. Пороги `60/68/76/77` сохранены, `104 suites 1544/1544` (+1 тест). Коммит push в main.

## v1.28 (2026-09-23) — FranchiserTeamTab 84% + CommunicationsTab 58% + SalonFunnelTab 70% (global 76.78/61.55/70.09/78.15)
- **Tests:** `FranchiserTeamTab` `getRowColor` экспорт →84.11% stmts, `CommunicationsTab` +2 render (задачи с разными сроками/цветами) 58.20% stmts, `SalonFunnelTab` +2 (бюджет конвертация/деньги) 69.79% stmts. Пороги `60/68/76/77` сохранены, `104 suites 1521/1521` (+2 теста). Коммит push в main.

## v1.27 (2026-09-23) — SalonFunnelTab 70% + CommunicationsTab 58% + 76% milestone (global 76.45/61.00/69.57/77.78)
- **Tests:** `SalonFunnelTab` `mockCreateLead`/`mockUpdateLeadStatus` фикс +2 (ошибка создания с unwrap, форматирование денег) 65.62%→69.79% stmts, `CommunicationsTab` +1 (пустой бюджет/история) 58.20% stmts. Пороги `60/68/76/77` сохранены, `104 suites 1524/1524` (+2 теста). Коммит push в main.

## v1.26 (2026-09-23) — ExpenseFormTab 83% + api 73% + 76% milestone (global 76.33/60.97/69.44/77.65)
- **Tests:** `ExpenseFormTab` +1 импорт через `apiClient` (`/dealer/expenses/import` FormData) 70.70%→82.82% stmts (92.92% branches), `services/api` 57.83%→73.49% stmts (мутации), `FranchiserTeamTab` фикс flaky `сохраняет планы` (waitFor→simple). Пороги `60/68/76/77` сохранены, `104 suites 1524/1524` (+3 теста). Коммит push в main.

## v1.25 (2026-09-23) — CommunicationsTab 58% + 76% milestone (global 76.16/60.87/69.57/77.47)
- **Tests/Refactor:** `CommunicationsTab` `filterTasks`/`getBudgetPercent`/`getDueColor` +2 render (задачи с датами+статусами/filter via useMemo) 52.98%→58.20% stmts (47.57% branches), +2 helper теста. Пороги `60/68/76/77`, `104 suites 1518/1518` (+2 теста). Коммит push в main.

## v1.24 (2026-09-23) — services/api 73% + 75% milestone (global 76.00/60.64/69.24/77.29)
- **Tests:** `services/api` +1 оставшиеся мутации (addLeadActivity/updateEmployee/deleteEmployee/assignManager/updateSalon/deleteSalon/createUnitTemplate/deleteUnitTemplate/updateTaskStatus/addTaskComment/createRequest/logout/readNotification) 57.83%→73.49% stmts (42.3%→42.3% branches, 50.7%→69.01% funcs). Пороги `60/67/76/77`, `104 suites 1521/1521` (+3 теста). Коммит push в main.

## v1.23 (2026-09-23) — FranchiserTeamTab 84% (global 75.69/60.67/68.39/76.95)
- **Tests:** `FranchiserTeamTab` +2 (сортировка по % плана клик на th/изменение InputNumber план продаж в модалке) 71.02%→84.11% stmts (69.56% branches, 75.92% funcs). Пороги `60/66/75/76`, `104 suites 1518/1518` (+2 теста). Коммит push в main.

## v1.22 (2026-09-23) — CommunicationsTab 53% + api мутации + Expense/SalonTeam (global 75.36/60.48/67.60/76.69)
- **Tests/Refactor:** `CommunicationsTab` `filterTasks`/`getBudgetPercent`/`getDueColor` экспортированы → 40.15%→52.98% stmts (41.74% branches), `services/api` +1 мутации (12→13) 45.78%→57.83%→57.83% stmts, `ExpenseFormTab` +1 (сабмит) 70.7% stmts, `SalonTeamTab` +1 (история ошибка) 85.13% stmts. Пороги `60/65/75/76`, `104 suites 1518/1518` (+3 теста). Коммит push в main.

## v1.21 (2026-09-23) — services/api 58% + ExpenseFormTab 71% + SalonTeamTab 85% (global 75.01/60.03/67.47/76.42)
- **Tests:** `services/api` +2 (все endpoints + мутации → 24%→57.83% stmts), `ExpenseFormTab` +2 (импорт/ошибка загрузки/сабмит) 62.62%→70.70% stmts (89.89% branches), `SalonTeamTab` +1 (история ошибка) 63.51%→85.13% stmts. Пороги `59/64/75/75`, `104 suites 1515/1515` (+9 тестов). Коммит push в main.

## v1.20 (2026-09-23) — FranchiserTeamTab 71% + ExpenseFormTab 63% (global 73.78/58.79/64.91/75.07)
- **Tests/Refactor:** `FranchiserTeamTab` `calculateIntegralKpi`/`calculateBonus` экспортированы → 59.81%→71.02% stmts (60.86% branches), `ExpenseFormTab` удалён мёртвый `calculateTax` (12 строк) → 60.95%→62.62% stmts. Пороги `58/63/73/73` сохранены, `104 suites 1506/1506` (+1 тест). Коммит push в main.

## v1.19 (2026-09-23) — _app 100% + services/api 24% (global 73.42/58.48/64.74/74.72)
- **Tests:** `pages/_app` 5 (рендер/диспатч auth с token+user/нет токена/ошибка парсинга/showChild) 100% stmts (90% branches), `services/api` 8 (reducerPath/tagTypes/хуки/prepareHeaders store+localStorage+parse error/login/createChecklist/getSalons) 24.09% stmts (3.61%→24%). Пороги `58/63/73/73`, `104 suites 1506/1506`. Коммит push в main.

## v1.18 (2026-09-23) — pages salon-manager 100% + franchiser-manager 100% + salons 95% + index 100% (global 72.61/57.96/64.14/73.84)
- **Tests:** `pages/salon-manager` 7 (загрузка/Header/salon_manager/редиректы super_admin/franchiser/dealer/unknown/login) 100% stmts, `pages/franchiser-manager` 8 (загрузка/franchiser/franchiser_manager/редиректы 4 роли) 100% stmts, `pages/salons` 6 (заголовок/таблица/loading/модалка/создание ok+err/ID) 94.73% stmts, `pages/index` 8 (загрузка/редиректы super_admin/franchiser/dealer/salon_manager/unknown/login) 100% stmts. Пороги `57/62/72/72`, `102 suites 1493/1493`. Коммит push в main.

## v1.17 (2026-09-23) — pages settings 100% + admin 100% + dealer 100% + employees 76% (global 70.37/56.40/63.29/71.45)
- **Tests:** `pages/settings` 6 (Настройки/тема/тёмная/переключение/Назад/редирект) 100% stmts, `pages/admin` 8 (загрузка/супер-админ/редиректы 5 ролей/login) 100% stmts, `pages/dealer` 7 (загрузка/дилер/редиректы 4 роли/login) 100% stmts, `pages/employees` 7 (заголовок/таблица 2 роли/модалки создание+редактирование/создание/ошибка/удаление) 75.51% stmts. Пороги `55/61/70/70`, `98 suites 1464/1464`. Коммит push в main.

## v1.16 (2026-09-23) — pages login 58% + register 80% + profile 86% (global 68.06/55.00/61.98/68.99)
- **Tests:** `pages/login` 8 (форма/ошибка/loading/редиректы 4 роли/переход на register) 57.57% stmts, `pages/register` 5 (форма/ошибка/loading/редирект) 80% stmts, `pages/profile` 7 (рендер/скелетон/ошибка+повтор/сохранение основной+контакты/инициалы/Назад) 86.27% stmts. Пороги `54/60/68/68`, `94 suites 1436/1436`. Коммит push в main.

## v1.15 (2026-09-23) — FranchiserTeamTab 60% + SalonFunnelTab 66% + SalonTeamTab interactions (global 66.28/53.86/60.80/67.12)
- **Tests:** `FranchiserTeamTab` +6 (сохранение планов/копирование квартала/детальная панель/регистрация ok+err/квартал) 49.53%→59.81% stmts, `SalonFunnelTab` +6 (модалка создания/фильтр воронки/горячие сделки/взять лида/ошибка) 54.16%→65.62% stmts, `SalonTeamTab` +5 (фильтры периода/статистика/скидки/прогресс/График) 63.51% stmts, `ExpenseFormTab` фикс таймаута (Popconfirm). Пороги `53/59/66/66`, `91 suites 1416/1416`. Коммит push в main.

## v1.14 (2026-09-23) — DealerReportTab 90.5% + ExpenseFormTab 61% (global 65.76/53.40/60.27/66.55)
- **Tests:** `DealerReportTab` 10 (генерация отчёта/загрузка/предпросмотр 7 блоков/история/генерация модалка/скачать PDF/отправить ok+err/комментарий) 90.47% stmts (67.85% branches), `ExpenseFormTab` 9 (форма поля/loading/загрузка данных/копирование прошлого месяца/сохранение onSave/api/переключение налога/алерт/ошибка) 60.95% stmts (54.36% branches). Пороги `52/58/65/65`, `91 suites 1400/1400`. Коммит push в main.

## v1.13 (2026-09-23) — api kpi/leads 100% + AlertSettings + checklists page (global 64.89/52.10/58.96/65.64)
- **Tests:** `api/kpi.test.ts` 6 (getMyStats/SalonStats/setGoal/getSchedule/createEvent/updateEventStatus 100%), `api/leads.test.ts` 5 (getLeads/createLead/updateLeadStatus/addLeadActivity/getLeadDetails 100%), `AlertSettingsPanel` 4 (рендер секций/switch/checkbox/сохранить+onClose), `FranchiserAlertSettings` 2 (рендер каналов+порогов/сохранить), `pages/checklists` 8 (отображение/загрузка/редирект/ошибка/создание/редактирование/завершить/удалить 90.1% stmts). Пороги `50.5/57/64/63`, `91 suites 1391/1391`. Коммит push в main.

## v1.12 (2026-09-23) — TerritoryMapTab + FranchiserManagerDashboard (global 61.87/50.6/56.07/62.84)
- **Tests:** TerritoryMapTab +4 render (статистика+теплокарта+статусы/поиск/Segmented+карточка фильтры/drill-down красной зоны → Салон 1 детализация). FranchiserManagerDashboard.render +3 (заголовок+метрики+вкладки lazy-табы/модалка сводки/fallback на дефолт при ошибке API). Stmts: MapTab 88.6, FranchiserManagerDashboard 82.8. Пороги `49/53/60/60`, `86 suites 1366/1366`. Коммит push в main.

## v1.11 (2026-09-23) — Territory-табы +16 render-тестов (global 60.05/47.96/54.43/61.02)
- **Tests:** TerritoryFunnelTab +3 (контролы/аномалии/drill-down до менеджеров), TerritoryBenchmarkTab +3 (маржинальность/рейтинг/риски+структура), TerritoryCommunicationsTab +5 (запросы/взять в работу/задачи-модал/история-контакт), TerritoryPlanFactTab +5 (сценарии/Топ-3/период месяц/PDF-ошибка/reopen). План-факт и коммуникации с моками axiosClient (fallback STUB) и antd message. Статус: Benchmark 85.7, Funnel 91, Comms 82.1, PlanFact 75.2 stmts. Пороги `46.5/43/53/59.5`, `85 suites 1359/1359`. Коммит push в main.

## v1.10 (2026-09-23) — DealerAlerts render + ChecklistBoard + employee api (global 53.36/44.02/45.69/54.87)
- **Tests:** DealerAlerts +5 render (bell/popup категории/Empty/Spin/settings save), ChecklistBoard 7 (canCreate/статусы/повторение/сроки/modal create+edit/delete), employee.ts api 4 (getAll/create/update/delete 100%). Пороги `42/44/53/51`, `85 suites 1343/1343`. Коммит push в main.

## v1.9 (2026-09-23) — дэшборды +6 render-тестов (global 50.62/42.78/43.53/51.97)
- **Tests:** SalonManagerDashboard 3 (top-bar success/fallback/tabs), DealerDashboard.render 2, FranchiserDashboard.render 1 (ленивые табы/Suspense, store/axiosClient mocked). Store-тесты DealerDashboard/FranchiserDashboard (50) сохранены в оригинальных файлах. Пороги `41/42/50/48`, `83 suites 1327/1327`, коммит push в main.

## v1.8 (2026-09-23) — 15 тестов на 10 компонентов (global 48.09/41.08/42.08/49.48)
- **Tests:** BackButton/Header/GeographyMap/ThemeProvider 100%, PlanList 6 тестов, SalonManagerWidget 5, charts (ManagerKpiChart/ManagerPlanFactChart ветки «Нет данных», ReportPlanFactChart smoke, SalesDynamicsChart Radio cumulative). Пороги `40/41/47/46`, `80 suites 1321/1321`, `make audit` зелёный, коммит `53ba5e6` push в main.

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
