# Deploy — Franchise-Management-SaaS (prod)

## 0. Человек: ключи и доступы (без этого прод НЕ запускать)
<!-- Реестр исключений: .audit-exceptions.yml. Каждый пункт требует человека,
     без срока давности. CI проверяет парность маркеров и реестра. -->
<!-- AUDIT-EXCEPTION(E01): owner-key, см. .audit-exceptions.yml -->
- `JWT_SECRET`: сгенерировать (`openssl rand -hex 32`), положить в `.env` на VPS.
<!-- AUDIT-EXCEPTION(E02): owner-key, см. .audit-exceptions.yml -->
- `DB_PASSWORD`: сгенерировать (`openssl rand -base64 24`), положить в `.env` на VPS.
<!-- AUDIT-EXCEPTION(E03): owner-key, см. .audit-exceptions.yml -->
- `TG_BOT_TOKEN`: создать бота у @BotFather, токен — в `.env`.
<!-- AUDIT-EXCEPTION(E04): owner-key, см. .audit-exceptions.yml -->
- `STARS_PROVIDER_TOKEN`: подключить платежи в @BotFather, токен — в `.env`.
<!-- AUDIT-EXCEPTION(E05): owner-key, см. .audit-exceptions.yml -->
- `VAPID_PUBLIC_KEY` / `VAPID_PRIVATE_KEY`: `npx web-push generate-vapid-keys`, обе — в `.env`.
<!-- AUDIT-EXCEPTION(E06): owner-key, см. .audit-exceptions.yml -->
- `ADMIN_TOKEN`: одноразовый токен владельца для bootstrap super-admin; после настройки удалить из `.env`.
<!-- AUDIT-EXCEPTION(E07): owner-key, см. .audit-exceptions.yml -->
- `ADMIN_TG_IDS`: id получателей алертов через запятую (@userinfobot), в `.env`.
<!-- AUDIT-EXCEPTION(E08): owner-key, см. .audit-exceptions.yml -->
- `UFW`: на VPS `ufw allow 22,80,443/tcp && ufw enable`; БД/Redis слушают только 127.0.0.1 (см. compose).
<!-- AUDIT-EXCEPTION(E12): owner-key, см. .audit-exceptions.yml -->
- `SSH`: только ключи (`PasswordAuthentication no`), root-логин запрещён (`PermitRootLogin no`), желательно нестандартный порт; проверить `ssh -o PreferredAuthentications=password` отказывает.
<!-- AUDIT-EXCEPTION(E09): owner-key, см. .audit-exceptions.yml -->
- `TLS/HSTS/certbot`: выпустить сертификат (`certbot --nginx -d <домен>`), включить 443-блок в `nginx.conf`; без этого HSTS-заголовок не работает.
<!-- AUDIT-EXCEPTION(E10): owner-key, см. .audit-exceptions.yml -->
- `GPG-ключ бэкапов`: импортировать публичный ключ владельца на VPS, шифровать бэкапы; нешифрованные бэкапы с прод-данными не хранить.

## 1. Подготовка
```bash
cp saas-b2b-b2c/.env.example saas-b2b-b2c/.env
# заполните: DB_PASSWORD, JWT_SECRET, NEXT_PUBLIC_API_URL=https://api.example.com
# опционально: NEXT_PUBLIC_WS_URL=wss://api.example.com/ws/alerts, CORS_ALLOWED_ORIGINS, DB_SSLMODE=require
```

## 2. Проверка (fail-closed)
```bash
DB_PASSWORD=CHANGE_ME_DB_PASSWORD JWT_SECRET=CHANGE_ME_JWT_SECRET NEXT_PUBLIC_API_URL=https://api.example.com docker compose -f saas-b2b-b2c/docker-compose.prod.yml config | head -40
NEXT_PUBLIC_API_URL=https://api.example.com npm --prefix saas-b2b-b2c/frontend run build
```

## 3. Поднять
```bash
make prod-up   # или: docker compose -f saas-b2b-b2c/docker-compose.prod.yml up -d --build
make prod-ps
make logs
curl http://localhost:8080/health
curl http://localhost:3000
```

### 3.1 TLS (прод, certbot — за человеком, E09)
```bash
# сертификаты в saas-b2b-b2c/certs (fullchain.pem + privkey.pem)
NGINX_CONF=./nginx.prod.conf docker compose -f saas-b2b-b2c/docker-compose.prod.yml up -d --build
# 80 редиректит на 443; без NGINX_CONF монтируется dev-конфиг без TLS
```

## 4. Бэкап
```bash
make backup    # pg_dump + volume tar в ./backups
# или вручную:
docker compose -f saas-b2b-b2c/docker-compose.prod.yml exec postgres pg_dump -U postgres franchise_db | gzip > backup_$(date +%F).sql.gz
```

## 5. Обновление
```bash
git pull
make prod-up  # пересоберёт backend/frontend (NEXT_PUBLIC_API_URL уже в .env)
```

## 6. Откат
```bash
git log --oneline -5
git reset --hard <prev-tag>
docker compose -f saas-b2b-b2c/docker-compose.prod.yml up -d --build
```
