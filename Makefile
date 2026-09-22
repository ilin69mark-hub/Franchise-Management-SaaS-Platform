# Franchise-Management-SaaS — корневой Makefile-агрегатор
#
# Единая точка входа для dev-стека, тестов, линта и аудита всего репо.
# Тяжёлая логика делегируется в под-Makefile (saas-b2b-b2c/backend, saas-b2b-b2c/frontend)
# и в docker-compose — здесь НЕ дублируются, поэтому правки живут в одном месте.
#
# Быстрый старт:
#   make dev            # поднять весь dev-стек (Postgres + Redis + backend + frontend)
#   make ps             # статус контейнеров
#   make test           # все тесты: backend (go) + frontend (jest + tsc)
#   make lint           # gofmt/vet (backend) + ESLint/tsc (frontend)
#   make build          # прод-сборки backend + frontend
#   make audit          # gitleaks-скан + полный регресс (зелёный CI-эквивалент)
#   make help           # список всех целей
#
# Требования: docker (+ compose v2), Go >= 1.22, Node >= 18.

SHELL := /bin/zsh
.DEFAULT_GOAL := help

COMPOSE      := docker compose
COMPOSE_DEV  := -f saas-b2b-b2c/docker-compose.yml
COMPOSE_PROD := -f saas-b2b-b2c/docker-compose.prod.yml

BACKEND  := $(MAKE) -C saas-b2b-b2c/backend
FRONTEND := $(MAKE) -C saas-b2b-b2c/frontend

.PHONY: help dev up down ps logs prod-up prod-ps \
	backend backend-test backend-test-pg backend-test-all backend-lint \
	frontend frontend-test frontend-typecheck frontend-lint frontend-build \
	test lint build clean audit

## ---------------------------------------------------------------------------
## Docker: dev-стек (PG + Redis + backend + frontend)
## ---------------------------------------------------------------------------

dev: ## Поднять весь dev-стек (build + up)
	$(COMPOSE) $(COMPOSE_DEV) up -d --build
	@printf '\n\033[1;32m✔ Dev-стек поднят:\033[0m\n'
	@printf '  Frontend  → http://localhost:3000\n'
	@printf '  Backend   → http://localhost:8080/api/v1\n'
	@printf '  Postgres  → localhost:5432\n'
	@printf '  Redis     → localhost:6379\n\n'

up: ## Поднять dev-стек (без пересборки)
	$(COMPOSE) $(COMPOSE_DEV) up -d
	@printf '\n\033[1;32m✔ Контейнеры подняты\033[0m\n'

down: ## Остановить и удалить контейнеры dev-стека
	$(COMPOSE) $(COMPOSE_DEV) down

ps: ## Статус контейнеров
	$(COMPOSE) $(COMPOSE_DEV) ps

logs: ## Хвост логов всех контейнеров
	$(COMPOSE) $(COMPOSE_DEV) logs -f --tail=100

## ---------------------------------------------------------------------------
## Docker: прод-стек (docker-compose.prod.yml, fail-closed на секретах)
## ---------------------------------------------------------------------------

prod-up: ## Поднять прод-стек (требует JWT_SECRET + DB_PASSWORD в окружении)
	$(COMPOSE) $(COMPOSE_PROD) up -d --build

prod-ps: ## Статус прод-контейнеров
	$(COMPOSE) $(COMPOSE_PROD) ps

## ---------------------------------------------------------------------------
## Backend (Go) — делегирование в saas-b2b-b2c/backend/Makefile
## ---------------------------------------------------------------------------

backend: ## Полный backend-цикл: build + vet + lint + unit-тесты
	$(BACKEND) test
	$(BACKEND) lint

backend-test: ## Быстрые unit-тесты backend (go test -short, без БД)
	$(BACKEND) test

backend-test-pg: ## Интеграционные тесты на реальном PostgreSQL (docker-compose.test.yml)
	$(BACKEND) test-pg

backend-test-all: ## Все тесты backend (unit + интеграционные)
	$(BACKEND) test-all

backend-lint: ## gofmt-проверка + go vet
	$(BACKEND) lint

## ---------------------------------------------------------------------------
## Frontend (Next.js/TS) — делегирование в saas-b2b-b2c/frontend/Makefile
## ---------------------------------------------------------------------------

frontend: ## Полный frontend-цикл: install + lint + typecheck + test + build
	$(FRONTEND) install
	$(FRONTEND) lint
	$(FRONTEND) typecheck
	$(FRONTEND) test
	$(FRONTEND) build

frontend-test: ## Jest-тесты frontend
	$(FRONTEND) test

frontend-typecheck: ## Проверка типов (tsc --noEmit)
	$(FRONTEND) typecheck

frontend-lint: ## ESLint frontend
	$(FRONTEND) lint

frontend-build: ## Продакшн-сборка (next build)
	$(FRONTEND) build

## ---------------------------------------------------------------------------
## Агрегаты: тесты / линт / сборка
## ---------------------------------------------------------------------------

test: ## Все тесты репо: backend (go) + frontend (jest)
	$(BACKEND) test
	$(FRONTEND) test

lint: ## Весь линт: gofmt/vet (backend) + ESLint/tsc (frontend)
	$(BACKEND) lint
	$(FRONTEND) lint
	$(FRONTEND) typecheck

build: ## Прод-сборки backend + frontend
	$(BACKEND) build
	$(FRONTEND) build

clean: ## Остановить стек и почистить артефакты сборки
	$(COMPOSE) $(COMPOSE_DEV) down
	$(BACKEND) clean
	$(FRONTEND) clean

backup: ## Бэкап БД (pg_dump) + volume в ./backups
	@mkdir -p backups
	$(COMPOSE) $(COMPOSE_PROD) exec -T postgres pg_dump -U postgres franchise_db | gzip > backups/backup_$(shell date +%F_%H%M).sql.gz
	@ls -lh backups/backup_*.sql.gz | tail -1
	@printf '\n\033[1;32m✔ Бэкап сохранён в backups/\033[0m\n'

restore: ## Восстановление из последнего бэкапа (backups/backup_*.sql.gz)
	@ls backups/backup_*.sql.gz 2>/dev/null | tail -1 | xargs -I {} sh -c 'gunzip < {} | $(COMPOSE) $(COMPOSE_PROD) exec -T postgres psql -U postgres franchise_db && echo "✔ Восстановлено из {}"'

audit: ## gitleaks-скан секретов во всей истории + полный регресс (CI-дубликат)
	@printf '\n\033[1;33m[1/5] Gitleaks-скан (история + HEAD)...\033[0m\n'
	gitleaks detect --source . --redact --log-opts='--all'
	@printf '\n\033[1;33m[2/5] Backend: build + vet + unit-тесты...\033[0m\n'
	$(BACKEND) test
	@printf '\n\033[1;33m[3/5] Backend lint...\033[0m\n'
	$(BACKEND) lint
	@printf '\n\033[1;33m[4/5] Frontend: lint + typecheck...\033[0m\n'
	$(FRONTEND) lint
	$(FRONTEND) typecheck
	@printf '\n\033[1;33m[5/5] Frontend: jest...\033[0m\n'
	$(FRONTEND) test
	@printf '\n\033[1;32m✔ Аудит пройден: секретов нет, регресс зелёный\033[0m\n'

## ---------------------------------------------------------------------------
## Справка
## ---------------------------------------------------------------------------

help: ## Показать список всех целей
	@printf '\n\033[1;36mFranchise-Management-SaaS — доступные цели\033[0m\n'
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "} {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@printf '\nБыстрый старт: make dev  →  затем открыть http://localhost:3000\n\n'
