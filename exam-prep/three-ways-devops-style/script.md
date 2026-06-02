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

FEEDBACK (~45s): three levels.
- Human loop = our STRENGTH: catch error → Discord → group swarms to fix together. Intuitive, not structured.
  [term: Swarm and Solve Problems to Build New Knowledge]
- Tooling loop = GAP: under-used monitoring/logging, didn't "see problems as they occur."
- LINK: weak feedback → batches stayed big (no signal pulling us smaller).

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

## Full script (spoken, ~3 min)

> **Opener.** "My main takeaway is that we learned, by ourselves, that **the tools
> aren't enough** — it was when we had to *operate* them that DevOps became real. That
> fits the course's point that **value only happens when the service is running in
> production.** And the honest pattern across all three Ways is that our practices came
> **reactively** — we'd have gained from being more **proactive**, putting structure in
> earlier instead of reaching for it once something hurt. So — Flow, Feedback, and
> Continual Learning."
>
> **Flow.** "Flow changed the most for us. Early on we piled work onto our `dev` branch
> in big batches before merging — long lead times, and we were bad at **reducing batch
> sizes.** The refactor is the clearest example: we rebuilt everything in Go until it
> worked, then made *one big merge.* But by the end you can see in our history we were
> merging in much smaller sizes. The honest pattern is that small batches came naturally
> in the **repair** phase — once production was running and we were debugging, we
> weren't afraid to merge — but in the **build** phase, new features, monitoring,
> logging, breaking things small was hard. We only learned this in the second half of
> the course. Going forward I'd apply it to *both* building and repairing."
>
> **Feedback.** "Feedback splits into three levels for us. The **human loop** was our
> strength — catching errors, posting in Discord, and the group **swarming to solve
> them together to build new knowledge.** But it was intuitive, never structured. Where
> we fell short was the **tooling loop** — we could've used monitoring and logging far
> more to **see problems as they occur** and push quality back to the source. And I
> think that's *connected to why our batches stayed big*: without that fast signal,
> nothing pulled us toward smaller, safer changes."
>
> **Learning.** "**Continual Learning and Experimentation** was real for us, but
> informal. The genuine learning happened when we sat down and talked things through —
> we built an honest, open environment where people shared what they'd figured out, and
> that was real. Where I'm more critical is the documentation: we wrote incident logs
> and a docs folder, but a lot of it was **AI-generated**, and a document isn't the same
> as learning. My own takeaway is that AI over-produces — so the real value was in the
> **conversations, not the write-ups** — and we did that reactively, rather than making
> it a habit. On the **experimentation** side, the Hetzner-to-Swarm migration is the
> telling story: it's where our development environment broke down — we never got the
> Swarm running locally to experiment with properly. The blue-green test on DigitalOcean
> before cutover was good, but with real local parity and everyone on board, we could've
> done it more sophisticatedly."
>
> **Close.** "So the through-line is **reactive versus proactive** — and that shift only
> really clicked once we were the ones *operating* the system. That, for me, is what
> DevOps actually is."

---

## Course terms each beat lands on (quick ref)

- **Frame:** *"value occurs when services are running in production"*
- **Flow:** Reduce Batch Sizes · Limit WIP / Make Work Visible · shorten lead time
- **Feedback:** Swarm and Solve Problems to Build New Knowledge · See Problems as They
  Occur · Keep Pushing Quality Closer to the Source
- **Learning:** Continual Learning *and* Experimentation · Institutionalize the
  Improvement of Daily Work · Transform Local Discoveries into Global Improvements
- ⚠️ Don't attribute CALMS / DORA / Westrum to the course (not taught).
