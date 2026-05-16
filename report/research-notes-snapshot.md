# Snapshot findings — Hetzner backup + live DO Swarm

Material for the wiring agent to fold into `report/report.md` reflection sections (lines 114–169).
Sources: the Hetzner snapshot tarball at `/home/leos/Dev/itu-minitwit-hetzner-snapshot/`
(captured 2026-05-15) and live SSH to the DO Swarm nodes (Manager 64.226.116.162,
Worker-1 134.122.90.176, Worker-2 206.189.59.60), queried 2026-05-16.

Use this as a fact bank. Each finding is tied to a specific sentence in `report.md`
and labelled **CONFIRMS**, **CORRECTS**, or **ADDS**. Apply CORRECTS verbatim. CONFIRMS
items are quote-ready if the paragraph needs a hard number. ADDS only land if the word
budget has room.

---

## Evolution and Refactoring (lines 117–130)

### Rebuild 2 — Replication

**ADDS — the two stacks were architecturally different, not just "Hetzner → DO":**
The Hetzner box ran `docker-compose.yml` (Traefik v3.4, single instance, env-var
secrets, GHCR pull only for webserver). The DO Swarm runs `docker-stack.yml`
(Traefik v3.6.13 pinned, **host-mode** port publishing to bypass IPVS/VXLAN, 3
replicas, **Swarm secrets** for `database_url` and `secret_key`, split
`frontend`/`backend` overlay networks, placement constraints). The Hetzner config
never approached the Swarm shape — they coexisted as parallel systems with
different operational properties.

Source: `/home/leos/Dev/itu-minitwit-hetzner-snapshot/server-repo/files/docker-compose.yml`
vs `docker-stack.yml @ HEAD`.

### Rebuild 3 — Hardening

**CORRECTS — UFW landed on DO Swarm nodes only, not Hetzner:**
The Hetzner snapshot shows `ufw status: inactive`. The DO manager has UFW
**active** with explicit allows for the Swarm overlay ports (2377/tcp, 7946/tcp+udp,
4789/udp) from the worker IPs only. The current draft mentions UFW in the
"Rebuild 3 — Hardening" bullet; consider qualifying as "UFW on the Swarm nodes"
to avoid implying it was a project-wide standard.

Sources: `snapshot/system/firewall-and-listeners.txt` and `ssh root@64.226.116.162 ufw status`.

**CORRECTS — Traefik v3.6 bump was DO-only:**
The draft of the Operation section says "bumped Traefik to v3.6 in the process
(PR #131)". The Hetzner snapshot shows `traefik:v3.4` in compose AND in `docker ps`
(image `traefik:v3.4`). The repo's stack file pins `traefik:v3.6.13` with a
comment about "bug 1 from `docs/incidents/session11-ops-debug.md`". The bump
applied to the Swarm stack only; Hetzner ran v3.4 until decommission.

---

## Operation (lines 132–141)

### Parallel-run window: 10 April → 4 May

**CONFIRMS the dates:**

- DNS cutover Apr 10: Hetzner Traefik's first ACME-challenge failure for
  `devtroopersminitwit.codes` is `2026-04-10T13:34:35Z`. From that moment, the
  Hetzner box could no longer answer the HTTP-01 challenge because the domain
  pointed at the Swarm.
- Decommission Apr/May 4: last `docker service update` on the Swarm completed
  `2026-05-04T15:01:55Z` (2m34s rolling update). Matches commit `e4c234c`
  (decommission-hetzner).
- DO Swarm nodes' uptime today: **36 days** as of 2026-05-16 17:29 UTC →
  last reboot Apr 10, consistent with parallel-run start.

Sources: `snapshot/logs/itu-minitwit-traefik-1.log`,
`ssh root@64.226.116.162 docker service inspect minitwit_webserver`, `uptime`.

### The Swarm-overlay outage — date and duration

**CORRECTS — the outage started 16 April, not 17 April:**

The current draft sentence is *"On 17 April a new DO cloud firewall silently blocked
Swarm's overlay ports between our own nodes."* The Manager's docker journal
(`journalctl -u docker.service`) shows:

- **Start: 2026-04-16 19:38:19 UTC** (21:38 CEST) —
  first `bulk sync to node 134.122.90.176:7946: i/o timeout`.
- **End: 2026-04-17 13:18:06 UTC** (15:18 CEST) — last entry in the
  continuous burst.
- **Duration: ~17h 40min** of broken control plane.
- **Volume: 8,381 timeout lines on Apr 17, 2,748 on Apr 16**, sustained at
  ~630/hour (10/min).
- After 13:18 UTC, only **isolated** single-line timeouts on Apr 19, 21, 25, 26,
  30, May 2, 6, 12 (1–2/day) — cloud-network noise, not an outage.

Suggested rewrite of the sentence: *"On the evening of 16 April a new DO cloud
firewall silently blocked Swarm's overlay ports between our own nodes."* — or
*"Across 16–17 April..."* if you want to keep "17 April" prominent because that
was the remediation day.

**CONFIRMS — data plane stayed up while control plane was broken:**
During the same window, Traefik on the manager kept serving the local replica;
external HTTP probes to 80/443 stayed green. The very point the paragraph makes
about uptime-checks-at-the-edge is exactly what the journal shows.

**CONFIRMS — the fix lives on the box:**
Manager UFW now has explicit allows for **2377/tcp, 7946/tcp+udp, 4789/udp**
from `134.122.90.176` and `206.189.59.60` only. The remediation is literally
visible in `ufw status` on the manager.

### Hetzner Traefik never renewed certs again

**ADDS — fresh "control plane independence" data point:**
After DNS cutover (Apr 10), the Hetzner Traefik kept trying to renew certs
and failed: **38 ACME failures across 5 days (Apr 10/14/15/17/20)**, then
Let's Encrypt backoff and silence until decommission. The Hetzner box served
fine via cached certs the whole time. Worth one half-sentence if there's room
("Hetzner kept serving for ~3 weeks on cached certs after we'd already cut DNS").

Source: `snapshot/logs/itu-minitwit-traefik-1.log` —
`grep "ACME challenge" | awk '{print substr($1,1,10)}' | sort | uniq -c`.

### 3-replica architecture is live and current

**CONFIRMS the "three replicas behind Traefik" claim:**

```
NAME                  MODE         REPLICAS   IMAGE
minitwit_traefik      replicated   1/1        traefik:v3.6.13
minitwit_webserver    replicated   3/3        ghcr.io/devtroopers-itu/minitwit:latest
minitwit_prometheus   replicated   1/1        ghcr.io/devtroopers-itu/minitwit-prometheus:latest
minitwit_grafana      replicated   1/1        ghcr.io/devtroopers-itu/minitwit-grafana:latest
minitwit_loki         replicated   1/1        grafana/loki:3.6.7
minitwit_promtail     global       3/3        grafana/promtail:3.6.0
```

Source: `ssh root@64.226.116.162 docker stack services minitwit`.

---

## Maintenance (lines 143–154)

### Webserver log — "what nobody hit, didn't get fixed" — perfect example

**ADDS — quantified evidence for the paragraph's thesis:**
The Hetzner webserver container log (24 days, 21 Apr → 15 May) contains
**64,981 lines, almost all the same warning**:

```
http: superfluous response.WriteHeader call from main.(*responseWriter).WriteHeader (main.go:50)
```

Fires every 30 seconds — almost certainly the Docker healthcheck `wget` hitting `/`.
Bug nobody fixed; nobody hit it directly. If the section has room for one extra
example beside `helpers.go:55` and `sim_api.go:33`, this one is concrete and
quantifiable. It's also a defensible "the warning was visible all along; we
just weren't reading our own logs" beat.

Source: `zcat snapshot/logs/itu-minitwit-webserver-1.log.gz | wc -l` → 64,981.

### Hetzner cert renewal — failed silently

Adjacent to the above: cert renewal was visibly broken in Traefik logs for
weeks on Hetzner and nobody acted on it. Same "nobody owned maintenance" theme,
if you want a second line. Probably skip — Operation already covers it.

---

## DevOps Style (lines 156–169)

### Continual Learning — quality tooling visible on the box

**CONFIRMS the section's quality-gate framing:**
Hetzner snapshot has running Grafana + Prometheus + Loki + Promtail containers,
4-week uptime (matches "Prometheus mid-March" if the original deploy was on
Hetzner first). The DO Swarm runs the same six containers plus Promtail as a
3/3 global service. Nothing in either snapshot disputes the timeline; just
note that Hetzner and DO each carried the full monitoring stack independently
during the parallel run.

### Use of Generative AI — concrete data from the box

**ADDS — 11 invocations of `claude` in /root/.bash_history on Hetzner:**
The team used Claude **on the production host itself** at least 11 times
during operational work (counted from `awk '{print $1}' history/bash_history | sort | uniq -c`).
Not just a local development tool. Tangentially relevant to the GAI section's
"asymmetric use" observation; skip unless you want a half-sentence in the GAI
paragraph about Claude being invoked during ops, not only during coding.

### Migration cost — "Feedback" arc, if the budget has room

**ADDS — pgloader iteration count:**
The SQLite → Postgres data migration (session 5) cost **25 pgloader
invocations** and **5 separate import logs** (`import.log` through `import5.log`)
before it landed clean. Visible in `snapshot/history/bash_history` and
`snapshot/migration-logs/`. If the Continual Learning paragraph wants one
concrete "we iterated noisily" example, this is it. Probably skip — the
section already names enough.

---

## Cross-section: dates the wiring agent should normalise

| Where | Current draft says | Snapshot/journal says |
|---|---|---|
| Operation §1 | "On 17 April..." | Outage started **16 April 19:38 UTC** (21:38 CEST), fix landed **17 April 13:18 UTC** |
| Operation §1 | "Hetzner and DO Swarm... 10 April to 4 May" | DNS cutover 10 Apr ✓, last DO deploy 04 May 15:01 UTC ✓ |
| Rebuild 3 | "UFW" | UFW on the DO Swarm nodes specifically (Hetzner had it disabled) |
| Rebuild 3 / Operation | "Traefik to v3.6 (PR #131)" | Specifically v3.6.13 on DO; Hetzner stayed on v3.4 |

---

## Secrets handling — read this before drafting

The snapshot directory contains live production credentials at the time of capture:

- `snapshot/server-repo/root-env-file.env` — `HCLOUD_TOKEN`, `DATABASE_URL`, `POSTGRES_PASSWORD`, `SECRET_KEY`.
- `snapshot/history/bash_history` line 399 — a Postgres password in clear text in a `pgloader` invocation.
- `report/research-notes-leo.md` already flags the Grafana admin password posted in `#generelt`.

**Do not** inline any of these into `report.md` quotes. Cite the snapshot by file
path (e.g. `snapshot/system/firewall-and-listeners.txt`) when evidence is needed,
not by content. The snapshot directory is at `/home/leos/Dev/itu-minitwit-hetzner-snapshot/`
— **not** committed to the repo. Keep it that way.

---

## Reproducing these queries

```bash
# Hetzner snapshot (local)
ls /home/leos/Dev/itu-minitwit-hetzner-snapshot/
zcat .../logs/itu-minitwit-webserver-1.log.gz | wc -l
grep "ACME challenge" .../logs/itu-minitwit-traefik-1.log | awk '{print substr($1,1,10)}' | sort | uniq -c

# DO Swarm (live, requires SSH from this workstation)
ssh root@64.226.116.162 docker node ls
ssh root@64.226.116.162 docker stack services minitwit
ssh root@64.226.116.162 "journalctl -u docker.service --since '2026-04-16' --until '2026-04-18' | grep '7946: i/o timeout' | awk '{print \$1,\$2}' | sort | uniq -c"
ssh root@64.226.116.162 ufw status
```
