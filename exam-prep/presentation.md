# Presentation — Wednesday run-through

Slides (group): https://docs.google.com/presentation/d/1kiGHwznjMtR_pc5oT92llKzMlrLH-Lgvmw3jGgxEpJU/edit

Can't read a Google Slides `/edit` link directly (needs auth). To work the actual slide text,
export to PDF into this folder (`exam-prep/slides.pdf`) or paste the text here.

Format: ~3 min speaking each, then longer Q&A at the oral exam.
Detail + git/live proof for every claim below lives in [findings.md](findings.md) and the session notes.

---

## PART 1 — Reflection: Evolution & Refactoring (~3 min)

**One-line thesis:** "We took over a 2012 legacy app and evolved it on every layer — language, tests, database, infrastructure — using refactoring discipline, not big rewrites."

**The evolution, layer by layer (concrete):**

- **Language:** Python 2 → Python 3 (minimal-change, test-guided) → **Go + Gorilla Mux** (one-to-one feature parity, not a redesign). The Py3 step wasn't wasted — it was how we *understood* the system and got a green test baseline to port against.
- **Tests:** Flask `test_client` (white-box) → rewritten as Go integration tests; the course's **black-box** sim contract test survived the language change; later added a **Playwright** browser test. Lesson: white-box tests die at a language port, black-box ones survive. 3 test layers in CI now.
- **Database:** SQLite → **PostgreSQL**, made swappable by the GORM abstraction layer.
- **Infrastructure:** single Vagrant-provisioned **Hetzner** VM (docker-compose + `.env`) → **3-node Docker Swarm on DigitalOcean** (Traefik + TLS + Docker secrets). Hetzner decommissioned.

**Honest reflection (this is the gold):** the **SECRET_KEY regression**. A security fix (remove hardcoded key) → a quick test patch (re-add a hardcoded *fallback*) → a swarm migration that left `SECRET_KEY` reading from `.env` while everything else moved to secrets. Refactoring and migrations carry risk; small well-meant edits compound into a live security bug. We caught it in prep and know the fix.

**If they say "show me":** `git log -L :getSecretKey:main.go` walks the whole story in one command.

---

## PART 2 — Reflection: DevOps Style (The Three Ways) (~3 min)

**Frame as an honest scoreboard.** Three Ways (session 5 / DevOps Handbook): Flow, Feedback, Continual Learning.

**First Way — Flow (dev → prod, fast):** = short **lead time** + small **batches** + few **handoffs**.

- Did well: CI/CD auto-deploys on merge to master (build → push → `docker stack deploy`) — the **manual-deploy handoff is gone** (no human SSHing in). Small batches: a fix (e.g. a DB-URL env change) ships on its own, no waiting for a big release. Multi-stage Docker = small images.
- Gap: **IaC isn't truly reproducible** — prod droplets were provisioned by hand; the Terraform is descriptive and unvalidated (doesn't even match prod's topology). Flow breaks if we lose the box.

**Second Way — Feedback (fast signal back):**

- Did well: Prometheus + Grafana metrics, Loki + Promtail logs, Discord alerts, healthchecks, a real test pyramid (fast Go → black-box contract → Playwright E2E).
- **Concrete story — caught a silent bug in pre-prod parity testing (strong):** the `latest` counter was an in-memory int that broke only across multiple replicas, so the simulator read stale ~2/3 of the time — *silent* (200 OK, wrong data, invisible to our up/5xx/latency alerts). We caught it by running `scripts/test-do-swarm.sh` against the **DO swarm while Hetzner was still production** (20 Apr), fixed it next day (PR #138, `b874e8e`), and cut over ~2 weeks later. = production-like pre-prod testing + Push Quality to the Source. (Also a GenAI example — Claude diagnosed it; a separate proposed join-fix was a no-op.) Contrast: a 41–49s JOIN query (`6e94660`) was *loud* — it would trip our latency alert.
- Gap: **dev/prod parity** — tests run on SQLite, prod on Postgres, so a green test doesn't fully prove prod behaviour. And the **secret-key bug was a missing security feedback loop** — nothing told us prod used a public key.

**Third Way — Continual Learning & Experimentation:**

- **Concrete story — the Traefik 504 incident (PR #129, `c3abcc2`):** HTTP/2 504s under Swarm ingress mesh, fixed by host-mode port publishing. The learning artifact: a **719-line incident log** (`docs/incidents/session11-ops-debug.md`) with a timestamped play-by-play, a running **"Lessons recorded"** list, a **group-meeting crib sheet**, and a **pre-flight checklist** for the cutover. Plus a whole `docs/incidents/` + `docs/architecture/` tree (e.g. `latest-counter-in-db.md` with a "One-Paragraph Defense"). = Institutionalize Improvement + turn local discoveries into global ones. **This is our strongest Way.**
- Also: weekly releases; the GHCR-token incident documented inline in `cd.yml`; this exam-prep audit = finding + fixing latent issues.
- Gap: **docs drift** — CLAUDE.md and progress.md had claims the code no longer matched. Learning wasn't always captured.
- **GenAI nuance (honest, also for the GAI report section):** much of this documentation was AI-assisted. Pitfall: AI over-produces — a 719-line log is *captured* but not *transformed* until a human distills it. The valuable part is the human-written **"Lessons recorded"** list, not the raw trace. So: AI to debug/understand fast → human to compress into readable, reusable lessons. "Are LLMs helping? Yes, but you must fight the slop and humanize the output." Volume ≠ value.

**Punchline:** "We score well on Flow and Feedback for the app itself; our weakest Way is reproducibility/learning at the infra+docs level — and we can name exactly where and why."

---

## To do before Wednesday

- [ ] Get slide content in (PDF export or paste) and tighten each ~3-min script
- [ ] Rehearse the git proof commands (see [git-cheatsheet.md](git-cheatsheet.md))
- [ ] Pick the 1-2 honest weak spots to lead with (secret-key story is the strongest)
