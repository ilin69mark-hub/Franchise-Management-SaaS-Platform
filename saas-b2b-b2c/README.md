Вот подробный, объединенный `README.md` для всего проекта. Я разбил его на логические блоки, объяснил зону ответственности каждого компонента и добавил пошаговые инструкции.

---

# 🏢 Franchise Management SaaS Platform

Полнофункциональная SaaS-платформа для управления франчайзинговой сетью. Позволяет франчайзеру управлять дилерами, товарами, заказами и аналитикой через современный веб-интерфейс.

## 📋 Содержание

- [Архитектура проекта](#-архитектура-проекта)
- [Технологический стек](#-технологический-стек)
- [Структура проекта](#-структура-проекта)
- [Установка и запуск](#-установка-и-запуск)
- [API Документация](#-api-документация)
- [Разработка](#-разработка)
- [Деплой](#-деплой)

---

## 🏗 Архитектура проекта

Проект построен по принципу разделения ответственности (Separation of Concerns) и состоит из трех основных слоев:

### 1. Frontend (Слой представления)
**Директория:** `/frontend`
**Технология:** Next.js 14 + React 18 + TypeScript + Ant Design

**Зона ответственности:**
- Отрисовка пользовательского интерфейса (UI)
- Взаимодействие с пользователем (формы, кнопки, навигация)
- Валидация данных на клиенте
- Управление состоянием приложения (Redux Toolkit)
- Маршрутизация на стороне клиента

**Ключевые особенности:**
- Server-Side Rendering (SSR) для SEO-оптимизации
- Адаптивный дизайн (Desktop/Mobile)
- Интеграция с картами (Yandex Maps)

### 2. Backend (Слой бизнес-логики)
**Директория:** `/backend`
**Технология:** Go 1.24 + Gin + GORM

**Зона ответственности:**
- Обработка HTTP-запросов
- Бизнес-логика приложения
- Валидация данных на сервере
- Аутентификация и авторизация (JWT)
- Взаимодействие с базой данных
- Кэширование данных (Redis)

**Ключевые особенности:**
- RESTful API архитектура
- Role-Based Access Control (RBAC)
- Мультиарендность (Multi-tenancy)

### 3. Infrastructure (Слой инфраструктуры)
**Файлы:** `docker-compose.yml`, `Dockerfile.*`, `nginx.conf`

**Зона ответственности:**
- Контейнеризация приложения
- Управление базами данных
- Обратное проксирование (Nginx)
- Сетевое взаимодействие между сервисами

---

## 🛠 Технологический стек

### Frontend
| Технология | Версия | Назначение |
|------------|--------|------------|
| Next.js | 14.0.0 | SSR-фреймворк |
| React | 18.2.0 | UI библиотека |
| TypeScript | 5.2.2 | Типизация |
| Ant Design | 5.10.0 | UI компоненты |
| Redux Toolkit | 1.9.7 | State management |
| Tailwind CSS | 3.3.3 | Стилизация |
| Axios | 1.13.4 | HTTP-клиент |
| React Hook Form | 7.47.0 | Работа с формами |
| Recharts | 2.8.0 | Графики и диаграммы |

### Backend
| Технология | Версия | Назначение |
|------------|--------|------------|
| Go | 1.24 | Язык программирования |
| Gin | latest | HTTP-фреймворк |
| GORM | latest | ORM для работы с БД |
| JWT | latest | Аутентификация |
| Viper | latest | Конфигурация |
| Air | latest | Hot-reload для разработки |

### Infrastructure
| Технология | Версия | Назначение |
|------------|--------|------------|
| Docker | latest | Контейнеризация |
| Docker Compose | 3.8 | Оркестрация |
| PostgreSQL | 15-alpine | Основная БД |
| Redis | 7-alpine | Кэширование |
| Nginx | latest | Reverse proxy |

---

## 📁 Структура проекта

```
saas-b2b-b2c/
├── 📂 backend/                 # Серверная часть
│   ├── 📂 cmd/                 # Точки входа
│   │   └── 📄 server/main.go   # Главный файл запуска
│   ├── 📂 config/              # Конфигурация
│   ├── 📂 controllers/         # Контроллеры (обработчики)
│   ├── 📂 middleware/           # Middleware (аутентификация и др.)
│   ├── 📂 migrations/           # SQL миграции
│   ├── 📂 models/               # Модели данных (ORM)
│   ├── 📂 repositories/         # Слой доступа к данным
│   ├── 📂 routes/               # Определение маршрутов
│   ├── 📂 services/             # Бизнес-логика
│   ├── 📂 tests/                # Тесты
│   ├── 📂 utils/                # Утилиты
│   ├── 📄 go.mod                # Зависимости Go
│   └── 📄 Dockerfile.backend     # Docker-сборка бэкенда
│
├── 📂 frontend/                # Клиентская часть
│   ├── 📂 public/               # Статические файлы
│   ├── 📂 src/
│   │   ├── 📂 app/              # Next.js App Router
│   │   ├── 📂 components/       # React компоненты
│   │   ├── 📂 hooks/            # Кастомные хуки
│   │   ├── 📂 pages/            # Страницы (Pages Router)
│   │   ├── 📂 services/         # API сервисы (axios)
│   │   ├── 📂 store/            # Redux store
│   │   ├── 📂 types/            # TypeScript типы
│   │   └── 📂 utils/            # Утилиты
│   ├── 📄 package.json          # Зависимости npm
│   ├── 📄 tailwind.config.js    # Конфигурация Tailwind
│   └── 📄 Dockerfile.frontend   # Docker-сборка фронтенда
│
├── 📄 docker-compose.yml        # Продакшн конфигурация
├── 📄 docker-compose.dev.yml    # Dev конфигурация
├── 📄 nginx.conf                # Конфигурация Nginx
├── 📄 .env                      # Переменные окружения (НЕ коммитить!)
├── 📄 config.yaml               # Конфигурация для Go
└── 📄 README.md                 # Этот файл
```

---

## 🚀 Установка и запуск

### Предварительные требования

- **Docker** >= 20.10
- **Docker Compose** >= 2.0
- **Node.js** >= 18.0 (для локальной разработки фронтенда)
- **Go** >= 1.24 (для локальной разработки бэкенда)

---

### Вариант 1: Быстрый старт (Docker) ⚡

> **Рекомендуется для тестирования и продакшена**

#### Шаг 1: Клонирование репозитория

```bash
git clone <your-repo-url>
cd saas-b2b-b2c
```

#### Шаг 2: Настройка переменных окружения

Создайте файл `.env` в корне проекта:

```env
# Database
DB_USER=postgres
DB_PASSWORD=your_secure_password_here
DB_NAME=franchise_db

# Server
PORT=8080

# JWT
JWT_SECRET=CHANGE_ME_JWT_SECRET

# Frontend (для сборки)
NEXT_PUBLIC_API_URL=http://localhost:8080
```

#### Шаг 3: Запуск через Docker Compose

```bash
# Сборка и запуск всех сервисов
docker-compose up -d --build

# Просмотр логов
docker-compose logs -f

# Остановка сервисов
docker-compose down
```

#### Шаг 4: Проверка работоспособности

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080/api/v1
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

---

### Вариант 2: Локальная разработка (без Docker) 🔧

> **Рекомендуется для активной разработки**

#### Backend (Go)

```bash
# Переход в директорию бэкенда
cd backend

# Установка зависимостей
go mod download

# Запуск миграций (опционально, если не используете Docker PostgreSQL)
# Создайте БД вручную и примените миграции из /migrations

# Запуск сервера с hot-reload
# Требуется установленное окружение из config.yaml или .env
go run cmd/server/main.go

# Или с использованием Air (hot-reload)
air
```

#### Frontend (Next.js)

```bash
# Переход в директорию фронтенда
cd frontend

# Установка зависимостей
npm install

# Создание .env.local для API
echo "NEXT_PUBLIC_API_URL=http://localhost:8080" > .env.local

# Запуск в режиме разработки
npm run dev
```

#### База данных (PostgreSQL + Redis)

Вы можете запускать только инфраструктурные сервисы через Docker:

```bash
docker-compose up -d postgres redis
```

---

### Вариант 3: Режим разработки (Docker + Hot Reload) 🔄

```bash
# Используйте специальный compose-файл для разработки
docker-compose -f docker-compose.dev.yml up -d

# Это позволит:
# - Монтировать исходный код как volumes
# - Использовать Air для Go (hot-reload)
# - Использовать next dev для фронтенда
```

---

## 📖 API Документация

### Базовый URL
- Development: `http://localhost:8080/api/v1`
- Production: `https://api.yourdomain.com/api/v1`

### Аутентификация

Большинство эндпоинтов требуют Bearer Token в заголовке:

```
Authorization: Bearer <your_jwt_token>
```

### Основные эндпоинты

| Метод | Эндпоинт | Описание | Доступ |
|-------|----------|----------|--------|
| POST | `/auth/register` | Регистрация | Публичный |
| POST | `/auth/login` | Вход в систему | Публичный |
| GET | `/auth/me` | Профиль пользователя | Авторизованный |
| POST | `/auth/refresh` | Обновление токена | Авторизованный |
| GET | `/dealers` | Список дилеров | Franchiser |
| GET | `/products` | Список товаров | Все роли |
| POST | `/orders` | Создание заказа | Dealer, Admin |

### Примеры запросов

#### Регистрация

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword123",
    "role": "dealer",
    "tenant_id": "tenant-001",
    "first_name": "Иван",
    "last_name": "Иванов"
  }'
```

#### Вход в систему

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword123"
  }'
```

**Ответ:**

```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "role": "dealer"
  },
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

---

## 💻 Разработка

### Стандарты кода

#### Backend (Go)
- Следуйте стандартам `gofmt` и `go vet`
- Используйте комментарии для экспортируемых функций
- Пишите unit-тесты для критичной бизнес-логики

```bash
# Форматирование кода
go fmt ./...

# Запуск тестов
go test ./...

# Проверка кода
go vet ./...
```

#### Frontend (TypeScript/React)
- Используйте строгую типизацию (strict mode)
- Разделяйте компоненты на Presentational и Container
- Используйте функциональные компоненты и хуки

```bash
# Линтинг
npm run lint

# Тесты
npm test

# Сборка
npm run build
```

### Git Workflow

```bash
# Создание ветки для фичи
git checkout -b feature/new-feature

# Коммит с понятным описанием
git commit -m "feat: add user profile page"

# Push и создание Pull Request
git push origin feature/new-feature
```

---

## 🚢 Деплой

### Подготовка к продакшену

1. **Измените пароли и секреты** в `.env`:
   ```env
   DB_PASSWORD=<strong_password>
   JWT_SECRET=<very_long_random_string>
   ```

2. **Настройте домен** в `nginx.conf` и `NEXT_PUBLIC_API_URL`

3. **Настройте SSL** сертификаты (Let's Encrypt)

### Сборка и запуск на сервере

```bash
# Клонирование репозитория
git clone <repo-url>
cd saas-b2b-b2c

# Создание .env с продакшен-переменными
nano .env

# Сборка и запуск
docker-compose -f docker-compose.yml up -d --build

# Проверка статуса
docker-compose ps
```

### Мониторинг

```bash
# Логи всех сервисов
docker-compose logs -f

# Логи конкретного сервиса
docker-compose logs -f backend
docker-compose logs -f frontend

# Использование ресурсов
docker stats
```

---

## 🔧 Устранение неполадок

### Частые проблемы

#### 1. Порт уже занят
```bash
# Проверка занятых портов
lsof -i :3000
lsof -i :8080

# Остановка конфликтующих сервисов
docker-compose down
```

#### 2. Ошибки базы данных
```bash
# Пересоздание БД (ВНИМАНИЕ: удалит все данные!)
docker-compose down -v
docker-compose up --build
```

#### 3. Проблемы с зависимостями
```bash
# Frontend: очистка кэша
cd frontend
rm -rf node_modules package-lock.json
npm install

# Backend: обновление зависимостей
cd backend
go mod tidy
```

---

## 📞 Поддержка

- **Документация API**: См. `APIDOCS.md`
- **Коллекция Postman**: Импортируйте `postman-collection.json`
- **Issues**: Создавайте в репозитории

---

## 📜 Лицензия

Copyright © 2024 Franchise SaaS Platform. Все права защищены.

---

**Удачной разработки! 🎉**