---
title: ITU-MiniTwit
subtitle: Group Q — DevTroopers
author:
  - Leo Sakharoff
  - Peter Juul Møller
  - Peter Kvist
  - Apoorva Sood
  - Håkon Refsvik
date: \today
---

# System's Perspective

<!-- Suggested word budget: ~700 -->

## Design and Architecture 
**Author(s):** Peter K, Håkon
<!--
Describe and illustrate the system. Include diagrams from at least the
allocation and component-and-connector viewpoints (see session_12
Documentation.md). One UML deployment diagram + C&C is sufficient, maybe or maybe not one sequence diagram is a
good minimum.
-->

  - Deployment (allocation) diagram. Docker swarm overview / serverside understanding 
  - C&C viewpoint explaining the runtime components and their communication
  - Sequence diagram (maybe not so important - see Mirceas lecture slides)

## Dependencies
**Author(s):** Peter K
<!--
List and briefly describe all technologies and tools we depend on across
all stages: language/framework, infra, CI/CD, observability, third-party
services. Group by layer.
-->

## Current State
**Author(s):** Peter J

We run four static analysis tools in CI on every pull request, all blocking merges if they fail.

**golangci-lint** is our main Go linter. We have `errcheck`, `govet`, `staticcheck`, `ineffassign`, and `unused` enabled. `errcheck` catches unchecked errors in Go code, `govet` catches common mistakes like misused format strings, `staticcheck` does deeper analysis for bugs and dead code, and `ineffassign` and `unused` flag variables or code that serve no purpose.

**hadolint** lints the Dockerfile against a set of best practice rules — things like pinning base image versions, avoiding `apt-get upgrade`, and making sure `RUN` layers are structured cleanly.

**Semgrep** does static application security testing (SAST) using three rulesets: `p/security-audit` for general security issues, `p/secrets` for accidentally committed credentials, and `p/owasp-top-ten` for common web vulnerabilities like SQL injection and XSS.

**Docker Scout** scans the final built Docker image for known CVEs in the OS packages and dependencies. It's configured to fail the build if any critical or high severity vulnerability is found, so we catch issues in base image dependencies before they reach production.

# Process Perspective
<!-- Suggested word budget: ~1100 -->

## CI/CD Pipeline
**Author(s):** Peter K, Håkon & Apoorva

<!--
Stages and tools, from push to deployed replica. Cover deploy and release.
Activity diagram fits well here.
-->

  - Docker swarm in CD pipeline 

## Monitoring
**Author(s):** Peter J, Apoorva

For monitoring we use Prometheus to scrape metrics from the app, and Grafana to visualize them. Inside the Go application we added a middleware that wraps every HTTP handler and records two custom metrics: `minitwit_http_responses_total` (a counter per route, method, and status code) and `minitwit_http_duration_seconds` (a latency histogram). These get exposed on `/metrics` and Prometheus scrapes all three webserver replicas every 15 seconds using DNS-based service discovery on the Swarm overlay network — so adding or removing replicas requires no config changes.

We set up three alert rules in Prometheus: one that fires if a webserver replica has been unreachable for over a minute, one for when more than 10% of responses are 5xx errors over a five-minute window, and one for when P95 latency goes above 1 second for five minutes. All alerts are routed to a Discord channel through a Grafana contact point, using a webhook URL stored as a Docker Swarm secret so it never ends up in the codebase.

The Grafana setup is fully provisioned from code — datasources and dashboards are baked into a custom Docker image. The main dashboard covers uptime and availability across all replicas, total HTTP responses broken down by route, P95 response time, and a combined error rate panel. Over the week of 9–15 May we saw 100% uptime, and P95 response times stayed comfortably under 50 ms for the vast majority of the time.

![Availability dashboard showing 100% uptime and HTTP response breakdown](images/monitor-availability.png)

![HTTP metrics: P95 response time per route, requests per route, and error rate](images/metrics-response-time.png)

## Logging
**Author(s):** Peter J

For logging we went with the Grafana LGTM stack — Promtail collects logs from every container and ships them to Loki, and Grafana lets us search and explore them.

Promtail runs as a global Swarm service, meaning one instance per node, so no replica gets skipped. It watches the Docker socket and auto-discovers containers, pulling in logs from all of them. For each container it tags logs with three labels: the container name, the Swarm service name (like `minitwit_webserver`), and the specific task/replica. These labels make it easy to filter down to exactly what you want in Grafana's log explorer.

On the Loki side, we use a simple single-node setup with filesystem storage and daily index rotation. Logs older than 7 days are dropped at ingestion to keep disk usage in check. The app itself just writes to stdout via Go's standard `log` package, which Docker captures automatically — no extra log library needed. We also enabled Traefik access logs so request-level data from the reverse proxy flows into Loki through the same pipeline, which turned out to be really useful for debugging.

The logging dashboard below shows log volume per Swarm service over time. Day-to-day it stays low and stable, with the occasional spike from Traefik or Loki itself when there is a burst of activity.

![Log volume by Swarm service](images/log-volume-swarm.png)

## Security Hardening
**Author(s):** Peter J

**Multi-stage Docker build.** The app Dockerfile uses two stages: a builder stage that compiles the Go binary, and a final `alpine:3.23` image that only contains the compiled binary, templates, and static files. The build tools are thrown away entirely, so the attack surface of the running container is minimal. We also create a dedicated non-root user (`appuser`) inside the image and drop to that user before the process starts, so even if something goes wrong inside the container it doesn't have root.

**Firewall.** Every node gets UFW rules applied automatically during Terraform provisioning. The manager node opens ports 80, 443, and 8080 for Traefik, plus SSH and the three Docker Swarm internal ports. Worker nodes only have the Swarm ports and SSH open — no public ports at all, so all traffic must go through Traefik on the manager. The database node only exposes SSH and Postgres (5432). On top of that, DigitalOcean's cloud firewall adds another layer at the network edge.

**TLS.** Traefik handles HTTPS via Let's Encrypt automatically. All HTTP traffic on port 80 is redirected to HTTPS, and certificates are stored in a named Docker volume so they persist across redeploys.

**Secrets.** Sensitive values like the database URL, app secret key, Discord webhook, and GHCR credentials are kept as Docker Swarm secrets or GitHub Actions secrets — never hardcoded or baked into images.

**Static analysis.** The CI pipeline runs `golangci-lint` on Go code, `hadolint` on the Dockerfile, and Semgrep for SAST on every push. These run as parallel jobs and block merges if they fail.

## Availability and Scaling
**Author(s):** Peter J, Apoorva

The webserver runs as three replicas with a `spread: node.id` placement preference in Docker Swarm, which tries to put replicas on different nodes so a single node going down doesn't take the whole service with it. Traefik automatically distributes traffic across whichever replicas are healthy.

Rolling updates are set to deploy one replica at a time with a 10-second pause between steps and `order: start-first` — this means the new replica comes up and passes its health check before the old one is taken down, so there is no gap in availability during a deploy. If an update goes bad, the rollback config mirrors the same approach, replacing one replica at a time.

Each replica runs a health check every 30 seconds (`wget --spider` against localhost) and Swarm stops routing to a replica if it starts failing. Combined with the Prometheus `WebserverDown` alert, we get two independent signals if something goes wrong — the load balancer stops sending traffic and we get a Discord notification within a minute.

The monitoring services (Prometheus, Grafana, Loki) each run as a single replica pinned to the manager node. We could have run them in a more resilient configuration, but these services all hold persistent state that is genuinely tricky to replicate without extra tooling, and a brief monitoring outage is much less bad than a complex distributed setup breaking in production. The webserver replicas are stateless and are the only part of the system we actually scale horizontally. If we ever needed more capacity on the database or monitoring side, vertical scaling (resizing the droplet) is the more practical option.

# Reflection Perspective

## Evolution and Refactoring
**Author(s):** Håkon and Leo

The project moved through roughly six phases: bootstrapping, CI/CD, observability, production infrastructure, hardening, and wrap-up. Between these phases, most refactoring happened through smaller fixes and upgrades rather than one planned redesign.

The largest architectural change was the move from a single Hetzner deployment to a three-node Docker Swarm cluster on DigitalOcean (PR #120 and follow-ups). The migration changed the system from one server running everything to a distributed setup with replicated webservers, Traefik routing, managed PostgreSQL, and Swarm secrets. Running three webserver replicas behind Traefik also forced us to externalise shared state we had previously kept in memory — most visibly the `latest` simulator counter, which we moved into PostgreSQL (PR #138) so the replicas could agree.

The same pattern appeared elsewhere: solutions that worked at one scale often broke at the next. Metrics labels initially used raw paths until cardinality became an issue, and the personal timeline query appeared acceptable until realistic usage caused severe timeouts.

In hindsight, our refactoring was largely reactive rather than planned. This kept development moving, but also meant architectural weaknesses were often discovered only under operational pressure.

## Operation
**Author(s):** Leo and Apoorva

We operated two production environments in parallel for most of April: a single-node Hetzner deployment and a Docker Swarm cluster on DigitalOcean, both connected to the same managed PostgreSQL instance. This allowed a gradual migration without interrupting the simulator.

That parallel setup also exposed our most important operational lesson. On 16 April, a newly configured DigitalOcean firewall blocked Docker Swarm's internal overlay traffic between nodes. The manager continued serving traffic through its local replica, so external uptime checks remained green while cluster redundancy had silently failed for nearly 18 hours; PR #131 later addressed related Swarm routing issues.

The incident reinforced that control-plane and data-plane failures are not the same, and that edge-level uptime checks alone are insufficient. Running three replicas behind Traefik gave us horizontal replication, though we never benchmarked whether this was meaningfully better than the simpler single-node setup. The manager also remains a single point of failure for several critical services.

## Maintenance
**Author(s):** Leo

Our maintenance story was largely reactive. Tooling improved steadily throughout the project, with linting, security scanning, and image hardening added over time rather than as part of an explicit maintenance strategy. A concrete example was Codacy, which immediately identified a `/health` route bug that had gone unnoticed for weeks (fixed in commit `c8ff76c`, whose message reads *"caught by codacy"*).

More generally, issues that affected visible behaviour were fixed, while quieter problems remained. For example, bcrypt errors are still swallowed in helper code, simulator authentication contains hardcoded values, test coverage remains limited, and logs accumulated recurring warnings that nobody investigated.

The main lesson is that maintenance requires ownership. Improvements happened when a specific issue became painful enough to address, not because we systematically worked to improve maintainability.

## DevOps Style
**Author(s):** Leo

This was the first project where most of us were responsible not only for development, but also for deployment and operations. That changed how we worked.

Applying the DevOps Handbook's Three Ways, our strongest area was flow. We established pull requests and continuous deployment early (PR #65), which created a clear delivery path and fast iteration. Initial branch protection on both `dev` and `master` felt too heavy for day-to-day work, so we relaxed `dev` and kept `master` as the stricter integration gate. In practice, PRs often functioned more as coordination and deployment checkpoints than as strict human review gates.

Feedback was more mixed. Monitoring helped us detect some operational issues quickly, including performance degradation in the timeline query, but other failures went unnoticed because our monitoring assumptions were incomplete.

Continual learning was ad-hoc rather than systematic. We never established a documentation or estimation practice. Coordination mostly happened through Discord pings on PRs, while deeper docs were written only for larger refactors or incidents, such as the live debug doc from the 17 April outage (`docs/incidents/session11-ops-debug.md`). Operational knowledge therefore remained concentrated among a few contributors.

The main takeaway is that DevOps was not just about adding tools. It became concrete when we had to operate the system ourselves.

# Use of Generative AI
**Author(s):** Leo

We used Anthropic Claude, mainly through the Code interface, throughout the project. AI-assisted commits carry `Co-Authored-By: Claude`, and `.mailmap` maps the tool to `LLM <none>` as required by the course.

The clearest place AI helped was the early refactor of the inherited Python/Flask app into Go (PR #15). None of us knew Go; Claude helped us scaffold the package structure and read errors as we learned the type system. Without it, the rewrite would likely have taken much longer. The same was true later for Docker Swarm, Traefik, the PostgreSQL migration, CI/CD setup, and security tooling: AI reduced iteration time by helping explain errors and suggest configurations.

Its usefulness depended on active validation, though. Plausible but wrong suggestions sometimes slowed debugging rather than helping, and we came to treat AI as a fast exploratory assistant rather than an authoritative source.

A further reflection is that AI use was not evenly distributed within the team. It increased individual productivity, but also created asymmetry in how quickly contributors could work across unfamiliar technical areas. Overall, generative AI improved development speed, but only when paired with technical judgment and active validation.