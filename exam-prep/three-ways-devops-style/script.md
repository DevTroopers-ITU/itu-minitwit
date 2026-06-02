# Oral Exam — DevOps Style (The Three Ways)

**Speaker:** Leo · **Time budget:** ~3 minutes · **Slides:** 1 (the "Reflection:
DevOps Style" slide — value-stream diagram on the left, project storyboard on the right)

> This is the version that came out of *thinking it through* — the framing and the
> honest judgements are mine. The examples are git-verified (see
> `../sessions/session-05.md` and `../presentation.md` for the commit trail).

Course anchor: **Session 05 / *The DevOps Handbook* Part 1** — the **Three Ways**.
DevOps = bridging Dev and Ops through communication, CI/CD, QA and automated
deployment (Jabbari mapping study; Handbook → the Three Ways).

---

## How to talk from the slide (left → right pointing path)

1. **Left diagram (~20s):** Business→Dev→Ops→Customer. "DevOps means Dev and Ops are
   the *same people*; value only happens when it runs in prod. These three rows are the
   Three Ways — top arrow Flow, middle arrow Feedback, bottom loops Learning."
2. **Flow** → point at the **DELAYED >2 WEEKS** lane (PR #65 big batch) vs the late small PRs.
3. **Feedback** → point at **SILENT CD FAIL / the latest-counter** (PR #138).
4. **Learning** → point at the **incident** row (Traefik 504, PR #129).
5. **Close** → the **Hetzner + DO parallel-run bar** = dev/prod parity, the thread under all three.

One slide, one path, ~3 min. Don't add a second slide.

---

## Full script (~420 words, ~3 min at a calm pace)

> We graded ourselves on the **Three Ways**, and the honest headline is: for us
> DevOps wasn't about *adding tools* — it became real once we had to *operate* the
> system ourselves.
>
> **First Way — Flow.** This is our clearest evolution story. Early on, work piled
> up on a long-lived `dev` branch — the Python-to-Go refactor accumulated for about
> *two weeks* before reaching production. Big batches, long lead times, no WIP limit.
> By the end we'd flipped: container hardening shipped as **thirteen small PRs**, and
> bug fixes were one-to-three-line changes straight to prod, no manual-deploy handoff —
> CD does build, push, deploy. So we drifted toward the First Way, but by instinct,
> not by adopting a board or WIP limits.
>
> **Second Way — Feedback.** Our *human* loop was genuinely strong — we'd spot
> something, drop it in Discord, and the team swarmed on it. Our *tooling* loop is
> where we fell short, and our clearest example is a *silent* bug. Our `latest`
> counter was an in-memory variable; with three replicas, each had its own copy, so
> the simulator read a stale value about two-thirds of the time. It returned HTTP 200
> with wrong data — so *none* of our alerts fired; up/down, error-rate and latency all
> looked fine. What caught it wasn't production monitoring — it was a **production-like
> test we ran against the new Swarm before cutover**, while the old single server was
> still live. That's the real lesson: a prod-like environment finds what dashboards
> can't, and you push quality back to the source. The contrast proves it — a separate
> 40-second query bug was *loud* and *would* have tripped our latency alert. So the
> honest verdict: we scaled *tools* faster than the *practices* around them.
>
> **Third Way — Continual learning.** This was our strongest, but informal. When our
> Swarm ingress started throwing 504s, we didn't just patch it — we wrote a
> **seven-hundred-line incident log** with a timestamped play-by-play, a "lessons
> recorded" list, and a pre-flight checklist for the cutover. That's institutionalising
> improvement. The honest gap: it grew *reactively*, out of one meeting a week, and
> some of our docs drifted out of date. And our one real *experiment* — running the
> new Swarm in parallel before cutover — we treated as a safe migration, but we never
> framed a hypothesis or measured whether Swarm actually beat the single server.
> Experimentation without a hypothesis.
>
> If I tie it together: the thread under all three is **dev/prod parity**. That silent
> bug was *invisible* on the single-server setup and only showed up with replicas — so
> it was the prod-like environment, not our dashboards, that surfaced it. Parity is what
> made the feedback possible. So the Three Ways aren't a checklist of tools — they're
> connected, and they came alive when we owned operations.

---

## Cue-card version (glance, don't read)

- **Frame:** Three Ways as an honest scoreboard. *"DevOps became real when we operated it."*
- **Flow (weak → better):** 2-week refactor batch on `dev` → ended with **13 small
  hardening PRs** + 1–3 line fixes; no manual-deploy handoff. Improved by instinct, no WIP limits.
- **Feedback (tools ≠ loop):** the **silent `/latest` counter bug** — in-memory, per-replica,
  stale ~2/3, HTTP 200 so **no alert fired**. Caught by **pre-prod parity test against the
  Swarm before cutover** (`test-do-swarm.sh`), not by monitoring. Contrast: a **40s query**
  was *loud* → would trip latency alert. *"Scaled tools faster than practices."*
- **Learning (strongest, informal):** Traefik **504** → **719-line incident log**
  (`docs/incidents/session11-ops-debug.md`) w/ lessons + pre-flight checklist. Gap: reactive,
  one meeting/week, **docs drift**.
- **Connective insight:** **dev/prod parity** sits under all three — the silent bug was
  invisible on single-server, only showed with replicas; parity is what *enabled* feedback.
- **Close:** Three Ways are connected, not a checklist; came alive with ownership.

**Held in reserve (drop into Q&A, not the 3-min talk):**
- *Three levels of feedback:* human/Discord swarm (strong) · CI gates (defended quality but
  never *iterated* the tests to improve it) · monitoring (installed, not wired into how we worked).
- *Why batches were big early:* we couldn't slice the Python→Go refactor small because there
  was **no production yet** — nothing to ship incrementally toward.
- *Learning is handover, not writing:* the `docs/` existed, but the real step is *handing
  knowledge over* — and AI-written docs over-produced, so the work was distilling them.
- *Other parity angle:* tests ran on SQLite, prod on Postgres — a green test didn't fully
  prove prod behaviour either.
- *Another feedback-gap example:* the 16 Apr overlay outage — edge uptime green ~18h while
  cluster redundancy had silently failed (control-plane vs data-plane).

---

## Likely follow-up questions (prep)

- *"How would you improve Flow?"* → WIP limits + a board; keep batches small by default; a
  `make run` one-liner + maintained parity so shipping stays cheap.
- *"Why didn't monitoring catch the `/latest` bug?"* → It returned **200 with stale data**;
  our 3 alerts are WebserverDown, error-rate >10% 5xx, and p95 >1s — a fast correct-looking
  200 trips none. Silent ≠ catchable; loud (the 40s query) is.
- *"Did the `/latest` bug actually cost grade?"* → Can't prove it. The contract test reads
  `/latest` back (single-instance, passed); the live simulator only *sends* `?latest=N`. So
  it **violated the documented contract / was a real bug**, but grade impact is unprovable —
  keep those two claims separate.
- *"Continuous delivery vs deployment?"* → We auto-deploy to prod on merge = **continuous
  deployment** (common def). The term is contested (the slide def is inverted) — describe the
  behaviour, then attach the label.
- *"What's a blameless post-mortem?"* → Third Way: write incidents up focusing on systemic
  cause, turn local fixes into global improvements. We did the write-ups (e.g. the 504 log)
  but never made it a *scheduled* practice.
- *"DORA metrics?"* → They measure the Ways: deployment frequency + lead time (Flow),
  MTTR + change-failure rate (Feedback). We never tracked them.
- *"Idempotent provisioning / where's Terraform?"* → Session 05 config-mgmt: IaC
  (Terraform/Vagrant) "makes the box", config mgmt "configures the box"; shared principle is
  idempotency. Honest gap: infra partly provisioned by hand, so not truly reproducible.
