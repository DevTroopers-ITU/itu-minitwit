# Oral Exam — DevOps Style (The Three Ways)

**Speaker:** Leo · **Time budget:** ~3 minutes · **Slide:** 1 (the three-ways
value-stream diagram — Flow / Feedback / Continual Learning & Experimentation)

> Built from the exam-prep conversation — these are *my* words, grounded in the
> course's exact terms. Course-term backup: `course-terminology.md` (same folder).
> Through-line: **reactive → proactive**. Thesis: **the tools aren't enough; DevOps
> got real when we had to operate it.**

---

## Speaker notes (paste into the slide's notes box)

```
FRAME (~20s): Tools aren't enough — DevOps got real when we had to OPERATE it.
"Value occurs when services are running in production."
Pattern across all 3 Ways = REACTIVE; we'd have gained from being PROACTIVE.

FLOW (~45s): the Way that changed most.
- Early: big batches piled on dev → long lead time. Refactor = rebuilt in Go, then ONE big merge.
- End: git history shows much smaller merges.
- KEY INSIGHT: small batches felt natural in the REPAIR phase (prod running, not afraid to merge),
  but hard in the BUILD phase (features, monitoring, logging).
- Only learned it 2nd half of course. Forward: apply to both.  [term: Reduce Batch Sizes]

FEEDBACK (~50s): three levels.
- 1) Human loop = STRENGTH: catch → Discord → swarm together. Intuitive but worked.
  [term: Swarm and Solve Problems to Build New Knowledge]
- 2) CI gates: caught errors before prod, but only CAUGHT — never improved the tests.
  [term: Keep Pushing Quality Closer to the Source]
- 3) Monitoring/logging: had it, barely used it to "see problems as they occur."
- END: feedback worked when a PERSON noticed, not yet through the tools.

LEARNING (~45s): Continual Learning AND Experimentation — real but informal.
- WIN: the real learning was in CONVERSATIONS — honest, open environment, shared what we figured out.
- HONEST: docs were largely AI-generated; a document ≠ learning. Value was in the talk, not the write-ups.
  "We captured knowledge more than we circulated it."
- EXPERIMENTATION: Hetzner → DO Swarm migration = where our dev env broke; never ran Swarm locally.
  Blue-green test on DigitalOcean before cutover = good; but no local parity, not everyone on board.

CLOSE (~15s): Through-line = REACTIVE → PROACTIVE.
It only clicked once WE were operating the system. That, to me, is what DevOps actually is.
```

Optional on-slide labels (one per diagram row):
- *First Way:* batches — big → small (repair vs build)
- *Second Way:* human swarm strong · tooling under-used
- *Third Way:* real in talk, not AI docs · no local Swarm

---

## Full script (spoken, ~2:40 — trimmed to stay under 3 min)

> **Opener.** "My main takeaway is that we learned, by ourselves, that **the tools
> aren't enough** — DevOps became real when we had to *operate* them. That fits the
> course's idea that **value only happens when the service is running in production.**
> And the honest pattern across all three Ways is that we were **reactive** — we'd have
> gained a lot from being **proactive.** So — Flow, Feedback, Learning."
>
> **Flow.** "Flow changed the most for us. Early on we piled work onto our `dev` branch
> in big batches — long lead times, bad at **reducing batch sizes.** The refactor is the
> clearest example: we rebuilt everything in Go, then made *one big merge.* By the end,
> our history shows much smaller merges. The honest pattern: small batches came naturally
> in the **repair** phase — production was running, we weren't afraid to merge — but in
> the **build** phase, features, monitoring, logging, breaking things small was hard.
> Going forward I'd apply it to both."
>
> **Feedback.** "Feedback has three levels for us. First, the **human loop** — our
> strength: we'd catch something, post it in Discord, and the group swarmed to solve it
> together. Intuitive, but it worked. Second, our **CI gates** — they caught real errors
> before production, but we stopped at *catching* them; we never improved the tests. And
> third, **monitoring and logging** — we had them, but barely used them to actually **see
> problems as they occur.** So in practice, our feedback worked when a *person* noticed
> something, not yet through the tools."
>
> **Learning.** "**Continual Learning and Experimentation** was real but informal. The
> genuine learning happened when we sat down and **talked things through** — an honest,
> open environment where people shared what they'd figured out. But the documentation was
> largely **AI-generated**, and a document isn't learning — so we **captured knowledge
> more than we circulated it.** On experimentation, the move from Hetzner to the
> DigitalOcean Swarm is telling — it's where our dev environment broke down: we never ran
> the Swarm locally to experiment with. The blue-green test before cutover was good, but
> without local parity we couldn't do it properly."
>
> **Close.** "So the through-line is **reactive versus proactive** — and that only
> clicked once we were the ones *operating* the system. That, for me, is what DevOps
> actually is."

---

## Course terms each beat lands on (quick ref)

- **Frame:** *"value occurs when services are running in production"*
- **Flow:** Reduce Batch Sizes · Limit WIP / Make Work Visible · shorten lead time
- **Feedback:** Swarm and Solve Problems to Build New Knowledge · See Problems as They
  Occur · Keep Pushing Quality Closer to the Source
- **Learning:** Continual Learning *and* Experimentation · Institutionalize the
  Improvement of Daily Work · Transform Local Discoveries into Global Improvements
- ⚠️ Don't attribute CALMS / DORA / Westrum to the course (not taught).
