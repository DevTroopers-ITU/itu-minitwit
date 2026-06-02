# Session 01 — Project start, SSH/SCP, Bash; taking over legacy MiniTwit

Material: `../../course-material/sessions/session_01/` (Slides.md, README_TASKS.md, README_PREP.md)

## Core concepts (from the material)
- The course's premise: take over a **2012 legacy system** (Python 2, Ubuntu 12.04, no git) and evolve/maintain it. Tools (SSH, SCP, bash) are means, not the point.
- **Maintenance = 40–80% of software cost** (Glass, "Frequently Forgotten Fundamental Facts", slide). M5: the dominant maintenance activity is *"understanding the existing product"* (~30% of time).
- Task discipline: first commit = code **as taken over, unmodified** (baseline to diff against). Then migrate Python 2→3 **changing as few lines as possible**, guided by the **existing test suite** (`minitwit_tests.py`). "Using existing test suites during refactoring is one of the most important tasks in software evolution." (README_TASKS.md)
- Exam is **pass/fail**; you fail by being unable to explain what you did (slide).

## What WE did (verified against git)
- `ae39e23` (30 Jan) — original taken-over code committed untouched (minitwit.py, flag_tool, control.sh, schema.sql, templates).
- `63f45f3` "think i got minitwit.py working now", `3ab1c8e` "Refactoring for week 1 exercises" — Python 2→3 migration (2to3, utf-8 decode fix, werkzeug import move).
- Next week: full rewrite to **Go + Gorilla Mux** (session 02).

## Key Q&A — "Why migrate to Py3 if we deleted it for Go a week later?"
Not wasted. The migration was the **understanding + verification** step:
1. Can't faithfully rewrite what you don't understand (Glass M5). Minimal-change migration = read the system without redesigning it.
2. Passing `minitwit_tests.py` gave a **verified behavioral spec / oracle** to port the Go version against → session 02's "one-to-one feature parity, not a big rewrite."

## Go rationale — defensible version (examiner will ask "would you choose Go again?")
- Honest motive: wanted to learn a new language (course rewards learning).
- Technical fit to pair with it: single static binary → small multi-stage Docker images, fast cold start (matters under simulator load), good concurrency + HTTP stdlib, Gorilla Mux course-recommended.
- Reflection gold: porting for feature parity forced reasoning at the **behavior/architecture** level, not syntax — the "reading others' code" skill (Spinellis reading).

## Test evolution — verified from git (be 100% sure)
Chronological, with commit receipts:
1. `ae39e23` (30 Jan, S01) — handed-out code had **one** test file: `minitwit_tests.py`, Flask `test_client()` suite ("(c) 2010 Armin Ronacher"). No sim test, no browser test.
2. `3ab1c8e` + `d0e26c6` (S01) — Py2→3 migration edited it: changes are essentially `'x' in rv.data` → `b'x' in rv.data` (Flask `.data` is bytes in Py3). Minimal-change "make existing tests pass."
3. `a2ab6fc` "Make master match refactor" — `minitwit_tests.py` renamed **byte-identical (R100)** into `python-references/` = parked as reference, **never run against Go**. Same commit adds `main_test.go` (Go httptest integration tests that replace it). Single-parent commit, not a PR merge.
4. `9bbf1f1` (S03) — `minitwit_sim_api_test.py` added. **Course-provided** (identical to `course-material/sessions/session_03/API_Spec/minitwit_sim_api_test.py`). We later changed it: port 5000→8080 (`a689204`), removed a delete case (`71b013e`), CI plumbing (`03109d3`,`7152670`,`d4fee21`).
5. `cc724e5` — `sim_api_test.go`, Go-native API tests.
6. `3d5ce5b` (PR #168) — `minitwit_ui_browser_test.py`, **Playwright** browser E2E.

CI (ci.yml) runs all three layers: `go test ./...`, `pytest minitwit_sim_api_test.py`, `pytest minitwit_ui_browser_test.py`.

### The exam point
- **White-box tests don't survive a language port.** Flask `test_client()` is bound to the app object → rewritten as `main_test.go`. Original kept only as a spec to read.
- **Black-box HTTP contract tests do survive** and double as the simulator contract (`minitwit_sim_api_test.py`). The simulator is itself a black-box HTTP client.
- **Trap to avoid:** the original Flask `minitwit_tests.py` (S01, handed out, white-box) is NOT the same as `minitwit_sim_api_test.py` (S03, course-provided, black-box). Different files, different sessions, and the latter was modified by us.

### Why ALSO a Go sim test (`sim_api_test.go`) when the Python contract test exists?
Test pyramid — same endpoints, different levels:
- Go tests reuse `setupTestServer` → in-process `httptest` + fresh temp SQLite, no Postgres/Docker/network → run in `go test ./...` in seconds. Fast feedback before image build (First Way / flow).
- Hermetic + deterministic per-test DB; the Python E2E shares a live DB (fragility: `71b013e` removed a delete case due to state coupling).
- White-box cases the contract test skips: exact codes (204/200), JSON shape, `TestSimAuthRequired` (401).
- Python `minitwit_sim_api_test.py` stays as the authoritative **contract / acceptance** test against the deployed artifact.
- Trade-off (examiner bait): yes it overlaps, but the two guard different failure modes — logic regressions (Go, cheap/early) vs deployment+contract correctness (Python, realistic).

## Open gaps / to verify
- `progress.md` S07 says "no Selenium/Playwright browser UI test", but `python-references/minitwit_ui_browser_test.py` exists. Verify: is it real, does it run in CI?
- (Optional) Confirm `main_test.go` actually covers the same cases as the original `minitwit_tests.py` (parity of coverage, not just surface).
