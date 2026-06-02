# Session 06 — Monitoring (metrics; Prometheus + Grafana)

Material: `../../course-material/sessions/session_06/` (Slides.md "What is Monitoring?" → "Grafana Dashboards", lines ~160–792).
Note: S06 = **Monitoring** (metrics, "is something wrong?"). S08 = **Logging** ("WHY is it wrong?"). Our "monitoring setup" spans both.

## Core concepts (from the slides)
- **What is monitoring?** ISO/IEC 19770-1: "determining the status of a system, a process, or an activity." Josephsen: "Monitoring is for asking questions" — alerts are just ONE possible outcome, not every metric needs one. Turnbull: "Without monitoring you are not doing your job."
- **Monitoring Maturity Model** (Turnbull): Manual/none → **Reactive** (automated, availability-focused, react after breakage) → **Proactive** (measures app performance + business outcomes, used routinely, product not "ready" without instrumentation). Measures *practice/outcomes*, not tools installed.
- (Not yet taught: Pull vs Push, Blackbox vs Whitebox, Passive vs Active, monitoring tactics/categories, Four Metric Types, Prometheus mechanics, Grafana.)

## What WE did
- App self-instruments in `main.go` (`metricsMiddleware` + `/metrics`): `minitwit_http_responses_total` (counter) and `minitwit_http_duration_seconds` (histogram), labelled by method/route/status. Route uses mux template (`/{username}`) to avoid high-cardinality label explosion.
- `monitoring/prometheus/prometheus.yml` scrapes every 15s; alert rules in `prometheus.rules.yml` (WebserverDown, HighErrorRate, SlowResponses).
- Grafana dashboards + Discord alerting provisioned as code under `monitoring/grafana/`.

## Where we sit on the maturity model (exam answer)
- **Reactive** — proactive-grade instrumentation (app-level latency + error metrics, not just CPU/disk), but operated at reactive level: dashboards existed, alerts fired, team rarely read or acted on them.
- Defensible line: *"We're reactive; the gap to close isn't more tooling — it's making the existing data interpretable and part of our routine."* (See findings.md.)

## Open gaps / weak spots
- Team understanding of our own implementation is thin (Leo's words: "missing a lot of understanding in our implementation"). Building it up via tracing the data path file-by-file.
- Rest of S06 concepts still to cover (pull/push, blackbox/whitebox, four metric types, Prometheus internals, Grafana).
