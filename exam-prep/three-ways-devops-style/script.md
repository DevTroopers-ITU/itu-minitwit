# Oral Exam — DevOps Style (The Three Ways)

**Speaker:** Leo · **Time budget:** ~3 minutes · **Slides:** 1

Course anchor: *The DevOps Handbook*, Part 1 — the **Three Ways** (Session 05,
"What is DevOps and configuration management"). DevOps is framed as bridging Dev
and Ops through communication, CI/CD, quality assurance and automated deployment.

---

## Slide (use ONE)

Reuse `report/exam-storyboard.drawio.png` and overlay three labels with a
one-word verdict each:

| Way | Label | Verdict |
|-----|-------|---------|
| First Way  | **Flow**     | Weak (improved over time) |
| Second Way | **Feedback** | Mixed |
| Third Way  | **Continual Learning** | Strong, but informal |

The storyboard already shows the on-time vs >2-weeks-late lanes and the
operational incidents, so the batch-size arc and the 16 Apr outage are visible
while you talk.

---

## Full script (~400 words, ~3 min at a calm pace)

> Our reflection uses the **DevOps Handbook's Three Ways** as a lens. The
> headline: for us, DevOps wasn't about adding tools — it became real once we
> had to *operate* the system ourselves.
>
> **First Way — Flow.** This was our weakest, but it's also our clearest
> evolution story. Early on, work piled up on a long-lived `dev` branch — the
> Python-to-Go refactor accumulated for *two weeks* before reaching production.
> Big batches, long lead times, no WIP limit. By the end, we'd flipped: security
> hardening shipped as **thirteen small PRs**, and bug fixes were one-to-three-line
> changes straight to prod. So we drifted toward the First Way — but by instinct,
> not by adopting a board or WIP limits.
>
> **Second Way — Feedback.** Three levels. Our *human* loop was strong — we'd
> spot something, drop it in Discord, and the team swarmed. But it was intuitive,
> not structured. Our *CI* gates caught real errors before prod — but we stopped
> at defending quality, never iterating on the tests to *improve* it. And our
> *monitoring* — Prometheus, Loki, Grafana — was installed but not wired into how
> we worked. The proof: on April 16th our Swarm overlay silently failed, and edge
> uptime stayed green for **eighteen hours**. The honest lesson: we scaled *tools*
> faster than the *practices* around them.
>
> **Third Way — Continual learning.** Our strongest, but informal. We kept a
> `docs/` folder and wrote up incidents — but it grew *reactively*, out of a
> single weekly meeting that couldn't spread one person's deep-dive to the team.
> Our best experiment was the migration: we ran the new Swarm cluster in parallel
> and tested it with the simulator before cutover — blue-green in spirit. But we
> de-risked the migration without ever *measuring* whether Swarm beat the single
> server. Experimentation without a hypothesis.
>
> If I tie it together: the thread under all three is **dev/prod parity**. When
> we moved to Swarm, local stayed on Compose while prod moved to a Swarm stack —
> so flow slowed, feedback moved downstream, and experimentation got harder. One
> missing capability quietly weakened all three Ways.
>
> So the real takeaway: the Three Ways aren't a checklist of tools. They're
> connected, and they only came alive when we owned operations.

---

## Cue-card version (glance, don't read)

- **Frame:** Three Ways as lens. *"DevOps became real when we operated it."*
- **Flow (weak → better):** 2-week refactor batch on `dev` → ended with **13 small
  hardening PRs** + 1–3 line fixes. Improved by instinct, no WIP limits.
- **Feedback (mixed):** human swarm = strong (Discord); CI = caught errors but only
  *defended* quality; monitoring = installed, not used. **16 Apr = 18h blind.**
  *"Scaled tools faster than practices."*
- **Learning (strong, informal):** `docs/` was reactive; one meeting/week =
  bottleneck. Migration = blue-green-ish experiment, but **never measured** Swarm vs
  single server.
- **Connective insight:** **dev/prod parity** broke at Swarm migration → weakened all
  three Ways at once. (Also a *12-Factor App* concept.)
- **Close:** Three Ways are connected, not a checklist; came alive with ownership.

---

## Likely follow-up questions (prep)

- *"How would you improve Flow?"* → WIP limits + a board; smaller batches by default;
  a `make run` one-liner and maintained dev/prod parity so shipping is cheap.
- *"What's a blameless post-mortem?"* → Third Way: write up incidents focusing on
  systemic causes, not blame, and turn local fixes into global improvements. We wrote
  incidents but never made it a *scheduled* practice.
- *"Where do DORA metrics fit?"* → They measure the Ways: deployment frequency + lead
  time (Flow), MTTR + change-failure rate (Feedback). We never tracked them.
- *"Best evidence the feedback loop was incomplete?"* → 16 Apr overlay outage: edge
  uptime green 18h while cluster redundancy was gone — control-plane vs data-plane.
- *"Idempotent provisioning?"* → Session 05 config-management point; our Terraform +
  UFW rules provisioned nodes reproducibly.
