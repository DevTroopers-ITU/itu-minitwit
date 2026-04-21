# `latest` Counter — From In-Process to Postgres

## Background

The simulator API carries a `?latest=N` query parameter on every write. The grader later calls `GET /latest` and expects to see the most recent `N` it sent. This is how the grader confirms a request was actually processed.

Originally, the webserver stored this counter as a package-level Go variable:

```go
var latest int = -1
```

That worked fine while we ran a single webserver process. Once we moved to Docker Swarm with `replicas: 3` spread across all three nodes, each replica ended up with its own copy of `latest`. Requests from the grader are load-balanced across replicas, so after writing `latest=5` on replica A, a subsequent `GET /latest` might land on replica B and return a stale value.

We did not observe this bug in production yet because the grader still points at the old single-server Hetzner deployment. The moment the simulator URL flips to `devtroopersminitwit.codes` (DO Swarm), the counter becomes inconsistent.

## What Changed

The counter now lives in the shared Postgres database that all three replicas already use. Every webserver instance reads and writes through the same row, so they all agree on the value.

Concretely:

- New GORM model `SimState` with a single row (`id = 1`, `latest int`), added to the auto-migration alongside `User`, `Message`, and `Follower`.
- Two new store methods:
  - `GetLatest()` returns the counter, or `-1` if the row does not exist yet (matching the old default).
  - `SetLatest(v int)` upserts the row.
- The `getLatest` handler and `updateLatest` helper in `sim_api.go` delegate to the store instead of touching package state.
- The package-level `var latest int = -1` is gone.
- Tests no longer reset `latest = -1` between cases — each test already spins up a fresh SQLite DB via `setupTestServer`, so the empty table naturally returns `-1`.

Total diff: six files, +47/−15 lines.

## Why This Matters

This aligns the counter with the stateless-webserver principle the rest of the architecture already follows: session data lives in client cookies, and user/message/follower data lives in Postgres. The in-process `latest` was the one leak of per-replica state, and scaling to three replicas exposed it.

## Trade-off We Accepted

`SetLatest` uses a plain GORM `Save` (upsert), which does not enforce that the counter only moves forward. In theory, two simultaneous writes could reach Postgres out of order and an older value could overwrite a newer one briefly.

In practice this does not bite us because the course grader is sequential — it waits for its own response before sending the next request — so the writes arrive in order. If we ever observe flakes under heavier parallel load, we can tighten `SetLatest` with a monotonic guard:

```sql
ON CONFLICT (id) DO UPDATE SET latest = EXCLUDED.latest
WHERE sim_states.latest < EXCLUDED.latest
```

Chose the simpler version for now because (a) it passes all our tests, (b) the grader's traffic pattern doesn't expose the race, and (c) it's easier to explain and reason about.

## One-Paragraph Defense

The `latest` counter used to be a variable in the webserver process. Once we started running three webserver replicas on the Swarm, each process had its own copy, so the grader could write the counter on one replica and read a stale value from another. We moved the counter into the shared Postgres database that all replicas already talk to, added two small functions — one to read, one to write — and updated the handlers to go through them. Now all three replicas agree on the value.
