# Session 05 — What is DevOps + Configuration Management (+ simulator starts, Three Ways, DB abstraction)

Material: `../../course-material/sessions/session_05/` (Slides.md, README_TASKS.md, README_PREP.md = DevOps Handbook Part 1 + Ansible video).

## Erratum: continuous delivery vs deployment (resolves the S04 slide)
- S05 slide admits the S04 *Python for DevOps* definitions were WRONG; Helge "sends an erratum." Authoritative defs:
  - **IEEE 2675-2021:** continuous *deployment* = automated deploy to **production**; continuous *delivery* = frequent automated releases to **staging/test**.
  - **IBM:** delivery = always releasable (person may click); deployment = no manual step, tests pass → auto to prod.
- So: deployment = auto to prod; delivery = automated up to "ready." **We do continuous deployment.** (Use IEEE/IBM, mention the erratum.)

## Configuration Management
- The step AFTER provisioning a VM: install + configure software before the app runs. We did it with the Vagrant `provision "shell"` block (+ `terraform/bootstrap.sh`). Shell scripts (we didn't adopt Ansible — optional task).
- Tools: Ansible, Chef, Puppet, Salt. Do it repeatable + automatic + version-controlled.
- Provisioning (make the VM) ≠ Vagrant's `provision` step (= configuration management).

## Idempotency (key exam word)
- Run a script once or many times → same state, no errors/duplicates. Safe to re-run.
- `mkdir -p`, `rm -rf`, append-only-if `! grep -q`, "skip if DB exists". Config tools give it for free.

## IaC vs Configuration Management (answers "where's Terraform?")
- **Provisioning/IaC** = create the infra (VMs, network). Tools: Terraform (S12), Vagrant. "Make the box."
- **Config management** = set up software on the box. Tools: Ansible/shell (S5). "Configure the box."
- Course order: S3 provisioning → S5 config mgmt → S12 Terraform. Shared principle: idempotency. Terraform is declarative+idempotent by design; shell we make idempotent by hand.

## What is DevOps?
- No standard def. Mapping Study (Jabbari, 117 papers): "bridging the gap between Dev and Ops, emphasizing communication & collaboration, CI, QA & delivery with automated deployment." Handbook → the Three Ways.

## The value-stream diagram (Business → Dev → Ops → Customer)
- Left→right production line. **Ops = running software in production** (deploy, monitor, keep up, fix outages). Ops sits next to Customer because **value only happens when the app is live in prod** ("value occurs when services run in production").
- DevOps = Dev and Ops are the **same people** (wall gone). For us: Dev = Go code; Ops = swarm/Traefik/monitoring/CD/incident fixes; Customer = simulator + web users.
- 3 rows = 3 Ways: top arrow → (Flow, Dev→Ops→customer), middle arrow ← (Feedback back to Dev), bottom small loops (Learning everywhere).

## The Three Ways (DevOps Handbook) + principles
- **First Way — Flow** (→): Make Work Visible; Limit WIP; Reduce Batch Sizes; Reduce Handoffs; Find & fix Constraints; Eliminate Waste.
- **Second Way — Feedback** (←): See Problems as They Occur; Swarm & Solve; Push Quality to the Source; Optimize for Downstream.
- **Third Way — Learning**: Institutionalize Improvement of Daily Work; Local → Global improvements; Inject Resilience; Leaders Reinforce Learning Culture.

## PROJECT GROUNDING (the un-fluffy examples — for the DevOps Style slide)
- **Flow:** short lead time + small batches + no manual-deploy handoff. Concrete: a DB-URL env fix ships on its own, no waiting for a big release. (Handoff removed = no human SSH-deploy; CD does build→push→deploy.)
- **Feedback — TWO close incidents (don't conflate; both real, traced):**
  - **20 Apr `6e94660`:** a JOIN query took **41–49s** for one user → switched to a subquery. A LOUD problem (visible latency). [This is the "join change" Leo remembers.]
  - **21 Apr `b874e8e` (PR #138, branch `fix/latest-in-db`):** `latest` was a package-level int; with 3 replicas each had its own copy; simulator polls `GET /latest` and read a stale value ~2/3 of the time → fixed with a shared `SimState` Postgres row (+`ON CONFLICT WHERE latest < EXCLUDED.latest` guard). A SILENT problem (wrong value, no error).
  - The join/subquery fix did NOT fix the `/latest` failures — different root cause (the counter). Leo's memory merges the two; git separates them.
  - **Discovery mechanism (VERIFIED — good story):** caught by `scripts/test-do-swarm.sh` (`2e18cd2`, 20 Apr) — a simulator-style test run against the **DO swarm** (`64.226.116.162:8080`, tests `/latest`) **while Hetzner was still production**. Fix `b874e8e` next day (21 Apr); Hetzner decommissioned 4 May (PR #162). The bug is multi-replica-only → invisible on single-instance Hetzner. So it was found by **pre-production parity testing before the cutover**, not by prod monitoring. = "production-like test environments" (continuous delivery) + "Push Quality Closer to the Source." Strong Feedback AND Evolution material.
  - **Why our monitoring couldn't catch the counter bug (VERIFIED from alert rules + code, not guessed):** the `/latest` route returns HTTP **200 with a stale value** — no error, no log (the only log on that path fires only on a DB-write failure, absent in the in-memory version). Our 3 Prometheus alerts are `WebserverDown` (up==0), `HighErrorRate` (5xx>10%), `SlowResponses` (p95>1s) — a fast 200-with-wrong-data trips NONE. The 41–49s JOIN, by contrast, WOULD trip `SlowResponses`. Loud=catchable, silent=not.
  - **Can't re-read April logs:** Loki retention ~7 days (`reject_old_samples_max_age: 168h`). So claim what's verifiable (alert rules + code), flag what isn't (expired logs). Don't assert "no log events existed."
  - **GenAI angle (for GAI report section):** Claude helped diagnose the distributed-state bug our monitoring didn't flag, and caught that the proposed join-change wouldn't fix the real symptom. Honest + critical LLM use.
  - **Was it a real grader problem? (honest nuance — strong exam material):** Our `test-do-swarm.sh` `/latest` check (send `latest=N` → GET `/latest` → assert `==N`) mirrors the course's own contract test `test_latest()` exactly — so it's NOT a manufactured requirement. BUT the actual `minitwit_simulator.py` only SENDS `?latest=N`; it never GETs `/latest` and asserts. Only the contract test reads it back (single-instance, passes). We don't have the live grader's source, so we CAN'T prove the prod grader penalized the staleness. Likely low/zero grade impact → explains why other groups didn't report it (that + only triggers with in-memory counter + multiple replicas). Still: inconsistent `/latest` violates the documented contract, and per-replica mutable global state is wrong for a scaled service → DB-backed fix correct regardless. Distinguish "violates contract / real bug" from "affected the grade" (unprovable).
- **Learning — Traefik 504 incident (taught, verified):** `c3abcc2` (17 Apr), **PR #129** `fix/traefik-ingress-host-mode`. Bug: Swarm ingress mesh (IPVS+VXLAN) broke HTTP/2 → 504 timeouts; fix = host-mode port publishing. Learning artifact: **719-line incident log** `docs/incidents/session11-ops-debug.md` — timestamped play-by-play, running **"Lessons recorded"** list, **group-meeting crib sheet**, **pre-flight checklist**, post-fix verification. Plus a whole `docs/incidents/` + `docs/architecture/` tree (e.g. `latest-counter-in-db.md` with a "One-Paragraph Defense"). = Institutionalize Improvement + Local→Global. **Strongest Way.**
- Honest balance: strong at incident write-ups, but **docs drift** (CLAUDE.md/progress.md stale) — learning capture wasn't perfect.

## How to find incidents in git (exam skill)
```
git log -i --grep="fix" --oneline          # reactive fixes
git show <hash>                            # symptom+cause+fix (good messages tell the story)
git log --merges --ancestry-path --reverse <hash>..master | head -1   # the PR that carried it to master
```

## Tasks (coursework)
1. Simulator starts. 2. **Map yourselves to the Three Ways** (= the DevOps Style slide exercise). 3. DB abstraction layer / ORM, no raw SQL (we did GORM). 4. Idempotent config scripts. 5. Maintenance: fix issues ASAP, add missing features, weekly releases.

Also: ephemeral containers (don't store data in the container — use volumes); feature flags to remove sim code after the simulation.
