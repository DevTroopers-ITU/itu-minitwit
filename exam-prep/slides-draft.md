# Slide draft — Leo's 3 slides (3 MIN TOTAL → ~60s each)

Your slides: 1) Evolution & Refactoring  2) DevOps Style (Three Ways)  3) AI Use & Maintenance.
Slides are lean anchors — depth goes in the Q&A (see [presentation.md](presentation.md) + [findings.md](findings.md)).
The "Say" lines are how you'd actually talk — plain, first-person. Read them aloud and make the words yours.

---

## SLIDE 1 — Reflection: Evolution & Refactoring  (~60s)
**Title:** We evolved it — we didn't rewrite it

- 2012 Flask app, taken over and changed piece by piece
- Python 2→3 → **Go** · SQLite → **Postgres** · one VM → **3-node Swarm**
- Honest bit: a migration left our **session key** behind → prod ran on a hardcoded one (found it prepping for this exam)

*Say: "We took over this old 2012 Flask app and changed it bit by bit — we didn't rewrite it. We moved to Go, swapped SQLite for Postgres, and went from one server to a 3-node Swarm. The honest part: somewhere in the Swarm migration we left our session key behind, so production ended up on a hardcoded one. We only caught that prepping for this exam. The lesson: when you migrate, double-check the config moved with it."*

---

## SLIDE 2 — Reflection: DevOps Style (Three Ways)  (~60s)
**Title:** Three Ways — an honest scoreboard

- **Flow:** auto-deploy on merge, small changes — *but infra isn't really reproducible (by hand)*
- **Feedback:** monitoring, logging, alerts; caught a bug testing the Swarm before we switched — *but tests use SQLite, prod uses Postgres*
- **Learning:** wrote incidents up in detail — *but some docs went stale*
- Verdict: strongest on **Flow + Feedback**, weakest on **reproducibility**

*Say: "We graded ourselves on the Three Ways. Flow — we deploy automatically on merge, in small changes; the weak spot is our infrastructure isn't really reproducible, we set it up by hand. Feedback — we've got monitoring, logging and alerts, and we caught a bug by testing the Swarm before we switched to it; but our tests use SQLite while prod is Postgres. Learning — we wrote our incidents up in detail, though some docs went stale. Overall: best at Flow and Feedback, weakest on reproducibility."*

---

## SLIDE 3 — Reflection: AI Use & Maintenance  (~60s)
**Title:** Maintaining it — and using AI honestly

- **Maintenance:** once the simulator started → fix issues fast, release regularly (e.g. the replica `latest` bug, a Swarm 504) — *honest: releases were manual / after-the-fact*
- **AI use:** used it to debug, understand the code, draft docs; credited it as **co-author** where it helped (course rule)
- Learned: AI over-produces — the real work was cutting it down to readable lessons, and only keeping what we understood

*Say: "Once the simulator started we were basically in maintenance mode — fixing problems as they showed up and releasing regularly; honestly our releases were a bit manual rather than fully automated. On AI — we used it a lot for debugging, understanding the code, and drafting docs, and we credited it as co-author where it helped, like the Terraform docs and a security fix. What we learned is it produces way too much text, so the real work was cutting it down to a few readable lessons — and only committing what we actually understood."*

---

## Hold in reserve for Q&A (NOT on the slides)
- Full secret-key story + git trail: `git log -L :getSecretKey:main.go`
- The `latest`-counter bug: loud-vs-silent, "can't prove the grader cared"
- Traefik 504 → the 719-line incident write-up
- Terraform: written after the fact, never run, doesn't match prod

## If you're over 3 min
- Drop the secret-key bullet from Slide 1; save it for Q&A.
- Say each Way in one line on Slide 2, skip the gaps, let them ask.
