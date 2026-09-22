# Deploy — Franchise-Management-SaaS (prod)

## 1. Подготовка
```bash
cp saas-b2b-b2c/.env.example saas-b2b-b2c/.env
# заполните: DB_PASSWORD, JWT_SECRET, NEXT_PUBLIC_API_URL=https://api.example.com
# опционально: NEXT_PUBLIC_WS_URL=wss://api.example.com/ws/alerts, CORS_ALLOWED_ORIGINS, DB_SSLMODE=require
```

## 2. Проверка (fail-closed)
```bash
DB_PASSWORD=test JWT_SECRET=test1234567890 NEXT_PUBLIC_API_URL=https://api.example.com docker compose -f saas-b2b-b2c/docker-compose.prod.yml config | head -40
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
