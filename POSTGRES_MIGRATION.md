# PostgreSQL Migration Plan

Nedetid: ~2-3 sek (kun container swap i steg 4).
PostgreSQL er fyldt med data FØR appen switcher.

## Pre-deploy (GitHub)

1. Gå til repo Settings → Secrets → New repository secret
2. Tilføj `DB_PASSWORD` med et sikkert password (fx `openssl rand -base64 24`)

## Deploy dag

### Steg 1 — Forbered serveren

```bash
ssh root@46.224.144.214
```

Opret `.env` fil:
```bash
echo "DB_PASSWORD=<dit-password>" > /root/itu-minitwit/.env
```

### Steg 2 — Start PostgreSQL manuelt (appen kører stadig på SQLite)

```bash
docker network create minitwit-net 2>/dev/null || true

docker run -d --name minitwit-postgres \
  --network minitwit-net \
  -e POSTGRES_DB=minitwit \
  -e POSTGRES_USER=minitwit \
  -e POSTGRES_PASSWORD=<dit-password> \
  -v pgdata:/var/lib/postgresql/data \
  postgres:16-alpine
```

Vent til PostgreSQL er klar:
```bash
docker exec minitwit-postgres pg_isready -U minitwit
```

### Steg 3 — Migrer data (appen kører stadig på SQLite)

Merge `feature/postgres-migration` → `dev` → `master` (via PR).
Vent til CD pipeline har bygget nyt image (tjek GitHub Actions).

Kør migration i en midlertidig container med adgang til både SQLite og PostgreSQL:
```bash
docker run --rm \
  --network minitwit-net \
  -v minitwit-db:/tmp \
  -e DATABASE_URL="postgres://minitwit:<dit-password>@minitwit-postgres:5432/minitwit?sslmode=disable" \
  ghcr.io/devtroopers-itu/minitwit:latest ./migrate
```

Verificer at data er kopieret:
```bash
docker exec minitwit-postgres psql -U minitwit -c "SELECT 'users:', COUNT(*) FROM \"user\" UNION ALL SELECT 'messages:', COUNT(*) FROM message UNION ALL SELECT 'followers:', COUNT(*) FROM follower;"
```

### Steg 4 — Switch appen til PostgreSQL

Stop den gamle PostgreSQL container og lad docker-compose overtage:
```bash
docker stop minitwit-postgres && docker rm minitwit-postgres

cd /root/itu-minitwit
git pull origin master
docker compose pull
docker compose up -d
```

Appen starter nu med PostgreSQL. ~2-3 sek nedetid under container swap.

### Steg 5 — Verificer

```bash
curl http://46.224.144.214:8080/latest
curl -s -H "Authorization: Basic c2ltdWxhdG9yOnN1cGVyX3NhZmUh" "http://46.224.144.214:8080/msgs?no=3"
```

## Rollback

Hvis noget går galt: fjern `DATABASE_URL` fra docker-compose.yml og kør `docker compose up -d`.
Appen falder tilbage til SQLite med al gammel data intakt.
