# Franchise Management SaaS Platform

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
**Файлы:** `docker-compose.yml`, `Dockerfile.*`, `nginx.conf`

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
├── docker-compose.yml           # Продакшн
├── docker-compose.dev.yml       # Разработка
├── nginx.conf                   # Reverse proxy
├── config.yaml                  # Конфигурация Go
└── APIDOCS.md                   # Документация API
```

---

## Быстрый старт

### Предварительные требования

- Docker >= 20.10
- Docker Compose >= 2.0

### Запуск

```bash
# Клонирование и запуск
docker-compose up -d --build

# Проверка статуса
docker-compose ps
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

Основные миграции:
- `001_full_schema` - основная схема (пользователи, продукты, заказы)
- `002_add_analytics` - аналитика
- `005_franchise_manager_foundations` - иерархия менеджеров
- `006_contracts` - контракты
- `007_add_schedule_events` - события расписания
- `008_add_daily_goals` - дневные цели

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
# С hot-reload
docker-compose -f docker-compose.dev.yml up -d
```

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
git clone <repo-url>
cd saas-b2b-b2c
nano .env
docker-compose up -d --build
```

### Мониторинг

```bash
docker-compose logs -f
docker stats
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
docker-compose down -v
docker-compose up --build
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