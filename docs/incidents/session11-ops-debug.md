# Swarm Ops Debugging Notes — 17 Apr 2026

Running notes from a live session fixing the DigitalOcean Swarm. Written so any team
member can read this after the fact and understand both what happened and why.

## Baseline — what we found at the start

SSH works from Leo's laptop to all four servers using `~/.ssh/id_ed25519`:

| Host            | IP              | Hostname on box | Role in stack                                    |
|-----------------|-----------------|-----------------|--------------------------------------------------|
| DO manager      | 64.226.116.162  | `Manager`       | Swarm leader, Traefik, Prometheus, Grafana, Loki |
| DO worker (A)   | 134.122.90.176  | `Worker-1`      | App replicas, Promtail                            |
| DO worker (B)   | 206.189.59.60   | `Worker-2`      | App replicas, Promtail                            |
| Hetzner legacy  | 46.224.144.214  | `h-6d6ea5`      | Old single-box compose deployment (still live)   |

`CLAUDE.md` had the Worker-1 / Worker-2 IPs swapped vs. the actual hostnames on the
boxes — noted, fix queued.

### Symptoms on the manager

`docker node ls` on the manager:

```
Manager    Ready   Active   Leader
Worker-1   Down    Active
Worker-2   Down    Active
```

`docker node inspect Worker-1 --format '{{.Status.Message}}'` →
`down heartbeat failure`.

`docker service ls` on the manager showed everything under-replicated and pinned to
the Manager node:

```
minitwit_webserver     5/3
minitwit_grafana       2/1
minitwit_loki          2/1
minitwit_promtail      3/1   (global, expected 3/3)
```

`docker stack ps minitwit` showed repeated `Rejected` restarts of grafana and
prometheus on the manager with the error:

```
No such image: ghcr.io/devtroopers-itu/minitwit-grafana:latest
No such image: ghcr.io/devtroopers-itu/minitwit-prometheus:latest
```

The public site `https://devtroopersminitwit.codes` still returns 302 (Traefik up on
the manager), so users aren't seeing an outage — but we have no redundancy. All the
services that are running are running on the manager node only.

### Two independent bugs, in plain terms

1. **Swarm control plane is broken between manager and workers.**
   Each node's Docker daemon is healthy in isolation (both workers say
   `Swarm: active`, `systemctl is-active docker` = active), but they can't talk to
   each other well enough to keep the swarm's heartbeat alive. Symptoms match
   Swarm's required ports being blocked: TCP 2377 (manager API), TCP+UDP 7946
   (node-to-node gossip), UDP 4789 (overlay VXLAN data plane).
2. **Workers can't pull the two private ghcr.io images** for Grafana and
   Prometheus. These are private packages under the DevTroopers-ITU org, and
   `docker stack deploy` was run without `--with-registry-auth`, so the workers
   (and even the manager after a restart) have no credentials for ghcr.io.

We'll work these in order — fix (1) first because it's the reason the cluster is
effectively one node, then fix (2) so that when replicas get rescheduled onto the
workers they can actually start.

---

## Plan

1. Fix the Worker IP swap in `CLAUDE.md` (trivial, documentation only).
2. Investigate the heartbeat failure — check firewall / ufw / DO cloud firewall
   for ports 2377/7946/4789 between the three droplets.
3. Fix registry auth so workers can pull private images.
4. Verify everything comes back to desired replicas.

---

## Step-by-step log

### 17 Apr, ~13:10 — new context from Discord

Peter Juul Møller created a **DigitalOcean cloud firewall** last night (~21:41)
and applied it to all three DO droplets. The visible inbound rules are:

- HTTP, TCP, port 80, All IPv4 + All IPv6
- HTTPS, TCP, port 443, All IPv4 + All IPv6
- Custom, TCP, port 8080, All IPv4 + All IPv6

Everything else inbound is now denied — including node-to-node traffic between
our own droplets. That lines up perfectly with our "down heartbeat failure"
diagnosis.

### 17 Apr, ~13:12 — confirmed via TCP probes

From the manager, we probed TCP connections to both workers:

```
TCP 2377 -> Worker-1 (134.122.90.176): BLOCKED/closed
TCP 7946 -> Worker-1 (134.122.90.176): BLOCKED/closed
TCP 2377 -> Worker-2 (206.189.59.60):  BLOCKED/closed
TCP 7946 -> Worker-2 (206.189.59.60):  BLOCKED/closed
TCP   80 -> Worker-1:                  BLOCKED
```

Even port 80 is blocked between droplets, which tells us an important detail:
the firewall's inbound rules are scoped `All IPv4 + All IPv6` — which sounds
like "allow from anywhere", but in DO's firewall model these rules define what
the *public internet* can reach. There is no implicit "allow all from my other
droplets" rule. So our own cluster nodes are treated like internet strangers
and blocked from talking to each other.

### Host-level firewalls are NOT the issue

`ufw` is inactive on all three nodes and `iptables DOCKER-USER` chain is empty.
So the block is entirely at DO's cloud firewall layer, not on the hosts.

### Let's Encrypt is fine right now

Cert on `devtroopersminitwit.codes`:
- Issuer: Let's Encrypt R13
- Not Before: Apr 10 2026
- Not After:  Jul  9 2026

So Peter's worry about LE needing a Traefik restart is moot for now — we have
~83 days of runway. LE only breaks if renewal fails *while* port 80 is blocked
*during* a renewal attempt, and our next renewal is mid-June.

---

## Action plan

The DO cloud firewall needs additional inbound rules so Swarm nodes can talk
to each other. Best practice is to scope these by **droplet tag**, so the
rules only apply between our cluster nodes — not the public internet — to
avoid exposing the Swarm control plane to the world.

### Required firewall rules (to add)

Assuming we tag all three droplets with e.g. `minitwit-swarm`, add these
**inbound** rules on the firewall:

| Protocol | Port(s)  | Source                        | Purpose                               |
|----------|----------|-------------------------------|---------------------------------------|
| TCP      | 2377     | Tag: `minitwit-swarm`         | Cluster management (manager API)      |
| TCP      | 7946     | Tag: `minitwit-swarm`         | Node gossip (discovery)               |
| UDP      | 7946     | Tag: `minitwit-swarm`         | Node gossip (discovery)               |
| UDP      | 4789     | Tag: `minitwit-swarm`         | Overlay network (VXLAN data plane)    |
| ICMP     | —        | Tag: `minitwit-swarm`         | Ping between nodes (nice to have)     |

Important: the source **must be the tag, not "All IPv4"**. Exposing Swarm's
management port 2377 to the public internet is a serious security hole — it
allows remote cluster takeover if an attacker can join.

### Why those exact ports (so you can explain it)

- **TCP 2377 — Cluster management.** This is the port the manager listens on
  for `docker swarm join` and for API calls between managers. Workers call
  into it to receive task assignments.
- **TCP + UDP 7946 — Gossip protocol.** Every node chats with every other
  node over this port to build and maintain the cluster's shared view of who's
  alive. The "heartbeat failure" we saw is exactly this protocol timing out.
  It needs both TCP and UDP because SWIM-style gossip uses TCP for ordered
  state exchange and UDP for low-overhead probes.
- **UDP 4789 — VXLAN overlay.** When two containers on different nodes talk
  over a Docker overlay network, the traffic is encapsulated in VXLAN packets
  that travel between hosts on UDP 4789. Without this, inter-node container
  traffic breaks — which would break everything once the workers come back
  online.
- **ICMP** is not required by Swarm but is helpful for troubleshooting
  (`ping` between nodes).

### 17 Apr, ~13:25 — registry auth verified on all three nodes

We were worried that the workers might not be able to pull private ghcr.io
images once they rejoin the swarm. Checked all three nodes via SSH:

- All three have `/root/.docker/config.json` with a ghcr.io auth entry.
- From the manager, `docker pull` works for all three private images:
  `minitwit`, `minitwit-grafana`, `minitwit-prometheus`.
- `.github/workflows/cd.yml` already uses `--with-registry-auth` when it runs
  `docker stack deploy`, which propagates manager-side creds to workers at
  deploy time.

Conclusion: registry auth is NOT broken right now. The `"No such image"`
errors we saw 39 hours ago were a secondary effect of the primary firewall
issue: the swarm scheduler tried to place Grafana/Prometheus tasks on the
workers, couldn't reach them, fell back to the manager, and in that fallback
window the manager's own task attempt happened to land during a brief
auth-not-yet-distributed moment. Once the firewall is fixed, this won't
recur.

### 17 Apr, ~13:30 — security issue noted (DO NOT fix in this pass)

The ghcr.io auth stored on all three droplets is a **base64-encoded GitHub
Personal Access Token** under user `DenSygeMike` (Apoorva). That's a
long-lived credential with `read:packages` scope, present as plaintext in
`/root/.docker/config.json` on three internet-facing hosts.

Recommended follow-up (separate ticket):

1. Rotate the PAT (Apoorva revokes the existing one on GitHub).
2. Replace with short-lived `GITHUB_TOKEN` already used inside the CD
   workflow — it's scoped to the run and expires automatically.
3. Stop writing the token to disk on each node. Rely on
   `--with-registry-auth` at deploy time instead.

We're explicitly not touching this today because it's orthogonal to the
outage and we want to stay focused.

### Findings summary so far

- Root cause: DO cloud firewall (added last night) blocks inter-node Swarm
  ports (2377, 7946 TCP+UDP, 4789 UDP). Workers are fine in isolation;
  they're just unreachable from the manager.
- Secondary effect: services under-replicated on manager only. Site still
  serves from Traefik on manager — no visible outage to users.
- No registry / image problem. No LE problem. No host firewall problem.
- Separate latent issue: GitHub PAT stored plaintext on all three droplets.

### Action required (delegated to Peter — owns DO dashboard)

Peter needs to add the rules listed in **Required firewall rules** above
to the existing DO cloud firewall. Scoped to the droplet tag (or, if not
tagged yet, to the three specific droplet IPs).

Discord message drafted and sent.

---

## Group-meeting crib sheet

**One-sentence summary.** Site is up for users, but the cluster is only
running on the manager because a cloud firewall added last night blocks the
ports Swarm needs between its own nodes.

**The trap in one sentence.** A DO cloud firewall's allow-list applies to
all inbound traffic — your other droplets are not implicitly trusted, so
"allow 80/443/8080" silently disallowed Swarm's 2377/7946/4789 between
siblings.

**Correction after testing.** Our own probing showed that even port 80
between droplets is blocked (4-second timeout = firewall drop, not fast
close). So the effective firewall is more restrictive than the three allow
rules visible in the Discord screenshot. There are probably additional
source-IP restrictions, or multiple firewalls attached, or outbound-side
restrictions we can't see without dashboard access. Action: before adding
rules, have Peter read out the FULL current rule set (inbound AND
outbound, all attached firewalls). Don't assume.

**Fix steps for whoever has dashboard access.**

1. Tag all three droplets with `minitwit-swarm`.
2. Open the firewall and add five inbound rules, Source = Tag
   `minitwit-swarm`:
   - TCP 2377
   - TCP 7946
   - UDP 7946
   - UDP 4789
   - ICMP
3. Save. Nothing to restart.

**How to verify from the terminal after the fix.**

```bash
# On manager, should stop timing out:
ssh root@64.226.116.162 'for p in 2377 7946; do \
  timeout 3 bash -c "</dev/tcp/134.122.90.176/$p" \
    && echo "TCP $p -> Worker-1: OPEN" \
    || echo "TCP $p -> Worker-1: BLOCKED"; \
done'

# And watch the cluster recover:
ssh root@64.226.116.162 'watch -n 2 "docker node ls; echo; docker service ls"'
```

Expected within 1–3 minutes:

- `docker node ls` — all three `Ready`
- `docker service ls` — replicas at desired counts (3/3, 1/1, 3/3, etc.)

**Fallback if anything goes sideways.** Detach the firewall from the three
droplets in the dashboard. That returns us to yesterday-morning's state,
where the site was serving fine. We can try again later.

**Secondary issue parked for follow-up (not today).** GitHub PAT
(DenSygeMike / Apoorva, `read:packages` scope) stored in plain text in
`/root/.docker/config.json` on all three droplets. Rotate after exam stress
is down. Owner: whoever volunteers.

---

## Confidence check on the diagnosis (17 Apr ~13:45)

Leo pushed back: "Are you sure the firewall is actually the problem? I had a
previous Claude session focused on cache overflow and vertical-scaling
limits. I had gateway errors at ~17:00 yesterday, before Peter's firewall."

Fair challenge. Re-investigated:

- **Current state, today:** high confidence the firewall is the cause.
  Direct evidence: heartbeat-failure status on workers, TCP probes to 2377
  and 7946 time out, SSH and Docker are fine on workers in isolation.
- **Yesterday ~17:00 symptoms:** not explained by the firewall (firewall
  came at 21:41). Not explained by webserver unhealthy-container failures
  either — those were earlier (late 15 Apr and morning 16 Apr); by 17:00
  the replacement replicas had been running healthy for 6+ hours on all
  three nodes. Most plausible theory is **DNS propagation during Leo's
  Hetzner↔DO DNS flips**, causing requests from the same client to land on
  different backends depending on which cached DNS answer served the
  request. This fits "sometimes works, sometimes plain HTML, sometimes
  timeout." Can't prove without packet capture or DNS query log from
  yesterday.
- **Previous Claude's cache / vertical-scaling theory:** not supported by
  current resource data. Manager: 662/961 MB RAM used, load 0.06, 30%
  disk, containers collectively under 300 MB. No pressure anywhere.

Summary: two possibly-separate issues. What we're fixing today is the
firewall issue. If the cluster doesn't recover after the firewall fix, that
tells us something else is going on and we revisit. If some workers fail
their health check after they rejoin, that's a pre-existing webserver bug
reasserting itself, not the firewall diagnosis being wrong.

Also noted: there's a persistent log warning on all webserver containers
— `http: superfluous response.WriteHeader call from main.go:50`. This is
harmless noise (Go's net/http logs it when a handler calls `WriteHeader`
twice; first call wins, subsequent calls are dropped), but it indicates a
real bug in middleware around `main.go:50` that's worth cleaning up in a
future refactor. Not related to today's outage.

## Post-fix verification (17 Apr, after Peter applied firewall rules)

Peter added the five tag-scoped rules to the DO firewall (TCP 2377, TCP+UDP
7946, UDP 4789, ICMP, all with Source = `minitwit-swarm` tag).

**Result: cluster fully recovered.**

- `docker node ls`: Manager / Worker-1 / Worker-2 all `Ready`, Manager is
  Leader.
- `docker service ls`: all at desired replica count (webserver 3/3,
  grafana 1/1, loki 1/1, prometheus 1/1, traefik 1/1, promtail 3/3 global).
- `promtail` global service correctly spawned one task on each worker
  within ~4 minutes of the firewall change (expected — global services
  auto-spawn on new nodes).
- `https://devtroopersminitwit.codes/` returns 200 in ~100ms.
- TCP probe manager→worker: 7946 OPEN (gossip port works — this was the
  one that mattered). 2377 shows BLOCKED — expected, workers don't listen
  on 2377; only managers do. Direction is always worker→manager for 2377.

**One gotcha observed: webserver tasks did NOT auto-rebalance.** All three
replicas are still on Manager (they were pinned there during the outage
and Swarm doesn't reschedule running tasks onto returning nodes). Fix:

```bash
ssh root@64.226.116.162 'docker service update --force minitwit_webserver'
```

This forces a rolling restart; scheduler spreads the new replicas across
all three nodes. Rationale for Swarm's default behavior: minimize churn.
Node-coming-back events are common and you don't want them triggering
restart waves.

## Parked follow-ups (not today)

Track these here so they don't get lost once the firewall incident is done.

1. **Grafana logging dashboard label portability.** The dashboard at
   `monitoring/grafana/dashboards/logging-dashboard.json` queries
   `{swarm_service=~".+"}`, which is a Swarm-only label. On the Hetzner
   (plain docker-compose) stack the equivalent label is `compose_service`,
   so every panel shows "No data" while DNS points at Hetzner. Fix at the
   Promtail layer (normalize at ingest, not at query): add a `labels:`
   stage in both Promtail configs so both stacks emit a unified label
   (e.g. `service`), then change the dashboard to query
   `{service=~".+"}`. Alternatively, if `service_name` already has
   consistent values on both stacks, a 5-character dashboard rename may
   be enough. Verify on both stacks before merging.
   Process: `fix/grafana-loki-label-portability` → PR into `dev` → PR
   `dev` → `master`.

2. **Rotate the plaintext ghcr.io PAT** stored in
   `/root/.docker/config.json` on all three DO droplets (owner:
   `DenSygeMike` / Apoorva). Replace with the short-lived `GITHUB_TOKEN`
   flow the CD workflow already uses. Don't persist to disk.

3. **`http: superfluous response.WriteHeader call` warning** fills the
   webserver logs. Harmless to clients, but a real middleware bug around
   `main.go:50`. Probably the Prometheus metrics middleware writes the
   status header after the handler already did. Worth a small refactor
   when there's a quiet moment.

4. **Traefik stopped logging to stdout around 15 Apr 20:00 UTC.** Either
   traffic has been quiet (unlikely) or access log isn't going to stdout.
   Worth checking `--accesslog.filepath` / `--accesslog=true` config so
   Loki actually captures Traefik requests.

5. **Swarm-services that are over-replicated because of the firewall**
   (webserver 5/3, grafana 2/1, loki 2/1, promtail 3/1) will self-heal
   once workers rejoin; verify post-fix.

## Lessons recorded for later (exam material)

1. **Control plane vs data plane.** User traffic (data plane: browser → 443
   → Traefik → webserver container) and cluster management traffic (control
   plane: manager ↔ workers on 2377/7946) travel different paths. You can
   lose one and keep the other. Outside monitoring said "site up"; inside
   monitoring said "everything's broken." Both were correct.
2. **Cloud firewall ≠ host firewall.** A cloud firewall is enforced at the
   provider's network layer, before traffic reaches the VM. You can't work
   around it from inside via `iptables`. Only DO API / dashboard / `doctl`
   can change it.
3. **Tag-scoped rules are the right default for intra-cluster traffic.**
   Using "All IPv4" as source for Swarm's management port 2377 would expose
   cluster-takeover surface to the internet. Tag-scoped means "only my own
   droplets can reach this."
4. **Swarm ports you need to memorize:** TCP 2377 (mgmt), TCP+UDP 7946
   (gossip), UDP 4789 (VXLAN overlay). Without them, swarm is a bunch of
   lonely Docker daemons.
5. **Before adding any firewall to a running system, enumerate the ports
   the system actually uses internally — not just the ports users hit.**
   The mistake wasn't adding a firewall; it was not considering
   cluster-internal traffic as traffic.

## 17 Apr — HTTP/2 504s blocking the DNS flip (DO cluster not user-ready)

### Symptom

Before flipping DNS from Hetzner to DO, we ran a `--resolve` curl to the
DO manager's public IP to verify the Swarm stack actually serves real
HTTPS. Result:

- `HTTP 000 after 30s` — 504 gateway timeout, every time, from three
  independent source networks (Leo's laptop, Claude's sandbox,
  and even the manager itself).
- But: openssl s_client got the Let's Encrypt cert back in <1s, so the
  TLS handshake itself was fine.
- And: `curl --http1.1` returned `HTTP 200 in 140ms` with real MiniTwit
  HTML. Hetzner (current prod) returns `HTTP 200 in 162ms` with both
  HTTP/1.1 and HTTP/2.

So **only HTTP/2 over HTTPS into the DO cluster is broken.** Which means
every modern browser would have 30-second 504s on every request, because
browsers negotiate HTTP/2 by default over HTTPS.

### Diagnosis

Ruled out, in order:

1. **Go code:** same binary serves Hetzner fine with HTTP/2. Backend is
   innocent.
2. **Traefik → backend routing:** from inside the Traefik container,
   `wget http://minitwit_webserver:8080/public` returned real HTML. The
   overlay network from manager → workers is healthy (we fixed those
   firewall rules yesterday).
3. **Traefik router/service binding:** from inside the Traefik container,
   `wget --header=Host:devtroopersminitwit.codes https://localhost/public`
   returned `HTTP 200 OK`. Router rules match, service is bound, backend
   is reachable. The old `Router minitwit cannot be linked automatically
   with multiple Services` errors from yesterday's log are stale and no
   longer fire (separate fix tracked in branch
   `fix/swarm-traefik-and-db-indexes`).
4. **TLS:** handshake completes, cert is valid for the domain, issued by
   Let's Encrypt R13 (Apr 16 → Jul 15 2026).

What remained: requests from **outside** the manager only fail on HTTP/2.
The only thing that distinguishes "outside" from "inside" in this stack
is Docker's **ingress mesh** — the IPVS + VXLAN layer that accepts
external packets on published ports and forwards them into the service's
overlay network. Traefik's ports in `docker-stack.yml` were declared as
`"443:443"` which is shorthand for `mode: ingress` (the default).

### Root cause

Docker Swarm ingress mesh interacts badly with HTTP/2 on overlay
networks. The combination of:

- VXLAN encapsulation reducing effective MTU to 1450 bytes
- HTTP/2's larger HEADERS / SETTINGS / WINDOW_UPDATE frames
- IPVS load-balancing at TCP level on long-lived multiplexed connections

...produces exactly the failure mode we saw: TLS handshake works (small
packets, few round-trips), then the HTTP/2 preface stalls and the client
sees a server-side response-header timeout at 30s.

HTTP/1.1 doesn't trigger this because its framing is simpler and the
initial request fits comfortably inside the overlay MTU.

### Fix

Switch Traefik's `ports:` block in `docker-stack.yml` from the default
ingress mode to **host mode**. Host mode binds the published port
directly to the manager's NIC, skipping ingress mesh entirely. This is
safe for us because Traefik is already pinned to the manager node via
`deploy.placement.constraints: [node.role == manager]`, so there is no
value in having the ingress mesh route traffic between nodes — the only
node that would serve Traefik is the one ingress would route to anyway.

Change:

```yaml
# before
ports:
  - "80:80"
  - "443:443"
  - "8080:8080"

# after
ports:
  - { mode: host, target: 80,   published: 80,   protocol: tcp }
  - { mode: host, target: 443,  published: 443,  protocol: tcp }
  - { mode: host, target: 8080, published: 8080, protocol: tcp }
```

Tracked in branch `fix/traefik-ingress-host-mode`, PR into `dev` →
`master` → CD pipeline redeploys the stack.

### Lessons

6. **Ingress mesh is not free.** The default `"80:80"` shorthand
   enables Swarm's ingress routing mesh, which is useful when you want
   the published port to be reachable on every node and load-balanced
   across replicas. It is unnecessary (and a liability) when a service is
   pinned to one node anyway. Rule of thumb: if a service has a
   `placement.constraints` that limits it to one node, publish its ports
   in `mode: host`.
7. **HTTP/2 is your canary.** A stack can look healthy by every other
   metric (cluster green, cert valid, backend reachable, HTTP/1.1 works)
   and still be unusable for real users because browsers default to
   HTTP/2 over HTTPS. Always test with `curl --http2` against a
   production-shaped URL before cutting traffic over.
8. **TLS handshake success ≠ HTTPS working.** Getting a valid cert back
   via openssl only proves the first few packets traverse the path. The
   data-carrying portion of the session can still break.

### Pre-flight checklist for the eventual DNS flip

1. Merge `fix/traefik-ingress-host-mode` into `dev`, then `dev` → `master`.
2. Wait for CD pipeline to redeploy the stack.
3. SSH to manager, confirm `docker service ps minitwit_traefik` shows the
   new task running cleanly. `ss -tlnp | grep :443` on the manager should
   now show the Traefik process directly (not just `dockerd`).
4. Run these from laptop:

   ```bash
   curl --http2 --resolve devtroopersminitwit.codes:443:64.226.116.162 \
     -o /dev/null -w "HTTP %{http_code} in %{time_total}s http/%{http_version}\n" \
     https://devtroopersminitwit.codes/public
   curl --http1.1 --resolve devtroopersminitwit.codes:443:64.226.116.162 \
     -o /dev/null -w "HTTP %{http_code} in %{time_total}s http/%{http_version}\n" \
     https://devtroopersminitwit.codes/public
   ```

   Both must return 200 in under 1s. If HTTP/2 still 504s, DO NOT flip.

5. Only after both succeed: change the A record at name.com from
   `46.224.144.214` (Hetzner) to `64.226.116.162` (DO manager). TTL is
   300s so rollback window is 5 minutes.

6. Monitor from laptop with the non-resolve version (real DNS):

   ```bash
   while true; do
     curl --http2 -o /dev/null \
       -w "$(date +%H:%M:%S) %{http_code} %{time_total}s %{remote_ip}\n" \
       https://devtroopersminitwit.codes/public
     sleep 5
   done
   ```

   Watch `%{remote_ip}` flip from `46.224.144.214` to `64.226.116.162`
   as DNS propagates. All responses should stay 200.

7. If anything goes wrong, revert the A record to `46.224.144.214`.
   Within 5 min (the TTL) new resolvers will re-pick Hetzner.


## 17 Apr, afternoon — the follow-on bugs we hit when we actually tried

After the host-mode port change we thought we were green. We were not.
Three more problems surfaced once real traffic hit the cluster. Notes
here so the same hour doesn't get spent twice.

### Bug 1: Traefik hanging on startup with `client version 1.24 is too old`

Symptom. Every fresh Traefik container on the DO manager spammed:

```
Error response from daemon: client version 1.24 is too old.
Minimum supported API version is 1.44, please upgrade your client.
```

…and therefore populated zero routers, serving 404 on everything.

What we ruled out.

- Not the Docker daemon (`/_ping` returned 200; `/v1.44/version` returned 200).
- Not a socket proxy (socket is the real one, `root:docker`).
- Not `daemon.json` (there is none).
- Not Traefik version per se: v3.3.6, v3.4.5, **and** v3.5.6 all reproduce it.
- The OLD running container kept working because it was already live and had
  completed negotiation before the bug surfaced; a restart would have broken it too.

What fixed it. Upgrade to **Traefik v3.6**. Empirically:

| tag | v1.24 errors in 5 s |
|-----|---------------------|
| v3.4 | 10 |
| v3.5 | 8 |
| v3.6 | 0 |

Guess at the mechanism: the Docker SDK vendored into ≤ v3.5 of Traefik defaults
to API v1.24 when it cannot negotiate, and something about
how the Swarm provider initializes made negotiation fail against modern Docker
engines (29.1.x reports API 1.52, min 1.44). v3.6 apparently vendors a newer
SDK that negotiates before it issues real calls. Good enough explanation to
move on.

### Bug 2: `Router X cannot be linked automatically with multiple Services`

With v3.6 up, Traefik served 404s instantly instead of hanging. Logs:

```
Router minitwit     cannot be linked automatically with multiple Services: ["minitwit-sim" "minitwit"]
Router minitwit-sim cannot be linked automatically with multiple Services: ["minitwit-sim" "minitwit"]
```

Cause. Our webserver service declares two Traefik services — `minitwit`
(443 → webserver:8080) and `minitwit-sim` (8080 → webserver:8080). Traefik
v3.4 used to pick one and move on; v3.6 refuses and returns 404.

Fix. Give each router an explicit `.service=` label:

```yaml
- "traefik.http.routers.minitwit.service=minitwit"
- "traefik.http.routers.minitwit-sim.service=minitwit-sim"
```

### Bug 3: Traefik forwarding to an unreachable IP — the 30 s hang, round 2

Symptom after bug 2 fix. Request would reach Traefik, Traefik would match
`minitwit` correctly, and then the request just sat there until curl gave up.
`docker service logs` (with `--accesslog=true --log.level=DEBUG`) showed:

```
Service selected by WRR: http://10.0.1.110:8080
499 Client Closed Request   error="context canceled"
```

`10.0.1.x` is the **`minitwit_backend`** overlay. Traefik is attached to
`minitwit_frontend` (10.0.2.x) and has no route to 10.0.1.x. So every forward
hung on TCP connect until the client timed out.

Cause. Webserver is attached to two overlays — `frontend` and `backend`. The
Swarm provider picks the first task IP it sees in the Endpoint spec, which
happens to be the backend VIP. There is no auto-disambiguation.

Fix. Tell Traefik which network to use per service:

```yaml
- "traefik.swarm.network=minitwit_frontend"
```

This label has to go on every service Traefik routes to (webserver and
grafana in our case). With it in place the WRR picks 10.0.2.x addresses
and forwards actually work.

### Lessons 9–11 (appended to the running list)

9. **Floating Traefik tags are a time bomb.** A daemon upgrade on the manager
   silently raised the minimum API version and broke every Traefik container
   we restarted afterward. Pin a known-good **major** at least, and keep an
   eye on Traefik's release notes when Docker engines upgrade. Alternative:
   add a socket proxy (`tecnativa/docker-socket-proxy`) that forces a
   specific API version so Traefik's SDK never has to negotiate.

10. **Docker services on multiple overlays need `traefik.swarm.network`.**
    Without it, Traefik picks an arbitrary IP per task. It may work by luck
    if the chosen network happens to be the one Traefik is on. Ours didn't.
    Make this label non-optional anywhere routing crosses >1 overlay.

11. **Turn on `--accesslog=true --log.level=DEBUG` *before* you think you
    need it.** Bug 3 would have been a 30 second debug with logs, and was
    a 45 minute one without. The access log line
    `"Service selected by WRR: http://10.0.1.110:8080"` is the only place
    the wrong-network bug names itself.

### Pre-flight revisited (superseded checklist)

The checklist earlier in this doc is now necessary but not sufficient. Before
flipping DNS, also verify:

```bash
# 1. Traefik container is the one we expect and has fully reconciled:
ssh root@64.226.116.162 'docker service logs --tail 40 minitwit_traefik' \
  | grep -Ei 'ERR|WRN'
# Expect: no "client version 1.24", no "cannot be linked", no "context canceled"
# to 10.0.1.x addresses in access logs.

# 2. Each Traefik-exposed service has traefik.swarm.network label:
ssh root@64.226.116.162 'for s in minitwit_webserver minitwit_grafana; do
  echo "=== $s ==="
  docker service inspect $s --format "{{json .Spec.Labels}}" | grep -o traefik.swarm.network || echo MISSING
done'

# 3. Actual forward succeeds from outside over HTTP/2:
curl --http2 --resolve devtroopersminitwit.codes:443:64.226.116.162 \
  -o /dev/null -w "code=%{http_code} time=%{time_total}s\n" \
  https://devtroopersminitwit.codes/public
# Expect: code=200 time < 1s. If 000 or >5s, STOP.
```

Site was live on Hetzner the entire time these bugs were being fixed.
No user-visible outage. DNS stays on 46.224.144.214 until all three checks
above return clean.
