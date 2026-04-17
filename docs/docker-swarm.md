# Docker Swarm Migration

## What we did

### Traefik Reverse Proxy (on Hetzner)

We added Traefik as a reverse proxy to our existing single-server Docker Compose setup on Hetzner:

- **Traefik v3.4** handles all incoming traffic on ports 80 (HTTP) and 443 (HTTPS)
- Automatic TLS certificates from Let's Encrypt for `devtroopersminitwit.codes` and `grafana.devtroopersminitwit.codes`
- HTTP automatically redirects to HTTPS
- Port 8080 remains open for backward compatibility with the course simulator

Previously, the Go app was directly exposed on port 8080 with no encryption. Now all traffic goes through Traefik with TLS, while the simulator can still reach the API on `:8080`.

### Docker Swarm Setup (on DigitalOcean)

We created a 3-node Docker Swarm cluster on DigitalOcean to provide high availability and horizontal scaling:

**Nodes:**
- Manager: 64.226.116.162
- Worker 1: 206.189.59.60
- Worker 2: 134.122.90.176

**Key changes to support Swarm:**

1. **`docker-stack.yml`** — Swarm-compatible stack file (separate from `docker-compose.yml` which stays for local dev). Differences from compose:
   - No `build:` (Swarm uses pre-built images from GHCR)
   - No `depends_on:` (not supported in Swarm)
   - `deploy:` blocks with replica counts, rolling update config, placement constraints
   - Overlay networks (`frontend`, `backend`) for cross-node communication
   - Docker Swarm secrets instead of `.env` files

2. **Webserver: 3 replicas** with rolling updates (`parallelism: 1, order: start-first`). Traefik load-balances across all replicas. Session handling works without sticky sessions because we use `gorilla/sessions` CookieStore (session data is in the client cookie).

3. **Swarm secrets** — `DATABASE_URL`, `SECRET_KEY`, `DISCORD_WEBHOOK_URL` are stored as Docker secrets (mounted at `/run/secrets/<name>`). The Go app reads these via a `getSecretOrEnv()` helper that checks for a secret file first, then falls back to environment variables. This keeps local dev and CI working unchanged.

4. **Prometheus** — Changed from `static_configs` to `dns_sd_configs` with `tasks.webserver` to scrape all individual replicas (not just the Swarm VIP).

5. **Promtail** — Runs as a `global` service (one per node) to collect Docker container logs from every node. Relabel config updated from Compose labels to Swarm labels (`com.docker.swarm.service.name`).

6. **Grafana** — Added an entrypoint script to export Swarm secrets as environment variables (needed for the Discord webhook URL in alerting).

7. **CD pipeline** — Builds and pushes 3 Docker images (app, prometheus, grafana) to GHCR. Deploys to both DigitalOcean Swarm and Hetzner (legacy) in parallel.

## Service distribution across nodes

| Service | Manager | Worker 1 | Worker 2 |
|---------|---------|----------|----------|
| Traefik | x | | |
| Webserver | x | x | x |
| Prometheus | x | | |
| Grafana | x (likely) | | |
| Loki | placed by Swarm | | |
| Promtail | x | x | x |

- **Traefik** and **Prometheus** are constrained to the manager node
- **Promtail** runs on every node (global mode) to collect local container logs
- **Webserver** replicas are distributed across all 3 nodes by Swarm
- **Grafana** and **Loki** are placed by Swarm (no constraint, typically on manager)

## What to do next

### 1. Set up the DigitalOcean nodes

Install Docker on all 3 droplets:
```bash
apt-get update && apt-get install -y docker.io && systemctl enable docker && systemctl start docker
```

Initialize Swarm on the manager (64.226.116.162):
```bash
docker swarm init --advertise-addr 64.226.116.162
```

Join both workers using the token from the init command:
```bash
docker swarm join --token <TOKEN> 64.226.116.162:2377
```

### 2. Create Docker secrets on the manager
```bash
echo "<DATABASE_URL>" | docker secret create database_url -
echo "<SECRET_KEY>" | docker secret create secret_key -
echo "<DISCORD_WEBHOOK_URL>" | docker secret create discord_webhook_url -
```

### 3. Clone the repo on the manager
```bash
apt-get install -y git
git clone https://github.com/DevTroopers-ITU/itu-minitwit.git /root/itu-minitwit
```

### 4. Add GitHub Actions secret

Add `DO_SSH_PRIVATE_KEY` to GitHub repo secrets (Settings > Secrets) — the SSH private key for root access to the DO manager node.

### 5. Merge PR to master

This triggers CD which deploys to both DO Swarm and Hetzner.

### 6. Update DNS

Once the Swarm is verified working, update the A records at Name.com to point to the manager IP (64.226.116.162).

### 7. Decommission Hetzner

Once DNS is pointing to DO and the simulator URL is updated:
1. Remove the Hetzner deploy step from `cd.yml`
2. Shut down the Hetzner server

## Host-mode ports vs Ingress mesh

### What are they?

When Traefik publishes ports (80, 443, 8080), Docker Swarm has two ways to handle incoming traffic:

**Ingress mesh (Swarm default):** Every node in the swarm listens on the published ports. When a request hits any node, Swarm's built-in load balancer (IPVS) tunnels it through a virtual network (VXLAN) to whichever node is actually running Traefik.

```
Client → Worker-2:443 → IPVS → VXLAN tunnel → Manager:Traefik
```

**Host-mode (what we use):** Only the node running Traefik listens on those ports. Traffic goes straight to Traefik, no tunnel, no extra hops.

```
Client → Manager:443 → Traefik (direct, no tunnel)
```

### Why we chose host-mode

1. **DNS only points to the manager.** Our A-record is `64.226.116.162` (the manager). No client ever hits Worker-1 or Worker-2 directly. So having all nodes listen on port 443 via ingress mesh is pointless — the traffic always arrives at the manager anyway.

2. **Ingress mesh caused 30-second timeouts.** HTTP/2 connections through the VXLAN overlay hit an MTU (packet size) issue that made requests hang for 30 seconds before failing. Host-mode bypasses the overlay entirely and fixes this.

### The tradeoff: What happens if the manager dies?

**With ingress mesh:**
- DNS points to all 3 nodes → clients can still reach Worker-1 or Worker-2
- Swarm elects a new leader and restarts Traefik on another node
- Ingress mesh routes traffic to the new Traefik automatically
- Result: ~30 seconds of downtime, then self-heals

**With host-mode (our setup):**
- DNS points only to the manager → clients cannot reach the dead IP
- Swarm elects a new leader and restarts Traefik on the new manager node
- Traefik is running and healthy on the new node
- BUT DNS still points to the old (dead) IP
- Result: site is down until someone manually updates DNS to the new manager's IP

### Why this tradeoff is fine for us

- The manager rarely dies
- Updating a DNS record at Name.com takes 2 minutes
- TTL is 300 seconds, so DNS propagates in under 5 minutes
- We avoid the VXLAN/MTU bugs that made the site unreliable

### When would ingress mesh be better?

If you have a setup where DNS points to multiple nodes (round-robin A-records), or if you need automatic failover without touching DNS. This is more relevant for larger-scale deployments or services with strict uptime SLAs.

### The code change

```yaml
# BEFORE: Ingress mesh (default)
ports:
  - "80:80"
  - "443:443"
  - "8080:8080"

# AFTER: Host-mode (direct binding)
ports:
  - { mode: host, target: 80,   published: 80,   protocol: tcp }
  - { mode: host, target: 443,  published: 443,  protocol: tcp }
  - { mode: host, target: 8080, published: 8080, protocol: tcp }
```

Same ports, just with `mode: host` added. That tells Docker: "bind directly to the machine's network interface instead of going through Swarm's internal load balancer."

## Decisions and rationale

- **Why Docker Swarm over Kubernetes?** Swarm is simpler, built into Docker, and sufficient for our scale. The course exercise specifically covers Swarm.
- **Why Traefik over Nginx/Caddy?** Traefik has native Docker/Swarm integration — it auto-discovers services via labels and handles Let's Encrypt automatically. No config reloads needed when replicas change.
- **Why separate docker-stack.yml?** `docker stack deploy` doesn't support `build:`, `depends_on:`, or `.env` files. Keeping `docker-compose.yml` for local dev avoids breaking the development workflow.
- **Why DigitalOcean instead of more Hetzner nodes?** A colleague had already provisioned the DO droplets.
- **Why keep Hetzner running?** Zero-downtime migration — the simulator and existing users continue to work while we set up and verify the Swarm.
