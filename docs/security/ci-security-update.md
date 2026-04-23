# CI Security Update — Semgrep + Docker Scout

## Overview

Two security scanning tools have been added to the CI pipeline to catch vulnerabilities early — before code reaches production. This follows the "shift-left" security principle: find problems as early as possible in the development process, not after deployment.

---

## What Was Added

### 1. Semgrep — Static Code Analysis

**What it does:** Scans our Go source code on every PR without running it, looking for known security vulnerabilities such as SQL injection risks, hardcoded secrets, insecure function calls, and OWASP Top 10 issues.

**When it runs:** After `lint`, in parallel with `test` — so it catches issues before anything gets built or deployed.

**Rules enabled:**
- `p/security-audit` — general security best practices
- `p/secrets` — detects hardcoded passwords, API keys, tokens
- `p/owasp-top-ten` — checks for the most common web vulnerabilities

**If it fails:** The pipeline stops and the PR cannot be merged until the flagged code is fixed.

---

### 2. Docker Scout — Image Vulnerability Scanner

**What it does:** After the Docker image is built in the `test` job, Scout scans every package inside the image and cross-references them against public CVE databases. It reports vulnerabilities in OS packages, Go dependencies, and the base image itself.

**When it runs:** After `test` — so the image is already built and ready to scan before it would be pushed to the registry.

**Severity threshold:** Fails the pipeline on any `CRITICAL` or `HIGH` findings, blocking deployment automatically.

**If it fails:** The pipeline stops and the image is never pushed to GHCR or deployed to the swarm.

---

## Pipeline Flow

```
lint
 ├── semgrep       ← scans source code (no build needed)
 └── test
      └── docker-scout  ← scans built image (runs after test)
```

---

## Setup (Already Done)

Docker Scout requires a Docker Hub login to run. Two secrets have been added to the GitHub repository:

| Secret | Value |
|--------|-------|
| `DOCKERHUB_USERNAME` | `peterjuulmoller` |
| `DOCKERHUB_TOKEN` | Personal access token (Public Repo Read-only) |

These are stored as GitHub Actions secrets — they are never visible in logs or to other contributors, and are injected automatically when the workflow runs.

---

## How to Verify It's Working

Open a PR against `master` and check the **Actions** tab — you should see all four jobs running:

```
✓ lint
✓ semgrep
✓ test
✓ docker-scout
```

If any job fails, click into it to see which step failed and why.

---

## Related Changes

These CI additions are part of a broader security hardening effort that also includes:

- Non-root user added to the app container (`USER appuser` in Dockerfile)
- Base image bumped from `golang:1.24-alpine` to `golang:1.25-alpine`
- `apk upgrade --no-cache` added to Dockerfile to patch Alpine OS packages
- UFW firewall configured on all three swarm nodes
- `pgx/v5` dependency bumped to `v5.9.2` to fix 2 CRITICAL CVEs
