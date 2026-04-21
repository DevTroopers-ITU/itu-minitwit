# DevTroopers — `docs/` Index

Start here. Each file is linked once below with a one-line summary of what's in it.

## Daily operations

- [`operations/ops-cli-reference.md`](operations/ops-cli-reference.md) — Cheat sheet for SSH, Swarm, logs, deploy/rollback, troubleshooting, DB access, CI/CD. **Start here when something looks broken in prod.**
- [`progress.md`](progress.md) — Session-to-implementation tracker. Which course session each change maps to.

## Deployment & infrastructure

- [`operations/docker-swarm.md`](operations/docker-swarm.md) — Why Swarm, node layout (Manager + 2 workers on DO), service distribution, initial setup steps.
- [`operations/firewall-changes.md`](operations/firewall-changes.md) — DO firewall rules: which ports are open on manager vs workers, and why.
- [`operations/ufw-setup-recap.md`](operations/ufw-setup-recap.md) — UFW baseline on all three nodes, layered with the DO firewall.

## Architecture & design decisions

- [`architecture/architecture.md`](architecture/architecture.md) — High-level architecture narrative. Companion to the diagram.
- [`architecture/architecture.drawio`](architecture/architecture.drawio) — Source for the diagram (edit in draw.io / VS Code extension).
- [`architecture/architecture.png`](architecture/architecture.png) — Rendered diagram for reports and slides.
- [`architecture/latest-counter-in-db.md`](architecture/latest-counter-in-db.md) — Why the simulator's `/latest` counter moved from an in-process `int` to a single Postgres row (PR #138).

## Monitoring & alerting

- [`monitoring/alerting.md`](monitoring/alerting.md) — Grafana alert rules and Discord webhook integration.

## Incident notes

- [`incidents/session11-ops-debug.md`](incidents/session11-ops-debug.md) — Debug diary from the Traefik v3.6 / host-mode incident (April 2026). Read before touching Traefik config.

## Conventions for new docs

- **Filename:** kebab-case, no date prefix. If it's an incident write-up, prefix with `sessionN-` or the actual date.
- **First line:** `# Title` — matches the filename's topic.
- **Style:** prose + tables + copy-pasteable code blocks. No tutorial-style filler. Match `firewall-changes.md` or `docker-swarm.md`.
- **Cross-link:** if your doc depends on or relates to another, link it. Add the new file to this index in the same PR.
- **Stage-appropriate:** don't pre-document hypothetical future work. Write it when it ships.
