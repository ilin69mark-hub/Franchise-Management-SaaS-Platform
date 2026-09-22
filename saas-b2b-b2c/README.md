# Franchise Management SaaS Platform

[![CI](https://github.com/ilin69mark-hub/Franchise-Management-SaaS-Platform/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/ilin69mark-hub/Franchise-Management-SaaS-Platform/actions/workflows/ci.yml)
[![Secrets scan](https://github.com/ilin69mark-hub/Franchise-Management-SaaS-Platform/actions/workflows/secrets-scan.yml/badge.svg?branch=main)](https://github.com/ilin69mark-hub/Franchise-Management-SaaS-Platform/actions/workflows/secrets-scan.yml)

Полнофункциональная SaaS-платформа для управления франчайзинговой сетью. Позволяет франчайзеру управлять дилерами, товарами, заказами и аналитикой через современный веб-интерфейс.

## Содержание

- [Архитектура](#архитектура)
- [Технологический стек](#технологический-стек)
- [Структура проекта](#структура-проекта)
- [Быстрый старт](#быстрый-старт)
- [Ролевая модель](#ролевая-модель)
- [База данных и RLS](#база-данных-и-rls)
- [API](#api)
- [Разработка](#разработка)
- [Деплой](#деплой)

---

## Архитектура

Проект построен по принципу разделения ответственности и состоит из трех основных слоев:

### Frontend
**Директория:** `/frontend`  
**Технологии:** Next.js 14 + React 18 + TypeScript + Ant Design + Redux Toolkit

- SSR для SEO-оптимизации
- Адаптивный дизайн (Desktop/Mobile)
- Интеграция с картами (Yandex Maps)

### Backend
**Директория:** `/backend`  
**Технологии:** Go 1.24 + Gin + GORM

- RESTful API архитектура
- JWT аутентификация
- Role-Based Access Control (RBAC)
- Мультиарендность (Multi-tenancy)
- Кэширование (Redis)

### Infrastructure
**Файлы:** `saas-b2b-b2c/docker-compose.yml`, `saas-b2b-b2c/Dockerfile.*`, `nginx.conf` (корень — Makefile-агрегатор `make dev/test/audit`)

- Контейнеризация (Docker)
- PostgreSQL 15 + Redis 7
- Nginx reverse proxy

---

## Технологический стек

### Frontend
| Технология | Версия |
|------------|--------|
| Next.js | 14.0.0 |
| React | 18.2.0 |
| TypeScript | 5.2.2 |
| Ant Design | 5.10.0 |
| Redux Toolkit | 1.9.7 |
| Tailwind CSS | 3.3.3 |

### Backend
| Технология | Версия |
|------------|--------|
| Go | 1.24 |
| Gin | latest |
| GORM | latest |
| PostgreSQL | 15 |
| Redis | 7 |

---

## Структура проекта

```
saas-b2b-b2c/
├── backend/                     # Серверная часть (Go)
│   ├── cmd/server/              # Точка входа
│   ├── internal/
│   │   ├── cache/              # Redis кэширование
│   │   ├── database/           # Подключение к БД
│   │   ├── handlers/            # HTTP обработчики
│   │   ├── middleware/          # Auth, CORS и др.
│   │   ├── models/              # Модели данных
│   │   ├── repository/          # Слой доступа к данным
│   │   ├── services/            # Бизнес-логика
│   │   └── jobs/                # Фоновые задачи
│   ├── migrations/              # SQL миграции
│   └── Dockerfile.backend
│
├── frontend/                    # Клиентская часть (Next.js)
│   ├── src/
│   │   ├── api/                # Axios инстанс
│   │   ├── components/         # React компоненты
│   │   ├── pages/              # Страницы (Pages Router)
│   │   ├── services/           # API сервисы
│   │   ├── store/              # Redux store
│   │   ├── types/              # TypeScript типы
│   │   └── utils/              # Утилиты
│   └── Dockerfile.frontend
│
├── docker-compose.yml           # Dev-стек (PG+Redis+backend+frontend, hot-reload)
├── docker-compose.prod.yml      # Продакшн (fail-closed: JWT_SECRET/DB_PASSWORD обязательны)
├── nginx.conf                   # Reverse proxy
├── config.yaml                  # Конфигурация Go
└── APIDOCS.md                   # Документация API
```

---

## Быстрый старт

### Предварительные требования

- Docker >= 20.10
- Docker Compose >= 2.0

### Быстрый запуск через Make (рекомендуется)

В **корне репозитория** доступен Makefile-агрегатор — единая точка входа
для dev-стека, тестов, линта и аудита:

```bash
# 1. Поднять полный dev-стек (PG + Redis + backend + frontend)
make dev

# 2. Статус контейнеров
make ps

# 3. Логи (режим follow)
make logs

# 4. Остановить стек
make down
```

Ручной эквивалент (без make, из корня репо):
```bash
docker compose -f saas-b2b-b2c/docker-compose.yml up -d --build
docker compose -f saas-b2b-b2c/docker-compose.yml ps
```

### Адреса сервисов

| Сервис | URL |
|--------|-----|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080/api/v1 |
| PostgreSQL | localhost:5432 |
| Redis | localhost:6379 |

### Переменные окружения

Создайте `.env` в корне проекта:

```env
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=franchise_db
PORT=8080
JWT_SECRET=CHANGE_ME_JWT_SECRET
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## Ролевая модель

Система использует иерархическую структуру управления:

```
franchiser (Владелец)
    └── franchiser_manager (Менеджер)
            └── dealer (Дилер)
                    └── salon_manager (Менеджер салона)
```

### Уровни доступа

| Роль | Описание |
|------|----------|
| `super_admin` | Полный доступ ко всей системе |
| `franchiser` | Владелец франшизы, управление всей сетью |
| `franchiser_manager` | Менеджер франчайзера |
| `dealer` | Дилер (представитель салона) |
| `salon_manager` | Менеджер отдельного салона |

---

## База данных и RLS

### Структура таблиц

Основные миграции (saas-b2b-b2c/backend/migrations/):
- `001_full_schema` — пользователи/планы/салоны/заказы/задачи/goals + индексы
- `002_add_analytics` — user_logs, tenants trial
- `003_add_stage4_fields` — trials
- `004_add_settings` — system_settings
- `005_franchise_manager_foundations` — leads/checklist_templates
- `006_contracts` — contracts
- `007_add_schedule_events` — schedule_events
- `008_add_daily_goals` — daily_goals
- `009_swarm_fixes` — dealer_*/notifications/alerts/users-дополнения
- `010_products_analytics` — products/lost_sales/promotions/category_turnover
- `011_add_analytics_fk_indexes` — FK-индексы аналитики (products/salon)
- `012_gorm_alignment` — выравнивание GORM ↔ SQL (invoices/checklists)
- `013_add_missing_fk_indexes` — остальные FK-индексы (P2-1)

### Row-Level Security (RLS)

Реализована система безопасности на уровне строк:

1. **Иерархия ролей** - поле `managed_by` для связи пользователей
2. **Функция `fn_user_subtree(uuid)`** - возвращает всех подчиненных пользователя (включая самого себя)
3. **Политика `rls_owner`** на таблице `plans`:
   - `super_admin` и `franchiser` видят все строки
   - Остальные роли видят только строки, где `owner_user_id` в их поддереве

### Примеры RLS-запросов

```sql
-- Установить текущего пользователя
SET LOCAL app.current_user_id = 'uuid- пользователя';

-- Запрос данных с учетом RLS
SELECT * FROM plans;
```

---

## API

### Базовый URL

```
Development: http://localhost:8080/api/v1
Production:  https://api.yourdomain.com/api/v1
```

### Аутентификация

Большинство эндпоинтов требуют Bearer Token:

```
Authorization: Bearer <jwt_token>
```

### Основные эндпоинты

| Метод | Эндпоинт | Описание | Доступ |
|-------|----------|----------|--------|
| POST | `/auth/register` | Регистрация | Публичный |
| POST | `/auth/login` | Вход | Публичный |
| GET | `/auth/me` | Профиль | Авторизованный |
| POST | `/auth/refresh` | Обновление токена | Авторизованный |
| GET | `/dealers` | Список дилеров | Franchiser |
| GET | `/products` | Список товаров | Все роли |
| POST | `/orders` | Создание заказа | Dealer, Admin |

### Примеры запросов

**Регистрация:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword123",
    "role": "dealer",
    "first_name": "Иван",
    "last_name": "Иванов"
  }'
```

**Вход:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "securePassword123"}'
```

**Ответ:**
```json
{
  "user": {"id": "uuid", "email": "user@example.com", "role": "dealer"},
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### Чеклисты

| Метод | Эндпоинт | Описание |
|-------|----------|----------|
| GET | `/checklists` | Список чеклистов |
| GET | `/checklists/:id` | Чеклист по ID |
| POST | `/checklists` | Создать чеклист |
| PUT | `/checklists/:id` | Обновить чеклист |
| DELETE | `/checklists/:id` | Удалить чеклист |
| POST | `/checklists/:id/complete` | Завершить чеклист |

Полная документация API: [APIDOCS.md](./APIDOCS.md)

---

## Разработка

### Локальная разработка (без Docker)

**Backend:**
```bash
cd backend
go mod download
go run cmd/server/main.go
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

### Docker для разработки

```bash
# С hot-reload (dev-стек из корня; это тот же стек, что поднимает make dev)
docker compose -f saas-b2b-b2c/docker-compose.yml up -d
```

### Тестирование

**Backend:**
```bash
cd backend

# Unit тесты (быстрые)
make test

# PostgreSQL интеграционные тесты
make test-pg

# Все тесты с покрытием
make test-coverage
```

**Frontend:**
```bash
cd frontend

# Тесты с покрытием
npm test -- --coverage

# Тесты в watch режиме
npm run test:watch
```

**CI/CD:**
- Тесты запускаются автоматически при push в main/develop
- Coverage отчёты доступны в Codecov
- Артефакты сохраняются в GitHub Actions

### Стандарты кода

**Backend (Go):**
```bash
go fmt ./...
go vet ./...
go test ./...
```

**Frontend:**
```bash
npm run lint
npm test
npm run build
```

### Git Workflow

```bash
git checkout -b feature/new-feature
git commit -m "feat: add new feature"
git push origin feature/new-feature
```

---

## Деплой

### Подготовка

1. Измените пароли и секреты в `.env`
2. Настройте домен в `nginx.conf`
3. Настройте SSL (Let's Encrypt)

### Сборка и запуск

```bash
git clone https://github.com/ilin69mark-hub/Franchise-Management-SaaS-Platform.git
cd Franchise-Management-SaaS-Platform
cp saas-b2b-b2c/.env.example saas-b2b-b2c/.env && nano saas-b2b-b2c/.env
# dev
make dev   # или: docker compose -f saas-b2b-b2c/docker-compose.yml up -d --build
# prod (fail-closed: требует DB_PASSWORD, JWT_SECRET, NEXT_PUBLIC_API_URL)
DB_PASSWORD=... JWT_SECRET=... NEXT_PUBLIC_API_URL=https://api.example.com make prod-up
# проверка: DB_PASSWORD=test JWT_SECRET=test NEXT_PUBLIC_API_URL=https://api.example.com docker compose -f saas-b2b-b2c/docker-compose.prod.yml config
```

### Мониторинг

```bash
make logs   # или: docker compose -f saas-b2b-b2c/docker-compose.yml logs -f
docker stats
```

### Бэкап (прод)

```bash
# БД (pg_dump, без даунтайма)
docker compose -f saas-b2b-b2c/docker-compose.prod.yml exec postgres pg_dump -U postgres franchise_db | gzip > backup_$(date +%F).sql.gz
# Восстановление
gunzip < backup_*.sql.gz | docker compose -f saas-b2b-b2c/docker-compose.prod.yml exec -T postgres psql -U postgres franchise_db
# Volume (на хосте)
docker run --rm -v saas-b2b-b2c_postgres_data:/volume -v $(pwd):/backup alpine tar czf /backup/pgdata_$(date +%F).tar.gz -C / volume
```

---

## Устранение неполадок

**Порт занят:**
```bash
lsof -i :3000
lsof -i :8080
```

**Пересоздание БД (удаляет все данные):**
```bash
docker compose -f saas-b2b-b2c/docker-compose.yml down -v
docker compose -f saas-b2b-b2c/docker-compose.yml up --build
```

**Очистка зависимостей:**
```bash
# Frontend
cd frontend && rm -rf node_modules package-lock.json && npm install

# Backend
cd backend && go mod tidy
```

---

## Лицензия

Copyright © 2024 Franchise SaaS Platform. Все права защищены.
## ⚠️ Обновление после переписывания истории (критично для всех)

21 сент. 2026 г. в репозитории была **переписана вся git-история** (две волны: чистка секретов + замена структуры). Это затронуло **все ветки**. Если у вас есть хотя бы один клон этого репозитория (локальный, fork, backup) — сделайте следующее:

```bash
git fetch --all --force
git reset --hard origin/<ваша_ветка>
git clean -fdx
```

**НЕ делайте** `git pull` обычным способом — история переписана, обычный pull приведёт к конфликтам/дублированию истории. Используйте `--force` (fetch + hard reset как выше) или просто **переклонируйте** репозиторий заново:

```bash
git clone https://github.com/ilin69mark-hub/Franchise-Management-SaaS-Platform.git
```

Ветки-«призраки» (старые `feature/*`, `qwen-code-*` и т.п.) удалены намеренно после чистки — не пытайтесь их восстановить. Полный откат истории при необходимости делается из бандла-бэкапа (150 МБ, хранится отдельно от репозитория).
