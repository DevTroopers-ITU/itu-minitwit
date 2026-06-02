# Coverage Tracker

Status: `todo` → `taught` (walked through) → `drilled` (mock-exam questions passed).
Topics below are the *assigned* tasks (from `docs/progress.md` + each `README_TASKS.md`), not slide titles
(slides reuse one generic title). Verify against course material when we get to each.

Lecturer: Helge (S1-7, 13-14), Mircea (S8-12). Lectures Fridays 12:00-16:00.

| # | Date | Lecture topic | Project task | Status |
|---|------|---------------|--------------|--------|
| 00 | — | Preparation / setup (Linux, tooling) | — | todo |
| 01 | 30 Jan | Project start, groups, SSH, SCP, Bash | Refactor MiniTwit to modern system (Python 2→3) | taught |
| 02 | 6 Feb | Packaging apps, Containerization w/ Docker | Refactor in another language/stack (we chose Go) | taught |
| 03 | 13 Feb | Provisioning local + remote VMs (Vagrant, DO) | Deploy MiniTwit to a remote server | taught |
| 04 | 20 Feb | Guest "What is DevOps?" + CI/CD | Setup CI & CD pipeline | taught |
| 05 | 27 Feb | What is DevOps + configuration management | DB abstraction layer, enter maintenance (sim starts) | todo |
| 06 | 6 Mar | Monitoring | Prometheus + Grafana + peer-review | todo |
| 07 | 13 Mar | Software Quality, Maintainability & Tech Debt | Tests + static analysis in CI (Sonar) | todo |
| 08 | 20 Mar | Logging & Log Analysis (EFK) | Add logging + UI-test another group | todo |
| 09 | 27 Mar | Availability | Isolate components into services (Swarm) | todo |
| —  | 3 Apr | Easter break | (keep ops running) | n/a |
| 10 | 10 Apr | Workshop (no slides) | Continue isolation, fix reported problems | todo |
| 11 | 17 Apr | Security | Pentesting + security hardening | todo |
| 12 | 24 Apr | Infrastructure as Code + Documentation | Encode infra (Terraform) | todo |
| 13 | 1 May | Guest (Kubernetes) + Exam prep (REPORT.md) | Write report | todo |
| 14 | 8 May | Exam prep, evaluation (sim stops) | Write report | todo |

## My presentation parts (see presentation.md)
- Reflection: Evolution & Refactoring
- Reflection: DevOps Style (Three Ways)

## Open gaps to close before Thursday
- [S01] Don't conflate original Flask `minitwit_tests.py` (white-box, not ported) with `minitwit_sim_api_test.py` (black-box, the simulator contract). See session-01.md.
- [S02] `CLAUDE.md` says SQLite is the local-dev default, but `db.go` is Postgres-only (SQLite only in tests). Doc is stale.
- [S02] Dockerfile enables CGO + installs sqlite-dev/build-base for SQLite the prod binary never uses → likely removable (CGO_ENABLED=0). Cleanup + exam talking point.
- [S02] No local Postgres in main docker-compose.yml; only docker-compose.test.yml bundles one. CLAUDE.md "compose = app+postgres" is stale. `minitwit-db` volume is dead/unused.
- [S02] Dev/prod parity gap: unit tests on SQLite, prod on Postgres. Have an honest answer ready for the exam.
- [S02/S11] SECURITY BUG (CONFIRMED LIVE IN PROD, 1 Jun 2026): `getSecretKey()` only reads SECRET_KEY when a `.env` exists; prod container has none → uses hardcoded fallback key. The `secret_key` Docker secret IS mounted at /run/secrets/secret_key but never read. Prod login cookies signed with a key public in the repo → cookie forgery. Fix: read SECRET_KEY directly via getSecretOrEnv. (Strong exam topic; should actually fix.)
- [S07] RESOLVED: `docs/progress.md` claim "no Selenium/Playwright test" is STALE. Playwright test `minitwit_ui_browser_test.py` exists (`3d5ce5b`, PR #168) AND runs in CI (ci.yml). Fix the report/progress if it repeats this.
