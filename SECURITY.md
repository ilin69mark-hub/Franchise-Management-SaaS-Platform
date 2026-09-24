# Security Policy

## Reporting
Found a security issue? Email `security@example.com` (replace) — do not open public issue.

## Secrets
- `.env` never committed (`.gitignore`), use `.env.example` as template.
- Prod secrets fail-closed: `docker-compose.prod.yml` uses `${VAR:?must be set}` for `DB_PASSWORD`, `JWT_SECRET`, `NEXT_PUBLIC_API_URL`.
- `gitleaks` scans every push (`--log-opts=--all`, `.gitleaks.toml` allowlist only for `unsafe-default-*` placeholders).
- Local pre-commit: `make hooks-install` → `.githooks/pre-commit` runs
  `gitleaks protect --staged` + `gofmt` on staged `.go` (fast, <5s).
  Bypass only deliberately: `git commit --no-verify`.

## Auth
- JWT `HS256`, `JWT_SECRET` required in prod (`config.go` fallback only for dev).
- `POST /auth/login` 5/min, `POST /auth/register` 10/min (`main.go` RateLimit), global 100/min.
- RBAC via `middleware.RequireRole`, IDOR checks in `kpi_handler`, `user_handler`.

## Network
- Prod `postgres/redis` bind `127.0.0.1` only (`docker-compose.prod.yml`), `nginx` fronts 80/443 with HSTS/CSP (`nginx.conf`).
- CORS `Vary: Origin` (`middleware/cors.go`), allowlist via `CORS_ALLOWED_ORIGINS`.

## Data
- Migrations `014` business settings in `system_settings` (no hardcode).
- Daily goals `UNIQUE NULLS NOT DISTINCT` (PG15) to avoid duplicates.
- Email unique регистронезависимо: `idx_users_email_lower` (миграция `019`), нормализация на регистрации/логине.
- Границы day/month — в зоне тенанта (`tenants.timezone`, миграция `020`), fallback UTC.

## Passwords
- Требования: 12..72 символа (bcrypt >72 байт обрезает), проверка HIBP по локальной части email.
- HIBP fail-open **по API**: если k-Anonymity endpoint недоступен, вход не блокируется, но событие логируется (`hibp_unchecked`) и видно в аудите. Приемлемо: HIBP не является контролем доступа, outage у стороннего сервиса не должен ронять вход.
- CAPTCHA-хук (`verifyCaptchaToken`) включается на пороге половины lockout-окна; без `CAPTCHA_SECRET`/`CAPTCHA_VERIFY_URL` — пропуск (исключение `E13`).

## Известные принятые риски
- **Timing-oracle на входе.** Неизвестный email и неверный пароль обрабатываются по-разному по времени (второй путь включает bcrypt). Защита — rate limit + lockout по паре IP+email + CAPTCHA-порог; полное выравнивание (dummy-bcrypt для несуществующего email) сознательно не сделано: это добавляет фиксированную задержку всем и усложняет код, а практическая польза при действующем rate limit ограничена. Решение: принято.
- **HIBP fail-open** — см. «Passwords» выше.
- **Человекозависимые пункты** (сертификаты TLS, GPG-бэкапы, ESO-секреты, SSH-харднинг, CAPTCHA-провайдер, `REDIS_PASSWORD`) вынесены в `.audit-exceptions.yml` (E01–E14) и проверяются скриптом `make audit-exceptions`.
