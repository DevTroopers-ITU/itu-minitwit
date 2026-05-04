# Docker Scout Vulnerability Scan — Minitwit

## How to Run the Scan

Requires Docker Desktop with Scout enabled and a Docker Hub login:

```bash
docker login
docker scout cves ghcr.io/devtroopers-itu/minitwit:latest
```

For a quick summary instead of the full report:

```bash
docker scout quickview ghcr.io/devtroopers-itu/minitwit:latest
```

---

## Scan Results (April 23rd 2026)

**Image:** `ghcr.io/devtroopers-itu/minitwit:latest`  
**Platform:** `linux/arm64`  
**Packages scanned:** 98  
**Total vulnerabilities:** 31 across 7 packages

| Severity    | Count |
|-------------|-------|
| CRITICAL    | 2     |
| HIGH        | 12    |
| MEDIUM      | 12    |
| LOW         | 3     |
| UNSPECIFIED | 2     |

---

## Vulnerabilities by Package

### `github.com/jackc/pgx/v5` @ 5.6.0 — 2 CRITICAL, 1 LOW
Our Postgres driver. The two CRITICAL findings have a CVSS score of 9.8/10.

| CVE | Severity | Fix |
|-----|----------|-----|
| CVE-2026-33816 | CRITICAL (9.8) | Upgrade to 5.9.0 |
| CVE-2026-33815 | CRITICAL | Upgrade to 5.9.0 |
| GHSA-j88v-2chj-qfwx (SQL Injection) | LOW | Upgrade to 5.9.2 |

**Fix:**
```bash
go get github.com/jackc/pgx/v5@v5.9.2
go mod tidy
```

---

### `openssl` @ 3.5.5-r0 — 5 HIGH, 2 UNSPECIFIED
Alpine OS package. All fixed in `3.5.6-r0`.

**Fix:** Run `apk upgrade` in the Dockerfile (see Dockerfile fix below).

---

### `stdlib` @ 1.24.13 — 4 HIGH, 4 MEDIUM, 1 LOW
Go standard library. All fixed in Go 1.25.9.

**Fix:** Bump the base image to `golang:1.25-alpine` when available, or add `apk upgrade` to pick up patches.

---

### `binutils` @ 2.45.1-r0 — 2 HIGH, 5 MEDIUM
Build tooling pulled in by `build-base`. **No fix available yet** — accepted risk, document and monitor.

---

### `musl` @ 1.2.5-r21 — 1 HIGH, 1 MEDIUM
Alpine's C library. Fixed in `1.2.5-r23`.

**Fix:** Run `apk upgrade` in the Dockerfile (see below).

---

### `zlib` @ 1.3.1-r2 — 1 MEDIUM, 1 LOW
Compression library. Fixed in `1.3.2-r0`.

**Fix:** Run `apk upgrade` in the Dockerfile (see below).

---

### `busybox` @ 1.37.0-r30 — 1 MEDIUM
Shell utilities. **No fix available yet** — accepted risk, document and monitor.

---

## Remediation

### 1. Bump the pgx dependency (go.mod)

```bash
go get github.com/jackc/pgx/v5@v5.9.2
go mod tidy
```

This fixes the 2 CRITICAL and 1 LOW findings in our Postgres driver.

### 2. Add `apk upgrade` to the Dockerfile

Adding this line forces Alpine to upgrade all OS packages to their latest patched versions at build time, fixing the openssl, musl, and zlib findings:

```dockerfile
FROM golang:1.24-alpine

WORKDIR /app

# hadolint ignore=DL3018
RUN apk add --no-cache build-base sqlite-dev && apk upgrade --no-cache

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o myserver .

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8080
CMD ["./myserver"]
```

### 3. Known/Accepted Risks

The following vulnerabilities have **no fix available** and are accepted risks:

| Package | CVEs | Reason |
|---------|------|--------|
| `binutils` | CVE-2025-69644 to 69652 | Required by `build-base` for CGO compilation, no patch released |
| `busybox` | CVE-2025-60876 | No patch released |

These should be monitored and patched as soon as fixes become available.

---

## After Applying Fixes

Commit and push — CI will rebuild the image. Then re-run the scan to verify:

```bash
docker scout cves ghcr.io/devtroopers-itu/minitwit:latest
```

Expected result: 0 CRITICAL, significantly reduced HIGH count.
