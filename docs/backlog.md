# Бэклог покрытия — Franchise Management SaaS (автономный режим)

> Создан: 2026-09-23 по команде «Да забей бэклог задачами и следуй по нему без моего разрешения»
> Текущее покрытие: **76.78/61.55/70.09/78.15** (lines/branches/functions/statements) — 104 suites 1524/1524
> Цель: **80/65/75/80** (lines 80% — следующий milestone, затем 85%)

## Приоритет P1 — 0% файлы (быстрый лифт, pages)
- [ ] **B1** `pages/_app` уже 100% — done
- [ ] **B2** `services/api` 73.49% → 80% (503 строки, +6.5% = +32 линии) — мутации `getProfile`/`logout`/`readNotification` + `providesTags` ветви
- [ ] **B3** `CommunicationsTab` 58.20% → 65% (527 строк, +6.8% = +36 линий) — `handleTaskStatusChange`/`filteredTasks` + `requestsTableColumns`/`budgetTableColumns`
- [ ] **B4** `FranchiserTeamTab` 84.11% → 90% (487 строк, +5.9% = +29 линий) — `renderDetailPanel`/`getRowColor` branches/`handleAddManager` error
- [ ] **B5** `SalonTeamTab` 85.13% → 90% (391 строк, +4.9% = +19 линий) — `DeviationIndicator` null/`formatMoney`/`fetchHistory` error

## Приоритет P2 — 60-70% (средний лифт, tabs)
- [ ] **C1** `ExpenseFormTab` 82.82% → 90% (500 строк, +7.2% = +36 линий) — `handleSave` total/`handleImport` FormData/`fields`/`hasPrevMonthData`
- [ ] **C2** `SalonFunnelTab` 69.79% → 75% (514 строк, +5.2% = +27 линий) — `handleCreateLead` budget/`handleStatusChange`/`handleAssignLead`/`stageLeads` filter
- [ ] **C3** `SalonTeamTab` уже 85% — см. B5
- [ ] **C4** `FranchiserTeamTab` уже 84% — см. B4
- [ ] **C5** `TerritoryFunnelTab` 91% — done
- [ ] **C6** `TerritoryBenchmarkTab` 85% — done

## Приоритет P3 — 70-80% (добивка, services/pages)
- [ ] **D1** `services/planApi` 18% → 50% (106 строк) — `fetchBaseQuery` + endpoints
- [ ] **D2** `services/goalApi` 18% → 50% (63 строки)
- [ ] **D3** `services/userApi` 18% → 50% (56 строк)
- [ ] **D4** `pages/salons` 94% — done
- [ ] **D5** `pages/index` 100% — done

## План волн (без ожидания ОК)

| Волна | Файлы | Цель | Пороги |
|-------|-------|------|--------|
| v1.29 | B2 `services/api` 73%→80% + C1 `ExpenseFormTab` 82%→85% | 77.5% lines | 61/69/76/78 |
| v1.30 | B3 `CommunicationsTab` 58%→65% + C2 `SalonFunnelTab` 69%→75% | 78.5% lines | 62/70/77/79 |
| v1.31 | B4 `FranchiserTeamTab` 84%→90% + B5 `SalonTeamTab` 85%→90% | 79.5% lines | 63/71/78/80 |
| v1.32 | D1-D3 `services/*Api` 18%→50% + фикс `gitleaks`/`tsc` | 80% lines | 64/72/79/81 |

Каждая волна: `npx tsc --noEmit` 0, `npx jest` 1524+ → зелёный, `jest --coverage` пороги, `git commit` + `push origin main`, `CHANGELOG.md` + `mem_save`.

> Режим: автономный, без запроса разрешения. При `confidence <0.7` или `supersedes/conflicts_with` для `architecture/policy/decision` — спросить.
