# Session 04 — CI/CD (+ finish simulator API)

Material: `../../course-material/sessions/session_04/` (Slides.md, README_TASKS.md). Guest lecture "What is DevOps?" by Jan (Eficode).

## Concepts (full lecture)
- **Motivation:** building/deploying is scary → automate it → deploy frequently, fix fast.
- **Deploy spectrum:** on the server → manual SSH/SCP → scripted → build system/pipeline.
- **Why deploy often:** big companies (GitHub dozens/day, Amazon every ~11.6s, Facebook daily) — small batches = small blast radius + easy rollback. Avoid "scary deployment days."
- **Tool landscape:** self-hosted (Jenkins, Bamboo, TeamCity, Concourse, Azure DevOps, Drone) vs SaaS (Travis, CircleCI, GitHub Actions, GitLab CI).
- **CI** = clone + build into an artifact (binary/tar/Docker image) + run tests.
- **CD term confusion (IMPORTANT):** the course slide (quoting *Python for DevOps*) defines them OPPOSITE to common usage:
  - Slide: *continuous deployment* = auto to non-prod, manual approval for prod; *continuous delivery* (advanced) = auto to prod + auto-rollback.
  - Common (DevOps Handbook/Jez Humble): *continuous delivery* = manual click to prod; *continuous deployment* = fully auto to prod.
  - Exam tactic: DON'T rely on the label — describe the behavior, then attach a label and note the term is contested.
- **Example pipeline (7 steps):** remote VM+keys → artifact store (DockerHub; alts: Artifactory, GitHub Packages, Maven/NuGet/PyPI/NPM, VCS) → secrets on GitHub → workflow YAML → build & test → deliver (push) → deploy (SSH + run script).
- **Workflow YAML anatomy:** `on: push` + `workflow_dispatch` (manual); steps: checkout, registry login (via `${{ secrets.X }}`), build+push (with registry cache), test, configure SSH, deploy via SSH.
- **Transferable rules:** secrets in CI not in code; build once → store artifact → deploy it; ordered stages; deploy = SSH + script.

## Our project (ci.yml / cd.yml) — verified
- **CI** on PR→master (+ manual): lint (gofmt/golangci-lint/hadolint) → semgrep (SAST) → test (go test + final image + pytest API + Playwright) → docker-scout (CVE scan, fail on critical/high). Richer than the example.
- **CD** on push→master: build+push 3 images to **GHCR** → SSH to swarm manager → `git pull` → `docker stack deploy --with-registry-auth`. Swarm rolling-update across 3 nodes.
- **Better than the example:** tests run BEFORE merge (PR), more gates (SAST + CVE scan), split CI/CD workflows, Swarm rolling update + auto-rollback.
- **Delivery vs deployment for us:** auto to prod on merge, no manual gate = **Continuous Deployment** (common def); by the slide's def (auto + rollback via `rollback_config`) = their "continuous delivery." Describe behavior, not label.

## Gaps (also in findings.md) — VERIFIED via GitHub API
- Releases: manual + retrospective (hand-picked important PRs in hindsight). Task wanted automatic ≥weekly. Not automated.
- Branch protection IS on (ruleset "master lock", no bypass): PR + 1 review + `test` check required, no force-push. BUT only `test` is required — `lint`, `semgrep` (SAST), `docker-scout` (CVE) run but DON'T block merge (advisory). `dev` ruleset only blocks force-push/deletion (no PR/CI).
- CD rebuilds the image rather than promoting the exact CI-tested artifact (test one build, deploy another — same commit).
