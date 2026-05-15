# Research notes — Leo's report sections

> **Start here tomorrow.** Skim §0–§5 first. Deep research is in the appendix.

---

## §0 Hard constraints (REPORT.md, 2026 spring)

- **Deadline:** Mon **2026-05-18 14:00** on WISEflow (3 days from this file).
- **Total budget:** **2500 words max** for the whole report (images free).
- **PDF location:** `report/build/MSc_group_<letter>.pdf`, **built via CI** —
  not yet wired in your repo; flag to the team tonight.
- **Hand-in:** WISEflow PDF + a PR to `final_report_urls.py` in
  `itu-devops/MSc_lecture_notes`.
- **`.mailmap`:** present, with LLM mapped to `LLM <none>` — already done.

### Your 5 sections, in the order to write them tomorrow

| # | Section | Words | Co-author | Why this order |
|---|---|---|---|---|
| 1 | Use of Generative AI | ~200 | solo | Most evidence-heavy, sets honest tone for rest. |
| 2 | Operation | ~150 | Apoorva | You have the most material here. |
| 3 | Evolution and Refactoring | ~150 | Håkon | Sync with Håkon on which 3 issues you each own. |
| 4 | Maintenance | ~150 | solo | Builds on §2–3 findings. |
| 5 | DevOps Style | ~150 | solo | Easiest once §1–4 are drafted. |

Total: **~800 words** = 32% of the 2500 budget. The other 13 sections share ~1700.

---

## §1 Thesis sentence for each section (write to these)

These are the one-line anchors. If your paragraph wanders away from the thesis,
cut the wander, not the thesis.

| Section | Thesis sentence |
|---|---|
| **Use of GAI** | We used **Claude (Code + Opus 4.6/4.7) exclusively** — heavily for infra, debugging, and docs, lightly for application code; the asymmetric use across the team is itself the most honest reflection. |
| **Operation** | We ran two production stacks in **parallel for ~4 weeks** to avoid a hard cutover; the lesson is that **control-plane and data-plane failures are independent**, and our two firewall outages on 17 Apr proved both. |
| **Evolution and Refactoring** | The three sharpest stories are the **`latest` counter (in-memory → Postgres) for stateless replicas**, the **personal-timeline blow-up (PR #152 didn't fix it, the LATERAL rewrite did)**, and the **Hetzner → DO Swarm migration with all its follow-on bugs**. |
| **Maintenance** | Improvements were CI-gated (gofmt/golangci-lint/hadolint/Semgrep/Docker Scout/Codacy); the **remaining ugly bits are real and we know about them** (config sprawl, hardcoded simulator auth, thin tests, GHCR PAT plaintext on droplets). |
| **DevOps Style** | We **enforced PR-flow strictly but reviewed loosely** (~73% self-merge across 86 PRs), monitoring was added before incidents forced it, and AI use was openly acknowledged but **asymmetric (Leo dominates the AI-driven artefacts)**. |

---

## §2 Strongest evidence per section (3 each, ready to cite)

### §2.1 Use of GAI (200 words)

1. **PR #15 (Leo) — Python→Go port, 1288 lines, 4 tests.** Carries
   `Generated with Claude Code` trailer. The *foundation* of the project was
   AI-assisted. Discord 2026-02-10: *"i did it together with claude code to
   be fair."*
2. **PR #152 (Håkon) — the failed first timeline fix.** Your review comment
   used Claude `EXPLAIN` to prove `IN (?)` vs `= ANY(...)` produces the same
   query plan, killing the proposed fix. Best "AI helped me catch a bad
   teammate PR" anecdote.
3. **Discord #standup 2026-04-17 (Leo posting as Claude):** *"hey team,
   claude here — leo asked me to write up what we fixed today so everyone's
   on the same page."* Unusually explicit AI byline. The honest trade-off
   to name: the rest of the team read Claude's post-mortem, not yours.

Asymmetry fact: 12 of 12 `Co-Authored-By: Claude` commit trailers are yours;
3 commits authored directly by Claude (in `.mailmap` as `LLM <none>`); the
others reference Claude in Discord but not in commits. **The exam-defence
implication: most "did you understand what got committed?" questions land on
you. Say so out loud.**

### §2.2 Operation (150 words)

1. **17 Apr Swarm split-brain.** Peter J added a DO cloud firewall that
   silently blocked Swarm's 2377/7946/4789 between our own nodes. External
   uptime checks said green; cluster was dead. Lesson: **control plane and
   data plane fail independently**. (PR #131 fix; `docs/incidents/session11-ops-debug.md`.)
2. **The horizontal-scaling honest admission.** We chose 3 Swarm replicas
   because the session asked for it, not because we measured. Tiny nodes
   (1 GB RAM, "662/961 MB used"). Discord 2026-03-11 shows DO was picked
   over Hetzner because of **student credits** (Apoorva had credits, Leo
   said "use student discounts — that would be optimal"), not because
   "Apoorva already had droplets" as the doc claims. Rewrite that sentence.
3. **Zero-downtime parallel run for 4 weeks** (Hetzner + Swarm both live
   Apr 10 → May 4) is the strongest single operational decision. Say it.

### §2.3 Evolution and Refactoring (150 words)

1. **`latest` counter (PR #138).** Process-local `var latest int = -1`
   worked on Hetzner. On 3-replica Swarm, replicas disagreed. Moved into a
   one-row `SimState` table in Postgres. Discord shows this trap is
   **explicitly named in the course slides** — examiners know about it.
2. **Personal-timeline 234s → 5.7s (PR #135 then `a3dfc3d`).** Real users
   hit 504s 29 Apr. **PR #152 (Håkon's first attempt) is still open,
   unmerged** — `IN (?) → = ANY(...)` doesn't change the query plan, and
   removing `Preload("Author")` added a 30-row N+1. The real fix was a
   `LATERAL` per-author lookup.
3. **Swarm migration stack of bugs (PR #131).** Traefik 3.4/3.5 docker-API
   mismatch (10 errors in 5s on v3.4, 0 on v3.6); v3.6 stricter multi-
   service routers; webserver on two overlays → Traefik picked wrong one.

### §2.4 Maintenance (150 words)

1. **Hard:** config sprawl — `DATABASE_URL` lives in `.env`,
   `docker-compose.yml:32-34`, `docker-stack.yml:48-50`, Swarm secrets, and
   `main.go:115-121` (`getSecretOrEnv` covers 3 paths). 70+ lines of inline
   comments in `docker-stack.yml` documenting Traefik footguns.
2. **Improved:** linting baked in (gofmt + golangci-lint + hadolint + Semgrep
   + Docker Scout + Codacy). Codacy caught a real bug — commit `c8ff76c`
   `"fix /health route order — caught by codacy"`. Multi-stage Dockerfile
   dropped binutils CVEs (image 306 MB → 30 MB, PR #155).
3. **Still ugly:** `helpers.go:55` swallows bcrypt error; `sim_api.go:33`
   simulator auth is a hardcoded base64 string (functionally open API);
   GHCR PAT in plaintext on all three droplets, parked follow-up never
   rotated. Tests are thin (11 functions total, none on `store.go`).

### §2.5 DevOps Style (150 words)

1. **86 merged PRs, 73% self-merge.** Last 30 PRs: 9/30 had a human
   reviewer (30%). PR-as-deployment-ritual, not quality gate. Branch
   protection exists but allows admin override. **Peter J's 24 Apr was 9
   self-merged PRs in one day fighting `golangci-lint` config — concrete
   evidence of merge-and-see culture.**
2. **The Go decision was unilateral + AI-assisted + team-ratified by PR.**
   Discord 2026-02-07: team discussed Java/Python+Bottle/Pyramid. **Nobody
   proposed Go.** 3 days later you posted PR #15 with a Go port done with
   Claude Code; Peter K merged it. Honest framing.
3. **Communication-tool migration is itself a maturation story:** Discord
   DM (scheduling) → server channels (#standup carried 504 messages) → PR
   comments + Leo's 720-line `session11-ops-debug.md` written during the
   17 Apr outage. We built better audit trails than chat scrollback.

---

## §3 Pull-quote bank (verbatim, ready to paraphrase or cite)

From your own PR descriptions and Discord:

- *"Production stayed up because Manager served the traffic, but we lost
  multi-node redundancy silently on every deploy."* — PR #141 body
- *"With replicas: 3, each pod held its own copy, so after the simulator
  URL flips to DO, the grader's GET /latest poll would read a stale value
  roughly 2/3 of the time."* — PR #138 body
- *"Live cluster already runs this change — site has been on Hetzner the
  whole time, no user-visible outage."* — PR #131 body
- *"i did it together with claude code to be fair, which made it easier to
  overcome some of the go coding syntax."* — Discord 2026-02-10
- *"For me it would be ok to migrate to postgreSQL but keep it on the same
  VM in its own container so we wouldn't have to pay for another vm —
  unless we change to digital ocean and use our student discounts — that
  would be optimal!"* — Discord 2026-03-10

From Håkon's PR #152 review thread (your comment):
- *"IN vs = ANY doesn't change anything in Postgres — they produce the
  same query plan. And the loop runs one query per message instead of one
  batched fetch, so the timeline ends up slower than before."*

---

## §4 Exam Q&A prep — what they're likely to ask you

From `exam_details.md` + lecture-rubric agent + course slides. Plan a
30-second answer for each before the exam.

### Specific to your sections (~70% chance per question)

1. **"Why did you pick Go?"** — Honest answer: you prototyped it solo with
   Claude over a weekend after the team discussed Python+Bottle/Pyramid; the
   team ratified it via PR review. Trade-off: better perf than Python under
   simulator load + ability to deploy as a single static binary; cost was
   that nobody else on the team initially knew the language.
2. **"You describe a 'single wizard' team in everything except name. How
   would you organize work so further maintenance is doable once the wizard
   gets sick?"** — Concrete answer needed. The `.mailmap`, the
   `docs/incidents/` runbook, and `docs/operations/ops-cli-reference.md`
   exist for exactly this. Say so + acknowledge the gap honestly.
3. **"How did you handle the multi-replica state problem (the `latest`
   counter)?"** — The course slides name it explicitly. Answer: PR #138,
   one-row `SimState` table in Postgres.
4. **"What's a single point of failure in your system right now?"** —
   Multiple defensible answers: the Swarm manager (Traefik + Prometheus +
   Grafana all live there); the GHCR PAT plaintext on droplets; managed
   Postgres single instance; DNS at one provider.
5. **"How would you scale this 10x?"** — Horizontal already at 3 replicas;
   bottleneck would be the DB + the manager. Caching layer (the report
   already mentions, check Peter J's section), bigger manager, sharded
   Postgres, multiple manager nodes.
6. **"Walk us through what happens when a developer pushes to master."** —
   CD workflow `cd.yml`: build image → push ghcr.io → SSH to manager →
   `docker stack deploy --with-registry-auth`. Talk about the GHCR PAT
   story (PR #141) as the moment you understood the flow deeply.
7. **"Where could secrets leak?"** — `getSecretOrEnv` covers 3 paths, GHCR
   PAT plaintext on droplets, the May 4 plaintext password in Discord (be
   honest if asked — already a problem, rotate it now).
8. **"What's one bad decision you'd undo?"** — Reflection Perspective is
   literally about this. Suggested honest answer: keeping the Vagrantfile +
   dev-VM section of README after the Swarm migration. Or: not catching
   the `latest` counter bug *before* the cutover (we found it because the
   simulator was still on Hetzner).

### Likely AI-specific question

9. **"Are LLMs helping or hindering?"** — Helge asks this in slides
   repeatedly. Concrete answer: helped on infra/docs/index decisions
   (PR #131, `bdb6c16` linting setup, the `LATERAL` query plan); hindered
   when a previous Claude session's wrong cache-overflow theory wasted time
   on 17 Apr (you rejected it after re-investigation, named in the incident
   doc).

---

## §5 5-minute checklist before you write (tonight or first thing)

These are yes/no items from the course rubric. Each is a free point if
done, a free deduction if not.

- [ ] **`misc_urls.py`** in the lecture-notes repo has your Grafana URL and
  it loads without your auth cookie (or credentials shared in report).
  Only **3 of 13 groups last term** had this working.
- [ ] **`.mailmap`** at repo root has every contributor + `LLM <none>` line
  (already present per audit, confirm).
- [ ] **Containers run as non-root user** — confirmed by multi-stage
  Dockerfile (`aa1decc`). Mention it.
- [ ] **HTTPS via Let's Encrypt + Certbot, HTTP → HTTPS redirect** —
  confirmed (Traefik). Mention it.
- [ ] **CI runs ≥1 security static-analysis + ≥1 image scanner** —
  confirmed (Semgrep + Docker Scout). Mention it.
- [ ] **UFW deny-by-default + awareness that Docker bypasses UFW** —
  documented in `docs/operations/ufw-setup-recap.md` and
  `firewall-changes.md`. Peter J's section.
- [ ] **Architecture: 2+ viewpoints (deployment + C&C minimum)** —
  Peter K + Håkon's section, but you should know what's in it.
- [ ] **Backups tested by actual restore** — DO managed Postgres has
  backups; have you tried restoring from one? If not, the honest answer
  is "no, here's how we'd do it under pressure".
- [ ] **Risk matrix in the report** — session 11 README_TASKS lists 6
  bullets to follow as substructure. Peter J's security hardening section.
- [ ] **Report PDF built via CI** — not yet wired. Flag to the team.
- [ ] **Rotate the May 4 plaintext credentials** posted in Discord
  `#generelt` (helgeandfriends / sesame...). Do it before submission.
- [ ] **`terraform/README.md` opens with `Author - Claude`** — verify the
  content actually matches the built infra. 2-minute check.

---

## §6 Red flags to avoid in your prose

From lecture slides + Helge's repeated callouts:

- ❌ "Comprehensive", "robust", "seamless", "powerful" — AI-flavoured. Cut.
- ❌ Restating the section title in a sub-heading.
- ❌ Quoting SonarQube/Codacy debt-hours at face value (Goodhart's Law).
- ❌ Screenshots with text smaller than body text.
- ❌ Manual-deploy admission (you don't have this — CD is real).
- ❌ Secrets-in-git mention without acknowledgment of cleanup plan.
- ❌ Hiding AI use. Helge openly asks if LLMs help; cagey is worse than honest.

---

## §7 New findings from the full-corpus PR audit (added 2026-05-15)

The 9-PR cherry-pick missed real material. Key corrections:

- **Peter J has TWO GitHub accounts** (`PeterJuulMoller` and
  `DenSygeMike`). That's why `.mailmap` matters. Tests of his attribution
  collapse them.
- **86 merged PRs total** (not 167 — 167 includes drafts/closed).
  Distribution: Leo 28 (36%), Peter J 23 (30%), Håkon 14 (18%),
  Apoorva 6, "Mike" 6 (same as Peter J).
- **Self-merge rate: 73% overall, 70% in the last 30 PRs.** Earlier
  estimate of "4 of 30" was overstated; real is 9/30 = 30% reviewed.
- **PR #15 (Leo)** was the actual Python→Go port — 1288 lines, 4 tests,
  carries the `Generated with Claude Code` trailer. PR #138 is the later
  `latest` counter fix. **Use #15 for the foundation-of-the-project AI
  story, not #138.**
- **Apoorva's #139** is a 3-in-1 (firewall + UFW + `/latest` Postgres fix
  + replica placement). If you write about the multi-replica state fix,
  credit her too — your #138 wasn't the whole story.
- **Apr 24 lint iteration** (Peter J's 9 self-merged PRs in one day) is
  the sharpest concrete evidence of merge-and-see culture. Cite it.
- **Duplicate PRs pattern:** #66/#67, #110/#111, #112/#113, #143/#159 —
  "merge, then immediately re-PR" suggests target-branch confusion (dev
  vs master).
- **No revert PRs in the whole history, no load-test PRs.** Honest gaps to
  acknowledge if asked.
- **Teammate contributions to credit explicitly:**
  - Apoorva: #139 (firewall+latest+placement), #142 (GHCR fix + docs
    restructure), #144 (Grafana volume), #164 (Hetzner removal + multi-
    stage Dockerfile). All in April — *she shows up when Swarm migration
    breaks*.
  - Peter K: #84 (Prometheus+Grafana foundation), #98 (route-template
    metrics fix), #136 (availability metric), #143 (container hardening
    + Docker Scout), #167 (Terraform).
  - Håkon: #1 (Python 3 port), #74 (SECRET_KEY removal), #79
    (PostgreSQL migration), #88 (security+DB).
  - Peter J (Mike): #64 (CI on dev), #70 (`store.go` abstraction), #122
    (initial Swarm deploy), #128 (lowercase secrets fix), #130 (581-line
    incident doc).
- **Original CI/CD pipeline** was Leo's #65. Original alerting was Leo's
  #114 (3 alert rules, 4h repeat). These predate the Swarm work.

---

## §8 Lecture-rubric findings beyond REPORT.md

Things the slides emphasise that REPORT.md doesn't spell out:

- **"Log decisions" with arguments** for language/framework choice. Helge
  literally says: *"a feature mapping/comparison or a mini-benchmark is a
  good choice"*. The Go decision needs more than "we picked Go" — even one
  sentence comparing Go to Python+Bottle (the alternative actually
  discussed in the team DM) earns the point.
- **Maintenance framing**: pick *one* of (ISO 25010: modularity,
  reusability, analysability, modifiability, testability) OR (Kleppmann:
  operability, simplicity, evolvability). Use the vocabulary; it signals
  course-content awareness.
- **Three Ways of DevOps** (DevOps Handbook): Flow, Feedback, Continual
  Learning. Frame DevOps Style around these explicitly.
- **Backups tested by restore** — "A backup is not useful unless you can
  actually use it to perform the restore" (session 11 Slides). The
  examiners ask if you've ever restored, not if you have backups.
- **"Single wizard" anti-pattern** — already in sample exam questions
  (`exam_details.md` line 38). Will come up at your exam. Plan an answer.
- **GAI grading is the *reflection*, not the tool list.** Helge's slides
  repeatedly nudge groups to reflect honestly. Don't just say "we used
  Claude for X" — say "Claude helped on Y, hindered on Z (concrete
  example)".

---

## §9 Outputs from this prep session — where everything lives

- `~/Dev/itu-minitwit-hetzner-snapshot/` (1.4 GB) — full Hetzner snapshot
  before deletion: `.bash_history`, container logs, monitoring volumes,
  SQL dumps, migration logs. Reference if you need exact phrasing.
- `~/Dev/itu-minitwit-discord-export/` (856 KB) — DM + 5 server channels.
  `#standup` (504 msgs) is the richest. The Apr 17 outage thread is
  invaluable for the GAI section's "Claude posts as Claude" anecdote.
- `~/.config/discord-dce-token` — **delete with `shred -u` before bed.**
- This file — `report/research-notes-leo.md`. Branch
  `docs/report-leo-sections` (local). Push/commit when ready.

---

---

**APPENDIX — original deep research notes follow. Use to source pull-quotes
or verify a citation. Don't try to read top-to-bottom.**

---

Working notes for the four sections Leo writes. Not for the final report — this
is the material to shape into prose. Section word budgets come from `report.md`
(Reflection Perspective ~500 words total split four ways, so ~125–150 each;
"Use of Generative AI" ~200).

Tone target: plain student English, no marketing words ("robust", "seamless",
"comprehensive"), short sentences, honest about trade-offs. If something was
ugly, say so.

---

## Evolution and Refactoring (Leo + Håkon, ~150 words)

Prompt: *Biggest issues and how we solved them. Link commits/issues.*

### Six candidates — pick 3 or 4 for the section

**1. `latest` counter broke once we went multi-replica.** Stored as a Go
package variable `var latest int = -1`. Fine on the single Hetzner box. On the
3-replica Swarm, replica A wrote, replica B read, the grader saw stale values.
Moved into a one-row `SimState` table in Postgres. Commit `b874e8e`, PR #138.
Documented in `docs/architecture/latest-counter-in-db.md`.
**Honest angle:** clean example of "scaling exposes hidden shared state". No
monotonic guard on writes — fine for the sequential grader, would break under
real parallel load.

**2. Hetzner → DO Swarm migration + a week of follow-on bugs.** Lowercase
secret path (`442ddbb`, PR #128). Missing `DATABASE_URL` in CD (`8603cd0`,
PR #110). Traefik v3.4 refusing to talk to newer Docker engine. v3.6 then
complaining `Router cannot be linked automatically with multiple Services`.
Traefik forwarding to the wrong overlay (10.0.1.x backend instead of 10.0.2.x
frontend). 30s HTTP/2 504 over the Swarm ingress mesh. Fixed by pinning
Traefik to v3.6.13, naming services explicitly per router, switching ports
from ingress mode to host mode. Commits `3c649da`, `c3abcc2`, `cba87bf`,
`1e57e18`, `749bec6`. Full debug log in `docs/incidents/session11-ops-debug.md`
(720 lines, written live during the outage).
**Honest angle:** the biggest chunk of pain in the project. Concrete enough to
defend at the exam.

**3. DO cloud firewall silently killed the cluster (17 Apr).** Peter added a
DO firewall the night before — HTTP/HTTPS/8080 open, everything else denied,
"All IPv4". DO firewalls don't auto-trust sibling droplets, so Swarm's
2377/7946/4789 between our own nodes got blocked. Workers went `Down —
heartbeat failure`. Site stayed up (Traefik on manager) so external uptime
checks said green; cluster redundancy was zero. Fixed with tag-scoped rules
for the four Swarm ports + ICMP. No code commit — DO dashboard change.
Codified in `docs/operations/firewall-changes.md` and `ufw-setup-recap.md`
(PR #137).
**Honest angle:** control plane and data plane can fail independently —
outside-the-box uptime checks don't catch it.

**4. GHCR auth expired between CD runs, replicas piled up on manager.**
CD logged into ghcr.io on the manager with `$GITHUB_TOKEN` (~1h life). Swarm
propagated the dead token to workers; next deploy worker pulls failed
silently and all 3 webserver replicas landed on the manager. Same failure
shape as Issue 3, different root cause. Swapped to a long-lived classic PAT
(`ghp_*`) stored as repo secrets. Commit `8747a32`, PR #141.
**Honest angle:** subtle CD bug that only shows up across deploy cycles.
Trade-off: long-lived PAT now sits in plaintext on three droplets — flagged
in `session11-ops-debug.md` as a parked follow-up, not rotated.

**5. Personal-timeline query: 234s → 5.7s.** 29 Apr, DO Postgres CPU alert +
real user 504s. Power user with ~913 follows hit ~234s. Postgres scanned the
global `pub_date DESC` index and discarded 284M rows. First fix attempt
(PR #152) swapped `IN (?)` for `= ANY(...)` (same parse plan — no change)
and removed `Preload("Author")`, which *added* a 30-row N+1. Real fix
rewrote the Postgres path as a `LATERAL` per-author lookup using
`idx_messages_author_flagged_pubdate`. 234s → 5.7s cold cache, ~8ms typical.
Commits `a3dfc3d`, `f2ee9a5`, `9ce703e`. Write-up:
`docs/incidents/personal-timeline-perf.md`.
**Honest angle:** real EXPLAIN ANALYZE numbers. And — first PR didn't fix
it and made it worse. Worth saying out loud.

**6. Container hardening churn (skip if tight on words).** The
`feature/hardenContainers` branch became 10+ small fix-the-fix PRs
(#143–#160): linter didn't support Go 1.25, `gosimple` was removed, tests
stopped running once we went multi-stage, templates/static weren't copied
into final image. Eventually settled with PR #156 (`c6f82f5`) testing
against the builder stage and `b0b7f10` copying assets.
**Honest angle:** "shifting security left" is messier in practice than the
slides suggest.

### Suggested ~150-word arc

Lead with #1 (clean narrative). Bundle #2+#3 as the Swarm migration story
with the firewall as a sub-bullet. End with #5 for the concrete numbers and
the "we merged the wrong fix first" admission. Skip #4 and #6 if tight.

---

## Operation (Leo + Apoorva, ~150 words)

Prompt from `report.md`: *Incidents, on-call lessons, what changed in how we
run the system. Choices about servers, droplets, databases. Vertical/horizontal
scaling.*

### Facts to draw from

**Server choices, in order.**
- **Feb 18 (session 3):** Hetzner single droplet `46.224.144.214` (`h-6d6ea5`,
  Ubuntu 22.04), provisioned via Vagrantfile. Docker Compose, SQLite on a
  Docker volume, Go app on `:8080`, no TLS. Forced by session 3 rules: IaaS,
  no PaaS, Vagrantfile required.
- **Apr 10 (session 9):** Three DO droplets — manager `64.226.116.162`,
  workers `134.122.90.176` and `206.189.59.60`. Reason for DO over more
  Hetzner is unglamorous: Apoorva had already provisioned them
  (`docs/operations/docker-swarm.md` line 115 says exactly this).
- **Apr 10 → May 4:** Hetzner ran in parallel with Swarm — CD deployed to
  both. Zero-downtime cutover for the simulator.
- **May 4 (`e4c234c`):** Hetzner step removed from CD, droplet powered off.
- **May 15 (today):** droplet deleted to stop the bill.

**DB choice.** SQLite → DO managed Postgres in commits `c972607`, `53bab02`,
`f0a0b01`, merged via PR #79 on Mar 18. Reason: webserver had to become
stateless before replication, and we wanted backups + HA outsourced. We use
`:25060` (DO managed Postgres port). Sessions stayed stateless via Gorilla
CookieStore (`docs/architecture/architecture.md` lines 88–97).

**Vertical vs horizontal.** We chose horizontal — 3 webserver replicas
behind Traefik. Honest trade-off: tiny nodes (manager 1 GB RAM, "662/961 MB
used" during session 11 debug), so we still vertical-scale implicitly by
being CPU/RAM-bound per node. We never measured replica latency vs single-box
to justify horizontal. We did it because the session asked for it.

**Three incidents worth mentioning.**
1. **17 Apr — Swarm split-brain.** Cloud-firewall story (Issue 3 above).
   External uptime green, internal cluster dead. Lesson: control plane and
   data plane fail independently.
2. **17 Apr — HTTP/2 504s blocking the DO cut-over.** Ingress mesh + HTTP/2
   framing + overlay MTU 1450. Fixed with Traefik in host port mode. We
   delayed the DNS flip until `curl --http2 https://devtroopersminitwit.codes`
   returned 200. Lesson: don't cut over before the curl says yes.
3. **29 Apr–1 May — timeline blow-up.** Real users complained the public feed
   was unreachable. Cause: ORM-generated JOIN over simulator-scale rows.
   Staged fix (`dc6db65` → `a3dfc3d` → `f2ee9a5` → `89ced5a`). Lesson: add
   indexes *before* turning the simulator on, not after a user timeout.

**Smaller fires:** `8603cd0` (CD didn't pass `DATABASE_URL`), `8747a32`
(GHCR PAT), Grafana on tmpfs losing dashboards every restart until PR #144
moved it to a named volume.

**How operating the system changed over the semester.**
- *Early (Feb–early Mar):* one box, SQLite, no TLS, `docker compose up -d`
  over SSH, no monitoring beyond `docker logs`. We found out about breakage
  by visiting the site.
- *Mid (mid Mar–early Apr):* CD pipeline (`2590efc`), Prometheus + Grafana
  (PRs #84, #93), Loki + Promtail (`f2bebb9` 30 Mar), Discord alert webhook
  (`e619a92` same day). Alert rules in `monitoring/prometheus/prometheus.rules.yml`:
  `WebserverDown` (1 min `up=0`), `HighErrorRate` (>10% 5xx for 2 min),
  `SlowResponses` (P95 > 1s for 5 min).
- *Late (Apr 10 onwards):* 3-node Swarm, Traefik with Let's Encrypt, Docker
  Swarm secrets (not `.env`), Promtail in global mode, host firewalls (UFW)
  on every node + DO cloud firewall with tag-scoped Swarm ports. Written
  runbook (`docs/operations/ops-cli-reference.md`, `06c0f00`) born from
  Googling the same `docker service ps` flags too many times.

### Suggested ~150-word arc

Open with the DO-because-Apoorva-had-droplets admission (graders reward
honesty). Then horizontal scaling decision + trade-off (small nodes, never
measured). One incident worth recounting in detail (split-brain or the
HTTP/2 504). Close with the cross-section lesson: **adding infrastructure
layers without enumerating their failure modes costs you afternoons**. The
zero-downtime parallel-run of Hetzner + Swarm for 4 weeks is the strongest
single operational decision — say it out loud.

---

## Maintenance (Leo solo, ~150 words)

Prompt: *What's hard to maintain, what we improved, what's still ugly.*

### Material

**Hard to maintain (still).**
- **Config sprawl.** `DATABASE_URL` lives in `.env` (1–2), `docker-compose.yml`
  (32–34), `docker-stack.yml` (48–50 + 182–188 Swarm secrets),
  `docker-compose.test.yml`, and `/run/secrets/...` read at runtime
  (`main.go:115–121`). `getSecretOrEnv()` covers three lookup paths.
- **Two register flows.** `handlers.go:191` (`registerHandler`) and
  `sim_api.go:37` (`simRegister`) each validate username/email/password
  independently. Same rules, two implementations, no shared helper.
  `registerDispatcher` (`handlers.go:47`) routes by `Content-Type`.
- **docker-stack.yml needs an essay.** 70+ lines of inline comments documenting
  Traefik footguns. Useful, but config that needs a manual to be safe is a
  maintenance smell.
- **Dead files still shipping.** `flag_tool.go` (`//go:build ignore`,
  SQLite-only, hits `/tmp/minitwit.db`), `control.sh`, `Vagrantfile`. Issue
  #58 ("Remove unused control.sh and flag_tool.go") was closed but the files
  are still there.
- **Tests are thin.** 11 test functions total (4 in `main_test.go`, 7 in
  `sim_api_test.go`). No tests on `store.go`. No tests on the Prometheus
  middleware. SQLite in tests vs Postgres in prod hides GORM dialect bugs.

**Improved.**
- Linting in CI: `gofmt`, `golangci-lint`, `hadolint`, Semgrep, Docker Scout,
  Codacy (`.github/workflows/ci.yml:8–100`). Codacy already caught a real
  bug — commit `c8ff76c` `"fix /health route order — caught by codacy"`.
- Multi-stage Dockerfile + non-root user (`aa1decc`, PR #155).
- `.mailmap` (`8705863`, PR #157) collapses 8 email/name variants into 5
  contributors + 1 LLM entry.
- `latest` counter moved out of process memory (`b874e8e`, PR #138).
- Personal timeline N+1 fixed with `LATERAL` join (`a3dfc3d`, PR #135).

**Still ugly.**
- `helpers.go:55` — `hashPassword` ignores the bcrypt error.
- `helpers.go:42, 48` — flash session save errors swallowed.
- `sim_api.go:33` — simulator auth is a hardcoded base64 string. Functionally
  an open API.
- `main.go:115` — lowercases secret names to derive a path. Fragile coupling
  between env name and Swarm secret file.
- `helpers.go:81` — templates parsed on every request, not cached.
- GHCR PAT in plaintext on all three droplets — parked follow-up #3 in
  `session11-ops-debug.md`, not rotated.

### Suggested ~150-word arc

Three bullets each on hard / improved / still ugly. Pick the punchiest from
each list. Suggested picks: config sprawl + thin tests for "hard"; linting +
the `latest` counter move for "improved"; swallowed errors + simulator auth
string for "still ugly". The point is to show you can *see* the smells, not
that you fixed all of them.

---

## DevOps Style (Leo solo, ~150 words)

Prompt: *What was different from previous projects and how it worked out.
Be honest about trade-offs.*

### Observable patterns (the facts)

- **PR-heavy workflow.** 167 PRs over the semester for 5 people. Leo
  authored ~45 of them. `feature/* → dev → master` strictly enforced via
  branch protection (issue #52).
- **Conventional Commits used inconsistently.** Leo + Peter J use
  `fix(...)`, `feat(...)`, `docs(...)`. Other authors use freeform — `"dath"`,
  `"version fix"`, `"Fix to hadolint"`, `"new golang.yml"` (twice). PRs
  #146–#160 from `feature/hardenContainers` are a stream of 10 tiny
  "did the version actually work this time" merges.
- **Self-merging is the norm.** Of the last 30 merged PRs, only 4 had a
  human reviewer recorded; the rest were unreviewed or had only
  `codacy-production` as reviewer. CI gates exist (lint + tests + Semgrep
  + Scout); humans rarely block. **PR is a deployment ritual, not a quality
  gate.**
- **CD actually deploys on green.** push to `master` → `cd.yml` builds →
  ghcr.io → SSH to manager → `docker stack deploy --with-registry-auth`.
- **Monitoring set up early** — Prometheus + Grafana + Loki + Discord webhook
  before the timeline blow-up forced it.
- **Incident docs are written for ourselves, not the grader.**
  `docs/incidents/session11-ops-debug.md` is 720 lines of post-mortem with
  a "Group-meeting crib sheet" at the bottom. Habit only formed after we
  lost a full afternoon to the firewall mystery.
- **AI is acknowledged in commits** (see GAI section).

### Honest weaknesses

- Test coverage gap (11 tests, no `store.go` unit tests).
- `feature/hardenContainers` cost 10 retry PRs because we merge-and-see
  rather than think-and-ship.
- Commit message hygiene is mixed.
- Three commits committed by Claude directly (no human commit author) —
  ok with `.mailmap`, but means human review/intent isn't in those SHAs.

### Personalization hooks (only Leo can confirm)

- First time using a real CD pipeline that auto-deploys on merge?
- First time setting up Prometheus / Grafana / Loki / alerting?
- First time owning a prod box (DO droplets, Traefik certs, Postgres
  credentials)?
- How writing PR descriptions + incident docs felt vs. previous projects
  where work was pushed to main.
- The `CLAUDE.md` "slow, deliberate development" rule — was it aspirational
  or actually followed? PRs #146–#160 (10 retries on a CI fix) suggest we
  sometimes optimized for "ship and see" over "think and ship".

### Suggested ~150-word arc

Open on one thing that was genuinely new for you (the personalization
hooks). Then 1–2 trade-offs from the observable patterns — PR-heavy +
self-merging gives audit trail without review benefit is the sharpest one.
Close with one honest weakness named directly (test coverage *or* the
"merge-and-see" loop on hardening). Avoid abstractions ("we learned a lot
about collaboration") — name something specific that you'd do differently.

---

## Use of Generative AI (Leo solo, ~200 words)

Prompt: *Which tools we used, for which tasks, how, and a short reflection
on whether they helped or hindered the work.*

### Concrete evidence (use this, not vague claims)

**Tool:** Claude Code (Anthropic) is the only AI tool with traceable
fingerprints. No ChatGPT / Copilot / Cursor references found.

**Three commits authored directly by Claude** (all from the Swarm migration
window, 10–17 Apr; mapped under `LLM <none>` in `.mailmap`):
- `8592fb3` — Docker Swarm migration plan doc.
- `749bec6` — Traefik v3.6 upgrade + router/service fix + DB indexes.
  `messages(flagged, pub_date DESC)` took PublicTimeline from ~11s to
  <1ms on 1.8M rows. **Load-bearing fix.**
- `299bb48` — `docs` explaining host-mode vs ingress-mesh trade-off.

All three carry `https://claude.ai/code/session_01APwSPLeWmkG4MBumCDVecc`
as a trailing reference — i.e. the sessions are linkable.

**12 commits with `Co-Authored-By: Claude` trailer**, all Leo's.
Distribution by task:
- Infra/deploy: `3c649da`, `8603cd0`, `1bdbb08`, `8705863`.
- DB migration: `c972607`, `f0a0b01`, `53bab02`.
- CI tooling: `bdb6c16` (gofmt + golangci-lint + hadolint),
  `1b99c35`, `e04ad0b`.
- Small Go refactor: `5722bef`.
- Docs: `e9ae5f6` (architecture + `progress.md`).

**Peter's terraform README** (`4a10373`) is co-authored with Claude in a
free-form trailer; the file itself begins with `Author - Claude`.

**Files that name Claude:** `CLAUDE.md` (project agent instructions),
`terraform/README.md`, `docs/incidents/session11-ops-debug.md`, `.mailmap`.

**Authorship totals (`git shortlog -sn`):** Leo 104, Peter J 81, Håkon 45,
Peter K 30, Apoorva 23, LLM 3.

### Where it helped (specifically)

- The DB index fix in `749bec6` is measurable and load-bearing.
- `session11-ops-debug.md` pairs AI-generated diagnostics (Swarm ports,
  ingress mesh, HTTP/2 over VXLAN MTU 1450) with my own domain pushback.
- Migration plans gave the team a shared artefact to argue about before
  doing the migration.
- Quality tooling install (`bdb6c16`) — pure boilerplate, AI did it in
  one go.

### Where it hurt or was wrong

- In the same incident doc I explicitly rejected a *previous* Claude
  session's cache-overflow theory after re-investigation. Hours saved by
  pushing back instead of accepting it.
- `CLAUDE.md` had Worker-1 and Worker-2 IPs swapped vs reality — AI-
  maintained context drifted from the actual cluster.
- `terraform/README.md` opens with `Author - Claude` and describes a setup
  that hasn't been verified against the actual built infrastructure.

### Honest trade-offs

- **Asymmetric use.** Most of the AI fingerprint is mine. That means most
  of the "did you understand what got committed?" questions at the exam
  will be mine too.
- **Speed vs ownership.** AI-generated migration plans were faster to
  produce than to read carefully.
- **Authorship hygiene.** We used three different conventions (Claude as
  git author, `Co-Authored-By` trailer, freeform "Co-author Claude" in
  subject). `.mailmap` cleanup is honest but late.

### Suggested ~200-word arc

Open with the tool (Claude Code) and the asymmetric usage fact. Then one
concrete win (`749bec6` index fix or the incident-doc partnership) and one
concrete miss (the cache-overflow theory I rejected, or the stale `CLAUDE.md`
worker IPs). Close on the trade-off you'd actually defend: speed vs.
ownership, plus the authorship-hygiene late `.mailmap` cleanup. Don't make
generic claims about productivity — every sentence should point at a SHA or
a file.

---

---

## Discord findings — 6 channels exported (655 messages total)

Sources: group DM "DevOps ITU" (121 msgs, Jan 30 → May 15), and the
**"ITU DevOps Groupwork" server**: `#generelt` 54, `#standup` 504,
`#links-and-ressources` 3, `#alerts` 0, `#report` 0. (`#keys` skipped to
avoid pulling plaintext secrets.) The DM was for scheduling; `#standup`
is where the real technical chat lived.

**Communication-tool migration over the semester (useful for DevOps Style):**
- *Jan–early Feb:* Discord DM only — scheduling + light tech coordination.
- *Feb 9:* Leo proposes the DevOps Groupwork server with `#generelt`,
  `#standup`, `#links-and-ressources`. DM dies after Feb 11.
- *Feb–Apr:* `#standup` is the primary channel (140 Feb / 249 Mar /
  106 Apr msgs). Real debugging, decision-making, async check-ins.
- *Late Apr–May:* `#standup` volume drops; coordination moves to PR comments
  + Leo's live-written incident doc (`session11-ops-debug.md`).
- *May 15 (today):* DM revives for report-handover coordination.

This *is* a positive maturation pattern: chat → server channels → PR
comments + incident docs as the audit trail. Worth mentioning in DevOps
Style.

**Author share in `#standup`:** Leo 175 (35%), Peter K 108 (21%),
Apoorva 82 (16%), Håkon 74 (15%), Peter J 65 (13%). Leo carries the chat
the same way he carries the commits — see the asymmetry caveat in the
GAI section.

---

### Big finding 1: The Go decision was Leo's solo call with Claude — then team-ratified by PR

- **DM, 2026-02-07:** the only written tech-stack discussion. Håkon: *"Im
  pretty sure we as a group have to agree on a tech stack"*. Peter K:
  *"I'm most familiar with Java, but I also think that staying with
  Python but changing the framework to Bottle with Jinja2 could be an
  option."* Apoorva, Peter J both nod toward Python+Bottle/Pyramid.
  **Nobody suggests Go.**
- **`#standup`, 2026-02-10, Leo:** *"i have created a pr with a full
  working go port. if you have time please review it and ask if you have
  questions. **i did it together with claude code to be fair, which made
  it easier to overcome some of the go coding syntax.** we tried to make
  it as modular and easy to maintain as possible splitting the main.go
  files into separate .go files with each of their own concerns."*
- **Same day, Peter K:** *"I think the go code looks good and the tests
  passes, can I merge the pr to refactor branch?"*

This is the project's single biggest technical decision. It was made over
a weekend by Leo with Claude Code, not by group debate. The team
ratified it via PR review. Worth saying out loud — both in **DevOps Style**
(decisions made unilaterally + ratified by code is a real working pattern)
and in **GAI** (Claude is in the foundation of the project).

### Big finding 2: The DO migration was group-decided and cost-driven, not "Apoorva already had droplets"

The `docs/operations/docker-swarm.md` line 115 explanation
("Why DigitalOcean instead of more Hetzner nodes? A colleague had already
provisioned the DO droplets") **oversimplifies what actually happened.**
The Discord shows it was an explicit group decision driven by student
credits and cost.

- **2026-03-09 Apoorva:** *"we have the digital ocean credits too"*
- **2026-03-09 Leo:** *"should we consider trying to migrate to
  digitalocean? maybe after getting the db setup done?"*
- **2026-03-10 Leo:** *"For me it would be ok to migrate to postgreSQL
  but keep it on the same VM in its own container so we wouldn't have to
  pay for another vm.. unless we change to digital ocean and use our
  student discounts — that would be optimal!"*
- **2026-03-11 Håkon:** *"Yes, i can agree on migrating to digital ocean."*
- **2026-03-11 Apoorva:** *"I also think switching to digital ocean makes
  the most sense!"*

For the **Operation** section: the honest answer to "why DO over Hetzner?"
is **student credits + free quota + the team having to pay out of pocket
for Hetzner**. We deliberated, we agreed, we picked the cheaper option.
That's a normal, defensible decision — say it.

### Big finding 3: Cost awareness was real and continuous

- **2026-03-02 Leo (re Hetzner volume):** *"I will have an eye on the
  server price. It is still below 1 dollar, so totally fine adding the
  hetzner-volume 🙂 I'll let you know if we go above let's say a couple
  of dollars in total."*
- **2026-03-02 Håkon:** *"Hetzner doesnt accept github student tokens
  right? So its something we have to pay for, but for 4 months its just
  €2"*
- **2026-03-03 Leo:** *"Heads up — if you're playing around with vagrant
  up and it creates new servers on Hetzner, just remember to shut them
  down afterwards if you're not using them, so we don't burn through
  credits unnecessarily."*
- **2026-03-10 Leo:** *"there are 2 servers at hetzner right now - is this
  on purpose?"*

For **Maintenance** (or DevOps Style): we shepherded cost out-of-band, not
via tooling. Today (May 15) Leo deleted the orphaned droplet — that's the
end-state of an informal habit, not an automated guardrail. Honest.

### Big finding 4: April 16–17 cluster outage debugged live in `#standup`, Claude wrote the post-mortem

This is the *richest* incident record in the whole project. Two firewall
bugs back-to-back:

**Apr 16 evening, Peter K:**
- DB-side firewall bug found and fixed: *"swarm is up and running now.
  The issue was that our nodes weren't whitelisted in the DO database
  firewall so the containers couldn't connect to postgres and kept
  crashing. Fixed by adding the 3 IPs to trusted sources on the DB."*
- DNS cutover attempted. Site dead. Roll back.
- *"Should we just see and solve it tomorrow?"* (Leo). *"Yeah I think so
  yes just change it back."*

**Apr 16 same evening, Peter J:**
- Adds the DO **cluster** firewall (the one that breaks Swarm overlay
  comms): *"I've created the firewall and Let's Encrypt might need to be
  restarted. Here is what claude wrote 🙂"*

**Apr 17 afternoon, Leo:**
- *"Im sitting outside (facing DR) in the sun if you feel like joining
  after lecture:) I had a bit of work to be done so werent able to join
  lecture. I have my laptop running so it is possible to try to fix our
  DO setup here in the sun hihi"*
- Iterates with Claude on the Traefik/overlay diagnosis.
- *"Okay! claude feels really sure about this fix. there is a pr to dev!"*
- *"message from claude!"* followed by a long, **byline-explicit** post:
  *"hey team, claude here — leo asked me to write up what we fixed today
  so everyone's on the same page. short version: the DO swarm cluster
  wasn't serving traffic properly. site stayed on hetzner the whole time
  so no users noticed, but we couldn't do the DNS flip until it was
  sorted. now it's sorted. three bugs, stacked on each other..."*
- The three-bug stack: Traefik 3.4/3.5 docker-API mismatch → 404;
  Traefik 3.6 stricter multi-service routers; webserver on two overlays
  → Traefik picks wrong one → 30s hangs. (Matches the Evolution research
  exactly.)

For **Operation:** this is the single most defensible incident at the
exam. Concrete root causes, real diagnostic process, day-after follow-up
documented. The honest "we kept Hetzner live so users didn't see this"
detail is gold.

For **GAI:** Leo posting a Claude-byline message to the team is unusual
and worth naming directly. It's not hidden, not laundered, not pretended-
human. Trade-off: it short-circuits team understanding — the rest of the
group reads Claude's post-mortem, not Leo's. Worth flagging.

### Big finding 5: April 29 – May 1 personal-timeline blow-up — monitoring caught it, fix took two passes

- **2026-04-29 Apoorva:** *"DigitalOcean monitoring triggered: CPU is
  running high — db-postgresql-fra1-53911 — hi guys, CPU activity of our
  DO database has spiked in the past 10 mins. Have there been any
  changes?"* Concrete proof that our DO-side monitoring actually worked
  and a teammate caught the alert first.
- **2026-04-29 Håkon:** *"We have an issue where we randomly get 504
  errors (Gateway Timeouts) on the devtroopers.com url."*
- **2026-04-29 Håkon:** *"I think i have fixed the problem with the
  Gateway Timeouts, but i need someone to review #152 for it to work"*
- **2026-05-01 Håkon:** *"I am not so sure anymore. I found out that a
  swarm worker was not responding, and after i reset it manually i havent
  got any 504s after"*

This is the iterative-debugging story. PR #152 didn't fix it; the actual
fix landed later in `a3dfc3d`. Honest version of the timeline. Pair with
the Evolution & Refactoring write-up of the LATERAL join.

### Big finding 6: AI usage is widespread, openly acknowledged, multiple team members

Not just Leo. Across `#standup`:
- **2026-02-10 Leo:** *"i did it together with claude code to be fair"*
  (the Go port).
- **2026-02-12 Leo:** asks an LLM about Docker on macOS vs Linux, shares
  the answer with the team.
- **2026-03-02 Håkon:** *"Claude talks about putting the .db on a
  'Hetzner Volume'"* — Håkon also uses Claude.
- **2026-03-10 Håkon:** *"The CI will still run on every push (according
  to claude lol)"* — light-touch but acknowledged.
- **2026-03-27 Leo:** pastes a long Claude analysis comparing repo state
  to the course-repo spec, asks the group *"Do you agree on what it
  lists?"*. Project management via Claude.
- **2026-03-27 Leo (also in `#generelt`):** Claude diagnoses the CD
  pipeline bug (missing `DATABASE_URL` write). Posted verbatim with
  attribution. Concrete win.
- **2026-04-10 Leo:** *"PR with docker swarm. there is a .md document
  explaining what claude code thinks is the best way to do it"*.
- **2026-04-16 Peter J:** *"Here is what claude wrote 🙂"* — Peter J uses
  Claude for the firewall plan.
- **2026-04-17 Leo:** posts Claude's post-mortem as Claude (the byline
  message above).
- **2026-05-01 Peter J:** posts Claude's recap of the Docker Scout / CVE
  / multi-stage fix as part of his standup.

Tools mentioned: **only Claude** (specifically "claude code", "claude opus
4.6"). Zero mentions of ChatGPT, Copilot, Cursor, GPT. Honest input for
GAI section.

**Asymmetric-use refined:** Leo uses Claude most visibly, Håkon and
Peter J both reference it openly, Peter K and Apoorva have no Discord
fingerprint of AI use. So roughly 3-of-5 acknowledge AI, but Leo dominates
the AI-driven artefacts. The exam-defence implication stands: most of the
"did you understand what got committed?" questions land on you.

---

### Smaller-but-useful detail nuggets

- **2026-02-09 Leo (setting up the server):** *"I hope it didn't come
  across as me trying to control things. Just wanted to make sure we were
  on the same page before I headed out."* Honest self-awareness about
  taking the initiative; usable in DevOps Style.
- **2026-02-14 Leo:** repo migrated from `leosakharoff/itu-minitwit` to
  `DevTroopers-ITU/itu-minitwit` org. Worth a one-liner in Maintenance.
- **2026-03-08 Leo:** *"should we instead focus on implementing the ORM
  (db independent way of writing sql) and then also having the database
  on its own vps? or will a hetzner-volume do the 'same' no matter if we
  migrate away from SQLite?"* — the pivot moment from "snapshot the
  SQLite file" to "migrate to a real DB". Quotable for Evolution.
- **2026-05-01 Apoorva:** posted **exam-prep TA notes** in `#generelt`
  listing what the examiner is likely to ask. Worth reading before the
  exam — covers peer review, alternatives to chosen services, how the
  team worked together, branching strategy, recovery from incidents,
  visualization. File is in the export.
- **`#generelt`, 2026-05-04 Apoorva:** ⚠️ a username/password pair posted
  in plaintext. Rotate if those credentials are still live anywhere.
  Honest "still ugly" data point — Maintenance section if you want a
  sharp one.
- **`#links-and-ressources` has only 3 messages.** The proposed
  knowledge-curation channel never took off — useful as a "what didn't
  work" detail in DevOps Style.
- **Overleaf is the final-report tool** (link posted 2026-02-09 in
  `#generelt`). The `report/` dir in the repo is the markdown source of
  truth that gets pasted into Overleaf — worth flagging if you mention
  documentation sprawl.

---

### Suggested edits to the sections based on Discord findings

- **Evolution and Refactoring:** swap one of the medium-strength issues
  for *"The Go decision was unilateral + AI-assisted + team-ratified by
  PR"* — it's the largest single tech pivot and you have receipts.
- **Operation:** rewrite the DO-vs-Hetzner sentence to mention student
  credits + cost. The current `docs/operations/docker-swarm.md`
  explanation is incomplete — say so.
- **Maintenance:** add a line about cost shepherded out-of-band (Hetzner
  cost watching, then the manual `rm` today).
- **DevOps Style:** include the "Leo proposes server + channels Feb 9"
  story as a concrete coordination-evolution example. Add the asymmetric-
  use observation here too.
- **Use of Generative AI:** ground the section in the Apr 17
  Claude-byline post-mortem as the canonical case study. It's
  *unusually* explicit AI use — and the honest trade-off (teammates
  read Claude's words, not yours) is the kind of self-criticism a grader
  rewards.

---

## Files to consult while writing

- `docs/incidents/session11-ops-debug.md` — richest single source
- `docs/operations/docker-swarm.md` — migration rationale + DO-vs-Hetzner reasoning
- `docs/architecture/latest-counter-in-db.md` — concrete state bug
- `docs/incidents/personal-timeline-perf.md` — timeline fix with EXPLAIN numbers
- `monitoring/prometheus/prometheus.rules.yml` — actual alert thresholds
- `docs/progress.md` — session-by-session timeline
- `.mailmap` — authorship + LLM entry
- `terraform/README.md` — verify it matches reality before referencing

## Hetzner snapshot (separate, not in repo)

`~/Dev/itu-minitwit-hetzner-snapshot/` has `.bash_history` (824 lines of
actual ops commands), `docs/incidents/`, container logs, monitoring volumes,
and migration logs from the decommissioned box. Useful for sourcing exact
phrasing if needed.
