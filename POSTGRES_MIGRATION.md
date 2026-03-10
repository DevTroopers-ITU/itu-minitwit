# PostgreSQL Migration Plan

## Pre-deploy (GitHub)

1. Gå til repo Settings → Secrets → New repository secret
2. Tilføj `DB_PASSWORD` med et sikkert password (fx genereret med `openssl rand -base64 24`)

## Deploy dag

### Steg 1 — Forbered serveren (SSH)

```bash
ssh root@46.224.144.214
```

Opret `.env` fil:
```bash
echo "DB_PASSWORD=<samme-password-som-github-secret>" > /root/itu-minitwit/.env
```

### Steg 2 — Merge til master

Merge `feature/postgres-migration` → `dev` → `master` (via PR).

CD pipeline bygger nyt image og deployer. Docker compose starter:
- PostgreSQL container (tom database)
- Webserver (forbinder til PostgreSQL, AutoMigrate opretter tomme tabeller)

**Appen er nu live med tom PostgreSQL.** Simulatoren registrerer nye brugere og poster beskeder — de gemmes i PostgreSQL. Gamle data mangler stadig.

### Steg 3 — Migrer data (SSH, mens appen kører)

```bash
ssh root@46.224.144.214
```

Find webserver container ID:
```bash
docker ps --filter "ancestor=ghcr.io/devtroopers-itu/minitwit:latest" -q
```

Kør migration:
```bash
docker exec -e DATABASE_URL="postgres://minitwit:<DB_PASSWORD>@db:5432/minitwit?sslmode=disable" <CONTAINER_ID> ./migrate
```

Dette kopierer alle brugere, beskeder og follows fra SQLite til PostgreSQL. Tager ~30 sek.

### Steg 4 — Verificer

```bash
# Tjek at API'en virker
curl http://46.224.144.214:8080/latest
curl -s -H "Authorization: Basic c2ltdWxhdG9yOnN1cGVyX3NhZmUh" "http://46.224.144.214:8080/msgs?no=3"

# Tjek database indhold
docker exec <POSTGRES_CONTAINER_ID> psql -U minitwit -c "SELECT COUNT(*) FROM \"user\";"
docker exec <POSTGRES_CONTAINER_ID> psql -U minitwit -c "SELECT COUNT(*) FROM message;"
docker exec <POSTGRES_CONTAINER_ID> psql -U minitwit -c "SELECT COUNT(*) FROM follower;"
```

## Hvad sker der med simulatoren under migration?

- Steg 2 deploy: ~3 sek container swap → maks 3 tabte requests
- Steg 3 migration: Appen kører hele tiden, ingen nedetid
- Nye brugere/beskeder fra simulatoren mellem steg 2 og 3 gemmes i PostgreSQL
- Gamle data kopieres i steg 3 (duplikater undgås via primary keys)

## Rollback

Hvis noget går galt, fjern `DATABASE_URL` fra docker-compose og redeploy. Appen falder tilbage til SQLite med al gammel data intakt.
