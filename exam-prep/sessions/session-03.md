# Session 03 — Provisioning VMs (Vagrant/DigitalOcean) + simulator API

Material: `../../course-material/sessions/session_03/` (Slides.md, README_TASKS.md, API_Spec/)

## Two tasks
1. Implement the **simulator API** (`/register`, `/msgs`, `/fllws`, `/latest`) to an OpenAPI/Swagger spec. We have `swagger.json` + `swagger.html` (served via `main.go:86`). Course gave the spec + `minitwit_sim_api_test.py` to check it.
2. Deploy to a **cloud VM**, provisioned **reproducibly by code** (IaC), public IP/domain, documented in README. IaaS not PaaS.

## Core concept — Infrastructure as Code (IaC)
Create servers by running code (Vagrant/Terraform/script), version-controlled, reproducible — no clicking in a web UI.
Benefits: reproducibility; version history (git/PRs, revertable); documentation (code = setup record); disaster recovery (rebuild by re-running); no drift / no "snowflake" servers.
Course rule: no manual UI clicking; no PaaS (Heroku etc.) — use a VM you control.

## Our infra EVOLUTION (great for Evolution & Refactoring reflection)
- Early (S03): `Vagrantfile` → single VM on **Hetzner** (Ubuntu 22.04). Real IaC.
- Now: **3-node Docker Swarm on DigitalOcean**, fronted by Traefik + Let's Encrypt. Hetzner decommissioned (PR #162).
- Terraform added later (S12 exercise).

## HONEST GAP (verified) — IaC is partial
- Prod DO droplets were **provisioned manually** (docs: "a colleague had already provisioned the DO droplets").
- `terraform/` written AFTER the fact ("edited to match our current setup"), only 4 commits, **never run to provision, not validated**. = snowflake server + untested IaC.
- README overstates Terraform as the provisioning method. Don't over-claim.
- Exam line: "early Hetzner = real Vagrant IaC; current prod = manual, Terraform is descriptive-only and unvalidated; with more time we'd prove `terraform apply` from scratch."

## Full lecture content (for completeness)
- **Provisioning ladder (7 rungs):** manual physical → manual local VM (VirtualBox) → scripted local VM → Vagrant local VM → manual cloud VM (DO UI) → cloud VM via API (curl/script) → cloud cluster via Vagrant. Trend: manual→automatic, local→remote, one→many.
- **VMs & hypervisors:** VM = full guest OS. Hypervisor = software that runs VMs on a host. Type 1/bare-metal (Xen, ESXi, Hyper-V) vs type 2/hosted (VirtualBox, Parallels). VM != VirtualBox (also KVM, QEMU, VMware...).
- **VM vs container:** VM = full OS, heavy; container = shares host kernel, light. Why VMs? a skill; run any OS. Our setup = containers running ON a VM host (both, layered).
- **Vagrant:** `Vagrantfile` (Ruby) describes a VM (box, ports, synced folders, RAM, provider); multi-machine in one file; providers VirtualBox(local)/DO/AWS(remote). Packer builds OS images. NB: Vagrant's `provision` step = **configuration management** (S5).
- **OpenAPI/Swagger:** machine-readable API description (endpoints, methods, payloads, status codes) for humans + tools (docs, code-gen stubs, tests). Course gave us the spec; we have `swagger.json` + Swagger UI at `/swagger`.
- **Process feedback (exam-relevant):** keep legacy code (separate repo / subfolder / old branch / git tag); meaningful branch names; distribute work; use GH issues; **log decisions** (why Go + comparison).

## Open gaps / to verify
- Is the `Vagrantfile` still used at all, or pure legacy? (likely legacy)
