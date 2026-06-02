# Session 02 — Packaging, Docker, rewrite to a new language

Material: `../../course-material/sessions/session_02/` (Slides.md, README_TASKS.md)

## Two tasks this week
1. Rewrite MiniTwit in a new language — **one-to-one copy, NOT a big rewrite**. Same features, small steps, endpoint by endpoint. We did Python/Flask → **Go + Gorilla Mux**.
2. Put it in Docker.

## Why Docker (the main exam point)
- Copying files (`scp`/zip) does NOT bring the **dependencies**. App breaks on another machine.
- Container ships the app + everything it needs together → runs the same everywhere (dev + prod). Reproducible environment for the whole team.

## Container vs VM (likely exam question)
- Container = **shares the host kernel**. Has its own files, processes, network. Light, starts fast.
  - Linux features: **namespaces** = isolation, **cgroups** = resource limits.
- VM = runs a **whole second operating system** inside. Heavy, slow to start.

## Missing dependencies + the sharing ladder
- `scp`/zip do NOT carry the app's libraries → breaks on the next machine.
- Ladder (each step carries more deps): a) scp → b) git tag + zip → c) language packages (PyPI) → d) OS packages (.deb/apt) → e) **containers + registry** (Docker). Docker = top of the ladder.

## Image vs container vs registry
- Image = the built package (recipe result). Container = a running copy of an image. Registry = store/share images (Docker Hub, **GHCR** — we use `ghcr.io`).
- Dockerfile = the recipe. Built in **layers** (cached steps → faster rebuilds).
- `docker compose` = run several containers together (app + db) with one file; `depends_on`, `healthcheck`.

## Replace → Refactor → Rewrite (Helge's order)
- Replace (off-the-shelf) — least work, not allowed for us (we own MiniTwit to learn).
- Refactor — small steps, feature by feature, tests running. **What we did.**
- Rewrite — full redo, risky/slow, last resort.

## Refactor discipline (links to S01)
- Task said: refactor the APP first, get the existing tests passing, THEN port the test suite. Keep templates/static.
- "Do not do the big rewrite" = avoid redesigning. Language change done with refactoring discipline (small, test-guided steps).

## What WE did — multi-stage Dockerfile
- Stage 1 (builder, `golang:1.25-alpine`): build tools, `go mod download`, `go build -o myserver .`.
- Stage 2 (final, `alpine:3.23`): copies only the binary + templates + static. No compiler. Runs as non-root `appuser`. `CMD ["./myserver"]`.
- Build-time vs run-time: `RUN` = at build (baked into image). `CMD` = when container starts. `USER` = setting for the running process.
- **Layer caching:** copy `go.mod`/`go.sum` + `go mod download` BEFORE copying code → dependency layer stays cached when only app code changes (no re-download). Order is deliberate.

## FINDINGS (verified from code) — strong exam material
- DB: `db.go` uses **Postgres only** (`postgres.Open(DATABASE_URL)`). SQLite appears ONLY in `main_test.go` (tests).
- **Stale doc:** `CLAUDE.md` says "SQLite for local dev / `go run .` uses SQLite by default" — NOT true; the code has no SQLite path, `go run .` needs `DATABASE_URL` (Postgres).
- **Dockerfile waste:** `CGO_ENABLED=1` + `build-base` + `sqlite-dev` exist for SQLite, but the prod binary doesn't use SQLite (Postgres driver is pure Go). Likely removable → build with `CGO_ENABLED=0` for a smaller static binary. (Test-build to confirm.)

## Where does the database come from? (3 situations)
1. **Unit tests** (`go test ./...`) → in-process SQLite temp file. Only use of SQLite. No server.
2. **Running the app** (`go run .` / `docker compose up`) → needs **Postgres** at `DATABASE_URL`. Main `docker-compose.yml` has NO postgres service; it passes `DATABASE_URL=${DATABASE_URL}` from `.env`. Empty URL → app stops.
3. **E2E tests** (`docker-compose.test.yml`) → DOES start `postgres:16` (`db` service), `DATABASE_URL=postgres://postgres:testpassword@db:5432/minitwit`. Only bundled app+DB setup.
- `getSecretOrEnv` (main.go:115): reads `/run/secrets/<name>` first, else env var.
- Local reality: `.env` → `postgres://...@localhost:5433/minitwit`; a local `postgres:17-alpine` runs on :5433. Local dev = own Postgres + `go run .`. Never hits prod. Tests (`make test`) use SQLite, need no DB.
- `.env` DOES configure the DB for `go run .` (verified empirically). Reason: `SECRET_KEY = getSecretKey()` is a package-level var → Go runs it before `main()` → its `godotenv.Load()` loads `.env` into the env before `initDB()` reads `DATABASE_URL`. So `set -x DATABASE_URL` is NOT needed if you run from the repo dir with `.env` present. (Earlier "footgun" note was wrong — corrected.)

## SECRET_KEY bug (security — strong exam material)
- `getSecretKey()` only reads `SECRET_KEY` if `godotenv.Load()` succeeds (i.e. a `.env` exists).
- Production containers have NO `.env` (Dockerfile doesn't copy it) → Load fails → returns hardcoded `"dev-fallback-key-change-in-production"`.
- `docker-stack.yml` provisions a real `secret_key` Docker secret (mounted at `/run/secrets/secret_key`) but the code NEVER reads it.
- Result: prod signs session cookies with a key hardcoded in the public repo → cookie forgery risk. Works "right" only locally (because `.env` exists).
- **CONFIRMED LIVE IN PROD (1 Jun 2026, ssh root@64.226.116.162):** container `d5d356f9aa0b` → `NO_DOTENV`, `SECRET_KEY` env unset, `/run/secrets/secret_key` mounted but unread. Deployed image has no `.env`. So prod is using the hardcoded fallback session key right now. Anyone reading the repo can forge login cookies.
- Fix: read `SECRET_KEY` directly via `getSecretOrEnv` (like `DATABASE_URL`), not gated behind `.env`. Ties to config management (S05) + security (S11).

## More findings
- **Stale doc:** `CLAUDE.md` "docker compose up = app + postgres + monitoring" — wrong, no postgres service in docker-compose.yml.
- **Dead leftover:** `minitwit-db` volume declared in docker-compose.yml but no service mounts it (old SQLite-volume relic).
- **Dev/prod parity gap (exam point):** unit tests = SQLite, prod = Postgres. SQLite-green ≠ Postgres-correct (SQL dialect differences). E2E tests use Postgres → softens it; Go unit tests don't.

## Defensible Dockerfile decisions (if examiner asks "why like that?")
- **Two stages:** small + safe final image (builder has Go compiler + C tools, ~300MB; final = tiny alpine + binary, ~20-40MB). Separate build-needs from run-needs. (Layer caching = separate thing, from instruction order, works in one stage too.)
- **Non-root user** (`addgroup`/`adduser`/`chown`/`USER appuser`): default container process = root; running as a powerless user limits damage if the app is breached. Least-privilege, standard best practice. A thing we did RIGHT.
- **`# hadolint ignore=DL3018`**: hadolint = our Dockerfile linter (1 of 3 static-analysis tools, S07). DL3018 = "pin apk package versions" for reproducibility. We ignore it because Alpine drops old versions fast → pinning breaks builds. Trade-off: build stability over strict reproducibility.
- **templates/static copied separately:** they're content (HTML/CSS) read from disk at runtime, not Go code, so not compiled into the binary. Could `//go:embed` them (see findings.md).

## 2-stage answer (resolved)
- Leo's first answer described layer caching, not the reason for 2 stages. Corrected: 2 stages = small/safe image; caching = instruction order. Both separate wins in the same Dockerfile.
