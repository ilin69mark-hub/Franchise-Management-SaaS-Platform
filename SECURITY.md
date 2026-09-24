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
