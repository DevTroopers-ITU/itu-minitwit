# Course Terminology — grounding sheet (Three Ways / DevOps Style)

Verbatim course wording (Sessions 04 & 05, *DevOps Handbook* Part 1), mapped to our
project and to your own phrasing. Goal: when you talk, land each beat on the course's
*exact* term. Sources cross-checked against the public `itu-devops/lecture_notes`
and your `exam-prep/sessions/session-04.md` / `session-05.md`.

---

## The exact names (use these, not paraphrases)

- **First Way — Flow**
- **Second Way — Feedback**
- **Third Way — Continual Learning *and Experimentation*** ← keep "Experimentation"

## Verbatim principle bullets (drop the exact phrase, then your evidence)

**First Way — Flow:**
Make Work Visible · Limit Work in Progress · Reduce Batch Sizes · Reduce the Number of
Handoffs · Continually Identify and Evaluate Constraints · Eliminate Hardships and Waste
in the Value Stream

**Second Way — Feedback:**
See Problems as They Occur · Swarm and Solve Problems to Build New Knowledge · Keep
Pushing Quality Closer to the Source · Enable Optimizing for Downstream Work Centers

**Third Way — Continual Learning and Experimentation:**
Institutionalize the Improvement of Daily Work · Transform Local Discoveries into Global
Improvements · Inject Resilience Patterns into Our Daily Work · Leaders Reinforce a
Learning Culture

## Quotable lines (verbatim)

- DevOps def (Jabbari): *"bridging the gap between Development (Dev) and Operations,
  emphasizing communication and collaboration, continuous integration, quality assurance
  and delivery with automated deployment."*
- *"Value occurs for customers when services are running in production."*
- *"DevOps focuses on shortening deployment lead time to a period of minutes."*
- *"work is usually invisible"* (the value-stream problem)
- Idempotency: *"the state of the provisioned machine is always the same after running
  the provisioner script — no matter if it was executed one or multiple times."*
- CI/CD via **IEEE 2675-2021**; *"deploy frequently without any manual intervention."*

---

## Course term → our evidence → your words

| Course term (exact) | Our project evidence | Say it like you |
|---|---|---|
| **Reduce Batch Sizes / Limit WIP** | 2-week Go refactor on `dev` → later 13 small hardening PRs + 1–3 line fixes | "we gathered everything in dev until it looked almost done — we had no language for keeping batches small" |
| **Make Work Visible** / *"work is usually invisible"* | no board, no WIP limit; work only visible in the PR queue | "our work was basically invisible until it landed" |
| **Reduce the Number of Handoffs** | CD does build→push→`docker stack deploy`; no human SSH-deploy | "the manual-deploy handoff is gone" |
| *shortening lead time to minutes* / *deploy without manual intervention* | merge → auto-deploy to Swarm | "we do continuous deployment — merge and it's live" |
| **See Problems as They Occur** | the *silent* `/latest` bug / 16 Apr 18h-blind — we mostly *didn't* | "the loud problems we caught; the silent ones slipped past" |
| **Swarm and Solve Problems to Build New Knowledge** | Discord swarming on incidents | "someone drops it in Discord and we all jump on it" |
| **Keep Pushing Quality Closer to the Source** | CI gates; pre-prod parity test that caught the silent bug | "we scaled the tools faster than the practices around them" |
| **Institutionalize the Improvement of Daily Work** | incident write-ups, `docs/` tree — but reactive, one meeting/week | "we wrote it down, but didn't make improving a habit" |
| **Transform Local Discoveries into Global Improvements** | 719-line Traefik-504 incident log w/ lessons + checklist | "one person's deep-dive had to reach the whole team" |
| **Inject Resilience Patterns into Our Daily Work** | rolling updates `start-first`, health checks, parallel-run cutover | "we built in some resilience — rolling updates, health checks" |
| **Experimentation** (Third Way) | parallel Hetzner+Swarm run, but never measured Swarm vs single server | "experimentation without a hypothesis" |

---

## Caveats (so you don't over-claim)

- **CALMS, DORA, Westrum are NOT in the course material.** Don't attribute them to the
  course. You may mention DORA as outside reading — but what was *taught* is the Three Ways.
- The **"CI/CD terms are contested"** point (slide def vs Handbook def) is from the *live
  lecture / your notes*, not the public slides. Safe to use ("the lecturer noted the labels
  are used inconsistently — so describe the behaviour, then attach a label"), just know
  there's no public doc to quote.
- The **Business→Dev→Ops→Customer** value-stream picture and the 3-rows-=-3-Ways reading
  are an accurate interpretation of the Handbook image on the slide, not a verbatim quote.
