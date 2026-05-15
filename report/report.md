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
<!--
Static analysis output (golangci-lint, hadolint, Semgrep, Docker Scout),
quality assessment, known issues.
-->

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
<!-- What we monitor and with what tools. Reference the Grafana dashboard. -->

## Logging
**Author(s):** Peter J
<!-- What we log, how we aggregate (Loki + Promtail), how we query. -->

## Security Hardening
**Author(s):** Peter J 
<!--
Container hardening (non-root, multi-stage), Semgrep + Docker Scout, UFW,
DO firewall, GHCR auth, TLS via Traefik.
-->

## Availability and Scaling 
**Author(s):** Peter J, Apoorva
<!--
Swarm replicas, placement constraints, what we did and what we left out
and why.
-->

# Reflection Perspective
<!-- Suggested word budget: ~500 -->

## Evolution and Refactoring
**Author(s):** Håkon and Leo
<!-- Biggest issues and how we solved them. Link commits/issues. -->

## Operation
**Author(s):** Leo and Apoorva
<!-- Incidents, on-call lessons, what changed in how we run the system. -->

Choices about servers, droplets, databases
Vertical/horisontal scaling

## Maintenance
**Author(s):** Leo
<!-- What's hard to maintain, what we improved, what's still ugly. -->

## DevOps Style
**Author(s):** Leo
<!--
What was different from previous projects and how it worked out.
Be honest about trade-offs.
-->

# Use of Generative AI
**Author(s):** Leo

<!-- Suggested word budget: ~200. Required per ITU GAI policy. -->

<!--
Which tools we used, for which tasks, how, and a short reflection on
whether they helped or hindered the work.
-->
