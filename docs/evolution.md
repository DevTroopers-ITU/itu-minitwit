# Project evolution

How ITU-MiniTwit went from a cloned Python skeleton to a Swarm-deployed Go service over four months, read off the git log and PR history. Useful as raw material for the report's Reflection section.

## Phases

| # | Phase | Span | Headline | Key PRs |
|---|-------|------|----------|---------|
| 1 | Bootstrapping | Jan 30 – Feb 18 | Python → Go port, simulator API, first Hetzner deploy via Vagrant | [#15](https://github.com/DevTroopers-ITU/itu-minitwit/pull/15), [#32](https://github.com/DevTroopers-ITU/itu-minitwit/pull/32), [#36](https://github.com/DevTroopers-ITU/itu-minitwit/pull/36), [#41](https://github.com/DevTroopers-ITU/itu-minitwit/pull/41) |
| 2 | CI/CD foundations | Feb 24 – Mar 13 | E2E in CI, ghcr.io pipeline, GORM + Postgres, static analysis | [#65](https://github.com/DevTroopers-ITU/itu-minitwit/pull/65), [#70](https://github.com/DevTroopers-ITU/itu-minitwit/pull/70), [#79](https://github.com/DevTroopers-ITU/itu-minitwit/pull/79), [#77](https://github.com/DevTroopers-ITU/itu-minitwit/pull/77) |
| 3 | Observability | Mar 17 – Apr 02 | Prometheus + Grafana, Loki/Promtail, Discord alerts | [#84](https://github.com/DevTroopers-ITU/itu-minitwit/pull/84), [#96](https://github.com/DevTroopers-ITU/itu-minitwit/pull/96), [#112](https://github.com/DevTroopers-ITU/itu-minitwit/pull/112), [#114](https://github.com/DevTroopers-ITU/itu-minitwit/pull/114) |
| 4 | Production infra | Apr 10 – Apr 21 | Docker Swarm, Traefik + Let's Encrypt, perf fixes, UFW | [#118](https://github.com/DevTroopers-ITU/itu-minitwit/pull/118), [#120](https://github.com/DevTroopers-ITU/itu-minitwit/pull/120), [#129](https://github.com/DevTroopers-ITU/itu-minitwit/pull/129), [#131](https://github.com/DevTroopers-ITU/itu-minitwit/pull/131), [#135](https://github.com/DevTroopers-ITU/itu-minitwit/pull/135), [#138](https://github.com/DevTroopers-ITU/itu-minitwit/pull/138) |
| 5 | Hardening | Apr 23 – May 04 | Non-root containers, Semgrep + Docker Scout, multi-stage Dockerfile, Hetzner decommissioned | [#143](https://github.com/DevTroopers-ITU/itu-minitwit/pull/143)–[#160](https://github.com/DevTroopers-ITU/itu-minitwit/pull/160), [#162](https://github.com/DevTroopers-ITU/itu-minitwit/pull/162), [#163](https://github.com/DevTroopers-ITU/itu-minitwit/pull/163) |
| 6 | Wrap-up & report | May 08 – May 16 | Playwright E2E, Terraform spike, report scaffolding | [#158](https://github.com/DevTroopers-ITU/itu-minitwit/pull/158), [#167](https://github.com/DevTroopers-ITU/itu-minitwit/pull/167), [#168](https://github.com/DevTroopers-ITU/itu-minitwit/pull/168), [#169](https://github.com/DevTroopers-ITU/itu-minitwit/pull/169), [#170](https://github.com/DevTroopers-ITU/itu-minitwit/pull/170) |

## Phase notes

**1. Bootstrapping.** Initial commit `ae39e23` (Jan 30) was the upstream Python MiniTwit. Two weeks later [#15](https://github.com/DevTroopers-ITU/itu-minitwit/pull/15) replaced it with the Go port on Gorilla Mux. [#36](https://github.com/DevTroopers-ITU/itu-minitwit/pull/36) added the simulator API + Swagger so we could pass the grader. [#41](https://github.com/DevTroopers-ITU/itu-minitwit/pull/41) put it on a Hetzner VM via Vagrant — the first version anyone outside the laptops could reach.

**2. CI/CD foundations.** [#65](https://github.com/DevTroopers-ITU/itu-minitwit/pull/65) wired GitHub Actions to ghcr.io and turned branch protection on. [#70](https://github.com/DevTroopers-ITU/itu-minitwit/pull/70) introduced a database abstraction layer, [#79](https://github.com/DevTroopers-ITU/itu-minitwit/pull/79) moved the production database to managed Postgres on DigitalOcean. Static analysis ([#77](https://github.com/DevTroopers-ITU/itu-minitwit/pull/77): gofmt, golangci-lint, hadolint) started gating PRs.

**3. Observability.** [#84](https://github.com/DevTroopers-ITU/itu-minitwit/pull/84) added Prometheus middleware + a Grafana stack. The first revision had unbounded label cardinality from raw paths; [#96](https://github.com/DevTroopers-ITU/itu-minitwit/pull/96) and [#98](https://github.com/DevTroopers-ITU/itu-minitwit/pull/98) normalised routes and capped memory. [#112](https://github.com/DevTroopers-ITU/itu-minitwit/pull/112) added Loki + Promtail; [#114](https://github.com/DevTroopers-ITU/itu-minitwit/pull/114) wired Discord alerts on downtime, 5xx rate, and p95 latency.

**4. Production infra.** [#118](https://github.com/DevTroopers-ITU/itu-minitwit/pull/118) brought Traefik + Let's Encrypt in front of the app. [#120](https://github.com/DevTroopers-ITU/itu-minitwit/pull/120) was the big one — three-node Docker Swarm on DigitalOcean, app deployed as a replicated service. Then a week of follow-ups: [#128](https://github.com/DevTroopers-ITU/itu-minitwit/pull/128) (secret path mismatch), [#129](https://github.com/DevTroopers-ITU/itu-minitwit/pull/129) (ingress host mode, fixed an HTTP/2 504), [#131](https://github.com/DevTroopers-ITU/itu-minitwit/pull/131) (Traefik v3.6 + overlay network pinning). [#135](https://github.com/DevTroopers-ITU/itu-minitwit/pull/135) replaced a 41–49 s JOIN on the personal timeline with a subquery + connection pool; [#138](https://github.com/DevTroopers-ITU/itu-minitwit/pull/138) moved the `latest` counter from an in-process var to Postgres so it survived rolling deploys.

**5. Hardening.** Container hardening turned into [#143](https://github.com/DevTroopers-ITU/itu-minitwit/pull/143)–[#160](https://github.com/DevTroopers-ITU/itu-minitwit/pull/160) — non-root user, multi-stage Dockerfile to strip binutils CVEs, hadolint and lint version bumps. Semgrep + Docker Scout joined CI as blocking gates. [#162](https://github.com/DevTroopers-ITU/itu-minitwit/pull/162) decommissioned the original Hetzner VM and Vagrant path.

**6. Wrap-up & report.** [#158](https://github.com/DevTroopers-ITU/itu-minitwit/pull/158) scaffolded `report/`. [#167](https://github.com/DevTroopers-ITU/itu-minitwit/pull/167) prototyped Terraform for the Swarm nodes (informational; not deployed). [#168](https://github.com/DevTroopers-ITU/itu-minitwit/pull/168) added Playwright browser tests for the user-facing flows. [#169](https://github.com/DevTroopers-ITU/itu-minitwit/pull/169) and [#170](https://github.com/DevTroopers-ITU/itu-minitwit/pull/170) filled in report sections.

## Monthly activity

Commits on `master`, merges excluded:

```
Jan |||||||  6
Feb |||||||||||||||||||||||||||||||||| 34
Mar ||||||||||||||||||||||||||||||||||||||||||||||||||||||| 55
Apr ||||||||||||||||||||||||||||||||||||| 37
May ||||||||||||||| 15
```

The March peak is the observability and CI work; the April plateau is Swarm rollout and the follow-up firefighting; the May tail is hardening plus the report itself. All-branches (incl. unmerged WIP and the report storyboard branch) is roughly 6 / 35 / 63 / 43 / 65 over the same months — the May number is inflated by report iteration.

## Contributors

`git shortlog -sn --all --no-merges` (post-`.mailmap`):

| Author | Commits |
|--------|---------|
| Leo Sakharoff | 88 |
| Peter Juul Møller | 53 |
| Apoorva Sood | 27 |
| Håkon Refsvik | 27 |
| Peter Kvist | 14 |

## Honest observations

- The `feature/hardenContainers` branch went through ~15 merged PRs in a few hours on Apr 24 ([#146](https://github.com/DevTroopers-ITU/itu-minitwit/pull/146)–[#151](https://github.com/DevTroopers-ITU/itu-minitwit/pull/151)) chasing a Go version bump through golangci-lint. That's not a clean history, but it's an accurate trace of "make CI green again."
- Two big-bang `dev → master` PRs ([#88](https://github.com/DevTroopers-ITU/itu-minitwit/pull/88), [#122](https://github.com/DevTroopers-ITU/itu-minitwit/pull/122)) bundled a week of feature work each. Useful for hitting deadlines, less useful for bisecting later — the Traefik 504 ([#129](https://github.com/DevTroopers-ITU/itu-minitwit/pull/129)) and the secret-path bug ([#128](https://github.com/DevTroopers-ITU/itu-minitwit/pull/128)) both surfaced *after* one of these landed in prod.
- `feature/code-quality` (SonarCloud, [#78](https://github.com/DevTroopers-ITU/itu-minitwit/pull/78)) was opened Mar 13 and never merged. We replaced it with Semgrep + Codacy later, but the open PR is still there.
- The Python reference app (`python-references/`) survived the whole project as a pytest harness against the simulator API. It's the only Python that's still load-bearing.
