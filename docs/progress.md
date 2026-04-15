# Project Progress — DevTroopers

Cross-referenced from course session tasks and git history. Updated April 2026.

## Session 1 (30 Jan) — Project Start, SSH, Bash

**Assigned:** Add version control, understand the Python/Flask codebase, migrate to Python 3 (2to3), share on GitHub, create release by Thu 5 Feb.

**Done:**
- `ae39e23` (Jan 30): Initial commit with original ITU-MiniTwit
- Python 3 migration (2to3, werkzeug import fixes)
- Created public GitHub repo, team formed
- Release created

## Session 2 (6 Feb) — Refactor to New Language + Docker

**Assigned:** Rewrite ITU-MiniTwit in a new language/framework (one-to-one feature parity, not a big rewrite). Containerize with Docker. Release by Thu 12 Feb.

**Done:**
- Chose Go with Gorilla Mux (course-recommended framework)
- `a2ab6fc` (Feb 12): Go implementation matching original features
- `d0945ab` (Feb 12): Docker setup — tagged **v1.0.0**
- Multi-stage Dockerfile (Alpine + Go build)
- Reused HTML templates from original (layout, login, register, timeline)

## Session 3 (13 Feb) — Simulator API + Deploy to Cloud

**Assigned:** Implement simulator API (endpoints: /latest, /msgs, /fllws, /register). Create VM in cloud via Vagrantfile/script (not PaaS). Deploy, make public. Send PR to repositories.py. Document deployment in README.

**Done:**
- `65e4ae9` (Feb 14): Transferred repo to DevTroopers-ITU organization
- `9bbf1f1` (Feb 16): Simulator API — registration, messages, follows
- `fad09c0` (Feb 16): Swagger API documentation
- `5601dd0` (Feb 18): Vagrantfile for Hetzner Cloud (Ubuntu 22.04)
- `a689204` (Feb 18): Port 8080, Docker volume for DB persistence
- Deployed to Hetzner VPS, public IP

## Session 4 (20 Feb) — CI/CD

**Assigned:** Complete simulator API (simulator starts next week). Create CI/CD pipeline with build, test, deploy. Automatic releases from now on.

**Done:**
- `d1594ba` (Feb 24): GitHub Actions CI — lint, build, test
- (Feb 27): Refactored CI with docker-compose.test.yml, isolated PostgreSQL
- `a966c92` (Mar 3): CI triggers on PRs to dev
- `2590efc` (Mar 6): CD pipeline — build + push to ghcr.io, SSH deploy to Hetzner

## Session 5 (27 Feb) — Simulator Starts, DB Abstraction, Config Management

**Assigned:** Simulator goes live. Reflect on DevOps "Three Ways". Introduce DB abstraction layer (ORM, no raw SQL). Idempotent config management. Fix issues as they occur — we're in maintenance mode now.

**Done:**
- Simulator started hitting our API
- `c7491b0` (Mar 6): Added GORM dependency
- `8a3f7ba` (Mar 6): Refactored all database access to GORM ORM — no more raw SQL
- `ae9853b` (Mar 9): Split store/session globals for cleaner architecture
- `b05cb22` (Mar 6): Tagged **v1.1.0**
- `c972607` (Mar 10): Added PostgreSQL support via GORM driver abstraction
- `53bab02` (Mar 10): PostgreSQL migration plan documented
- `f0a0b01` (Mar 10): Data migration strategy (migrate data before switching app)

*Note: DB abstraction and Postgres migration done in the week after session 5, overlapping with session 6 week.*

## Session 6 (6 Mar) — Monitoring + DB Migration

**Assigned:** 1) Add application and infrastructure monitoring (Prometheus + Grafana). Turn Grafana dashboards into code. Share monitoring URL via PR to misc_urls.py. 2) Migrate database away from SQLite (self-hosted or managed). Minimize downtime. 3) Continue software maintenance — fix issues within 24h, release at least weekly.

**Done:**
- `2e51fe0` (Mar 18): Prometheus metrics middleware + /metrics endpoint
- Custom counters: `minitwit_http_responses_total` (method, route, status)
- Custom histograms: `minitwit_http_duration_seconds` (method, route)
- `26560a1` (Mar 18): Prometheus + Grafana added to docker-compose
- `ed5d97b` (Mar 26): Normalized route labels in Prometheus middleware
- `88c7179`, `e90488f` (Mar 26): Memory limits, scrape config tuning
- Grafana dashboard as code (dashboard.json) with DNS-based service discovery
- PostgreSQL migration was done during session 5 work (Mar 6–10) — covered there

*Note: Monitoring implemented ~12 days after session, around Mar 18.*

**Gaps:**
- [ ] **misc_urls.py:** Group Q (DevTroopers) is completely absent from misc_urls.py — neither monitoring nor logging URLs were submitted. Need to send PR.
- [ ] Confirm Grafana is accessible to instructors (read-only user or shared credentials from Helge via Teams)

## Session 7 (13 Mar) — Code Quality & Technical Debt

**Assigned:** 0) Add integration, UI, and E2E tests to CI as quality gates. 1) Add at least 3 static analysis tools to CI (linter + formatter for main language, Dockerfile linter). 2) Add SonarQube and Codacy for maintainability/tech debt tracking. React on prominent issues. 3) Finish any incomplete features. Continue maintenance + weekly releases.

**Done:**
- `bdb6c16` (Mar 13): Static analysis — gofmt, golangci-lint, hadolint (3 tools ✓)
- `1b99c35` (Mar 13): SonarCloud integration (on feature/code-quality branch)
- `57a4f1c` (Mar 13): Removed secret key from codebase (security fix)
- `4be0cca` (Mar 18): Merged feature/static-analysis PR #77
- `.golangci.yml` configured: errcheck, govet, staticcheck, unused, gosimple
- Linter checks enforced in CI pipeline as quality gates
- Go integration tests (main_test.go) run in CI — tests HTML UI endpoints
- Python E2E API tests (minitwit_sim_api_test.py) run in CI via docker-compose.test.yml

**Gaps:**
- [ ] **SonarCloud:** Was configured on feature/code-quality branch with sonar-project.properties + CI step, but was never merged to master. Not running in production CI.
- [ ] **Codacy:** Never set up. No config, no CI step, no references in repo.
- [ ] **Browser-based UI test:** No Selenium/Playwright test like `test_itu_minitwit_ui.py` from lecture. Go tests cover UI endpoints via HTTP but not via a real browser.

## Session 8 (20 Mar) — Logging

**Assigned:** 1) Add logging (EFK or other stack). Share logging dashboard URL if possible. Improve logging over time. 2) Peer review: check another group's UI for functionality and report issues via GitHub issues.

**Done:**
- `f2bebb9` (Mar 30): Grafana Loki + Promtail logging stack
- Promtail collecting logs via Docker socket (global mode across Swarm nodes)
- Loki config for log storage and querying
- logging-dashboard.json in Grafana
- `83fc3d9` (Mar 30): Updated architecture diagram with monitoring + logging

*Note: Logging implemented ~10 days after session, during pre-Easter period.*

**Gaps:**
- [ ] **misc_urls.py:** Logging URL not submitted (same as monitoring — Group Q missing entirely)
- [ ] **Peer review:** DevTroopers is Group Q (MSc). Peer review assignment: Group o checks Group q, meaning someone should check *us*. We need to check — which group were we assigned to review? (The MSc assignment list: g→m, o→q, l→n, i→c, k→j — Group Q is *reviewed by* Group O, but Group Q is not listed as a reviewer)

## Session 9 (27 Mar) — Availability & Docker Swarm

**Assigned:** 1) Either hot/standby HA setup OR Docker Swarm cluster. 1b) Optional: measure downtime during migration. 2) Implement automatic update strategy (rolling updates or blue-green). 3) Add availability view to Grafana: service health, HTTP success rate, P95 latency. Bonus: alerting when availability drops. 4) Continue maintenance + weekly releases.

**Done:**
- `3c649da` (Apr 10): Docker Swarm — 3-node DigitalOcean cluster (option 2 ✓)
  - Manager: 64.226.116.162
  - Worker 1: 206.189.59.60
  - Worker 2: 134.122.90.176
- docker-stack.yml with 3 webserver replicas
- Docker secrets for DATABASE_URL, SECRET_KEY, DISCORD_WEBHOOK_URL
- Promtail in global mode on all nodes
- `59930d8` (Apr 10): Fixed missing ports configuration
- `e619a92` (Mar 30): Grafana alerting with Discord webhook (bonus ✓)
- **Rolling updates ✓** — explicitly configured in docker-stack.yml: `parallelism: 1`, `delay: 10s`, `order: start-first` with rollback config. Documented in docs/docker-swarm.md.
- **Healthcheck ✓** — webserver has `wget --spider` healthcheck in docker-stack.yml

*Note: Docker Swarm migration done 2 weeks after session (Apr 10), bundled with TLS work.*

**Gaps:**
- [ ] **HTTP success rate panel:** Dashboard has "Error rate (4xx and 5xx)" but NOT the required "percentage of non-5xx responses" success rate panel. Metric exists to build it: `minitwit_http_responses_total`.
- [ ] **Service health panel:** Only monitors single hardcoded instance (`webserver:8080`). Should monitor all Swarm services. Prometheus only scrapes the webserver job, not other services.
- [ ] **P95 latency panel:** Exists ✓ but dashboard time range is hardcoded to last 30 minutes. Course asks for "last week and last month" views.
- [ ] **Health check endpoint:** No `/health` endpoint in Go code. The `up` metric relies on Prometheus scraping `/metrics`.
- [ ] Downtime measurement during migration (optional but good for report)

## Session 10 (10 Apr) — TLS

**Task descriptions not yet released in lecture notes. Topic: TLS/SSL certificates.**

**Done:**
- `3eeb37c` (Apr 10): Traefik reverse proxy with Let's Encrypt TLS
- HTTPS for devtroopersminitwit.codes + grafana.devtroopersminitwit.codes
- Traefik on Swarm manager (constraint: node.role == manager)
- Alert rules in prometheus.rules.yml
- Contact points + policies provisioned in Grafana

---

## Current Position

**Date:** 14 April 2026
**Simulator:** Running since session 5, ends session 13
**Releases:** v1.0.0 (Feb 12), v1.1.0 (Mar 6)

### Recent merges
- Docker Swarm, Traefik, alerting, logging — merged to master via PRs #112, #114, #120
- PR #122 (dev → master) approved, pending merge by Peter
- CD pipeline deploys to both DigitalOcean Swarm (primary) and Hetzner (legacy)

### Outstanding debt (across all sessions)

**High priority — course deliverables:**
- [ ] **Weekly releases:** Only 2 tags since session 4 (v1.0.0, v1.1.0). Course requires at least weekly releases. 39 days since last release.
- [ ] **misc_urls.py PR:** Group Q is completely absent — neither monitoring nor logging URLs submitted to course repo.
- [ ] **SonarCloud:** Configured on feature/code-quality branch but never merged to master CI pipeline.
- [ ] **Codacy:** Never set up at all.

**Medium priority — incomplete implementations:**
- [ ] **HTTP success rate panel** in Grafana (have error rate, need success rate)
- [ ] **Service health panel** only monitors one hardcoded instance, not all Swarm services
- [ ] **P95 latency dashboard** time range hardcoded to 30min, should show week/month
- [ ] **/health endpoint** — no dedicated health check in Go code
- [ ] **Browser-based UI test** (Selenium/Playwright) — not in CI

**Lower priority / verify:**
- [ ] Peer review — unclear if DevTroopers was assigned to review another group (Group Q not in reviewer list)
- [ ] Grafana accessible to instructors with shared credentials

---

## Upcoming

### Session 11 (17 Apr) — Security & Pentesting

- [ ] Security audit of application
- [ ] Penetration testing
- [ ] Harden infrastructure and application
- [ ] Fix vulnerabilities found

### Session 12 (24 Apr) — Infrastructure as Code

- [ ] Terraform or similar IaC tool
- [ ] Codify DigitalOcean Swarm infrastructure
- [ ] Reproducible infrastructure provisioning

### Session 13 (1 May) — Kubernetes & Documentation

- [ ] Architecture documentation (for exam report)
- [ ] Possibly Kubernetes exploration
- [ ] Simulator ends — ensure full API compliance
- [ ] Guest lecture

### Session 14 (8 May) — Exam Prep

- [ ] Final report writing
- [ ] System documentation review
- [ ] Presentation preparation

---

## Exam

- **Submission:** 18 May 2026 14:00 via WISEflow
- **Oral exam:** 1–4 June 2026
- **Format:** Pass/fail — must be able to explain all project work
- **Key:** Every team member must understand every part of the system

## Team

- Leo Sakharoff (leosakharoff)
- Håkon Refsvik
- Peter Juul Møller (pemoe)
- Apoorva (DenSygeMike)
