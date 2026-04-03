# ==============================================================================
# Franchise SaaS Platform - Makefile
# ==============================================================================
# Использование: make <команда>
# ==============================================================================

COMPOSE := docker-compose
COMPOSE_DEV := docker-compose -f docker-compose.dev.yml

.PHONY: help install dev up down build rebuild logs ps shell-backend shell-front clean test-back test-front test lint-back lint-front lint deploy-back deploy-front status

help: ## ��� Показать список всех команд
    @echo "Доступные команды для управления проектом:"
    @grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install: ## ��� Установка зависимостей (локально)
    @echo "--> Установка зависимостей Frontend..."
    cd frontend && npm install
    @echo "--> Установка зависимостей Backend..."
    cd backend && go mod download
    @echo "✅ Зависимости установлены."

up: ## ��� Запуск проекта (Production режим)
    @echo "��� Запуск сервисов (Production)..."
    $(COMPOSE) up -d --build
    @echo "✅ Проект запущен. http://localhost:3000"

dev: ## ���️ Запуск проекта (Development режим)
    @echo "���️ Запуск сервисов (Development)..."
    $(COMPOSE_DEV) up -d --build
    @echo "✅ Проект запущен в режиме разработки. http://localhost:3000"

down: ## ��� Остановка всех сервисов
    @echo "��� Остановка сервисов..."
    $(COMPOSE) down
    $(COMPOSE_DEV) down 2>/dev/null || true
    @echo "✅ Сервисы остановлены."

build: ## ��� Сборка Docker образов
    @echo "��� Сборка образов..."
    $(COMPOSE) build

rebuild: ## ��� Полная пересборка (без кэша)
    @echo "��� Полная пересборка без кэша..."
    $(COMPOSE) build --no-cache
    $(COMPOSE) up -d --force-recreate

logs: ## ��� Просмотр логов (всех сервисов)
    $(COMPOSE) logs -f

logs-back: ## ��� Логи только Backend
    $(COMPOSE) logs -f backend

logs-front: ## ��� Логи только Frontend
    $(COMPOSE) logs -f frontend

ps: ## ��� Статус запущенных контейнеров
    @echo "��� Статус контейнеров:"
    $(COMPOSE) ps

status: ps ## Алиас для ps

shell-backend: ## ��� Войти в shell контейнера Backend
    $(COMPOSE) exec backend /bin/sh

shell-front: ## ��� Войти в shell контейнера Frontend
    $(COMPOSE) exec frontend /bin/sh

shell-db: ## ��� Войти в PostgreSQL
    $(COMPOSE) exec postgres psql -U postgres -d franchise_db

clean: ## ���️ Очистка (контейнеры, тома, кэш)
    @echo "���️ Полная очистка..."
    $(COMPOSE) down -v --remove-orphans
    $(COMPOSE_DEV) down -v --remove-orphans 2>/dev/null || true
    docker system prune -f
    @echo "✅ Очистка завершена."

test: test-back test-front ## ��� Запустить все тесты

test-back: ## ��� Тесты Backend (Go)
    @echo "��� Запуск тестов Backend..."
    cd backend && go test ./... -v

test-front: ## ��� Тесты Frontend (Jest)
    @echo "��� Запуск тестов Frontend..."
    cd frontend && npm test

lint: lint-back lint-front ## ��� Проверка кода линтерами

lint-back: ## ��� Линтинг Backend
    @echo "��� Проверка Backend..."
    cd backend && go fmt ./... && go vet ./...

lint-front: ## ��� Линтинг Frontend
    @echo "��� Проверка Frontend..."
    cd frontend && npm run lint

deploy: build up ## ��� Деплой (сборка + запуск)
    @echo "��� Деплой завершен."

deploy-back: ## ��� Деплой только Backend
    $(COMPOSE) build backend
    $(COMPOSE) up -d backend --no-deps

deploy-front: ## ��� Деплой только Frontend
    $(COMPOSE) build frontend
    $(COMPOSE) up -d frontend --no-deps

