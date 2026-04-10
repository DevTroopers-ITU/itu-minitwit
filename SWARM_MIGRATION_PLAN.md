# Docker Swarm Migration Plan

## Current State

| Component | Details |
|-----------|---------|
| **Compute** | 1x Hetzner Cloud VM (`cx23`, 2 vCPU, 4 GB RAM, Ubuntu 22.04) |
| **Database** | DigitalOcean Managed PostgreSQL (external, accessed via `DATABASE_URL`) |
| **Registry** | GitHub Container Registry (`ghcr.io/devtroopers-itu/minitwit`) |
| **Provisioning** | Vagrant with `vagrant-hetznercloud` plugin |
| **Deployment** | GitHub Actions SSH into single server, `docker compose up -d` |
| **Monitoring** | Prometheus + Grafana + Loki + Promtail (all on same VM) |
| **Alerting** | Prometheus rules -> Grafana -> Discord webhook |

### Current Architecture

```
                Internet
                   │
                   ▼
        ┌──────────────────┐
        │  Hetzner VM      │
        │  46.224.144.214  │
        │                  │
        │  ┌────────────┐  │          ┌──────────────────┐
        │  │ webserver   │──┼─────────▶ DigitalOcean     │
        │  │ :8080       │  │          │ Managed Postgres │
        │  └────────────┘  │          └──────────────────┘
        │  ┌────────────┐  │
        │  │ prometheus  │  │
        │  │ :9090       │  │
        │  └────────────┘  │
        │  ┌────────────┐  │
        │  │ grafana     │  │
        │  │ :3000       │  │
        │  └────────────┘  │
        │  ┌────────────┐  │
        │  │ loki :3100  │  │
        │  └────────────┘  │
        │  ┌────────────┐  │
        │  │ promtail    │  │
        │  └────────────┘  │
        └──────────────────┘
```

---

## Target Architecture

```
                Internet
                   │
                   ▼
        ┌──────────────────┐
        │  manager         │
        │  (Swarm Leader)  │
        │                  │
        │  ┌────────────┐  │
        │  │ traefik     │──┼──── Reverse proxy + TLS termination
        │  │ :80 / :443  │  │     Routes to webserver replicas
        │  └────────────┘  │
        │  ┌────────────┐  │
        │  │ prometheus  │  │
        │  │ :9090       │  │
        │  └────────────┘  │
        │  ┌────────────┐  │
        │  │ grafana     │  │
        │  │ :3000       │  │
        │  └────────────┘  │
        │  ┌────────────┐  │          ┌──────────────────┐
        │  │ loki :3100  │  │          │ DigitalOcean     │
        │  └────────────┘  │          │ Managed Postgres  │
        │  ┌────────────┐  │          └─────────▲────────┘
        │  │ promtail    │  │                    │
        │  └────────────┘  │                    │
        └──────────────────┘                    │
                                                │
        ┌──────────────────┐                    │
        │  worker-1        │                    │
        │                  │                    │
        │  ┌────────────┐  │                    │
        │  │ webserver   │──┼────────────────────
        │  │ (replica)   │  │
        │  └────────────┘  │
        │  ┌────────────┐  │
        │  │ promtail    │  │
        │  └────────────┘  │
        └──────────────────┘

        ┌──────────────────┐
        │  worker-2        │
        │                  │
        │  ┌────────────┐  │
        │  │ webserver   │──┼────────────────────
        │  │ (replica)   │  │
        │  └────────────┘  │
        │  ┌────────────┐  │
        │  │ promtail    │  │
        │  └────────────┘  │
        └──────────────────┘
```

### Service Placement Summary

| Service | Mode | Placement | Why |
|---------|------|-----------|-----|
| **traefik** | 1 replica | manager only | Needs Docker socket access, single entry point |
| **webserver** | 3 replicas | any node | Stateless, scales horizontally |
| **prometheus** | 1 replica | manager only | Stateful (TSDB), scrapes all services |
| **grafana** | 1 replica | manager only | Stateful (dashboards), needs prometheus + loki |
| **loki** | 1 replica | manager only | Stateful (log storage) |
| **promtail** | global (1 per node) | every node | Must collect logs from local Docker daemon |

---

## Migration Steps

### Step 1: Push all images to the registry

**Why:** Docker Swarm cannot run `build:` -- it can only pull pre-built images. Currently, only the webserver image is pushed to `ghcr.io`. Prometheus and Grafana use local `build:` directives.

**What to do:**

1. Add image names and push steps for prometheus and grafana in `cd.yml`:

```yaml
# In .github/workflows/cd.yml, add to the "Build and push" step:

- name: Build and push Docker images
  run: |
    docker build -t ghcr.io/devtroopers-itu/minitwit:latest .
    docker build -t ghcr.io/devtroopers-itu/minitwit-prometheus:latest ./monitoring/prometheus
    docker build -t ghcr.io/devtroopers-itu/minitwit-grafana:latest ./monitoring/grafana
    docker push ghcr.io/devtroopers-itu/minitwit:latest
    docker push ghcr.io/devtroopers-itu/minitwit-prometheus:latest
    docker push ghcr.io/devtroopers-itu/minitwit-grafana:latest
```

2. Update `docker-compose.yml` to reference the new image names:

```yaml
prometheus:
  image: ghcr.io/devtroopers-itu/minitwit-prometheus:latest
  # remove: build: ./monitoring/prometheus

grafana:
  image: ghcr.io/devtroopers-itu/minitwit-grafana:latest
  # remove: build: ./monitoring/grafana
```

**Files changed:**
- `.github/workflows/cd.yml`
- `docker-compose.yml`

---

### Step 2: Bake config files into images (eliminate bind mounts)

**Why:** Bind mounts like `./monitoring/loki/loki-config.yml:/etc/loki/...` require the file to exist on the host filesystem. In a Swarm, containers can land on any node -- the file won't be there. You have two options: Docker configs or baking into images. Baking is simpler and matches what you already do for prometheus and grafana.

**What to do:**

1. Create `monitoring/loki/Dockerfile`:

```dockerfile
FROM grafana/loki:3.6.7
COPY loki-config.yml /etc/loki/local-config.yaml
```

2. Create `monitoring/promtail/Dockerfile`:

```dockerfile
FROM grafana/promtail:3.6.0
COPY promtail-config.yml /etc/promtail/config.yml
```

3. Build and push these images too (add to `cd.yml`):

```yaml
docker build -t ghcr.io/devtroopers-itu/minitwit-loki:latest ./monitoring/loki
docker build -t ghcr.io/devtroopers-itu/minitwit-promtail:latest ./monitoring/promtail
docker push ghcr.io/devtroopers-itu/minitwit-loki:latest
docker push ghcr.io/devtroopers-itu/minitwit-promtail:latest
```

4. Update `docker-compose.yml` -- replace bare images with your custom ones and remove bind mounts for config files:

```yaml
loki:
  image: ghcr.io/devtroopers-itu/minitwit-loki:latest
  # remove: volumes: ./monitoring/loki/loki-config.yml:...

promtail:
  image: ghcr.io/devtroopers-itu/minitwit-promtail:latest
  # remove: volumes: ./monitoring/promtail/promtail-config.yml:...
  # KEEP: /var/run/docker.sock and /var/lib/docker/containers (these are host paths, not repo files)
```

**Files changed:**
- New: `monitoring/loki/Dockerfile`
- New: `monitoring/promtail/Dockerfile`
- `.github/workflows/cd.yml`
- `docker-compose.yml`

---

### Step 3: Provision 3 Hetzner VMs

**Why:** A Swarm needs multiple nodes. 3 nodes gives you fault tolerance (quorum requires majority of managers alive).

**What to do:**

Create 3 servers on Hetzner Cloud. You can do this via the Hetzner web console, `hcloud` CLI, or by updating the Vagrantfile.

**Option A: Using `hcloud` CLI (recommended for simplicity):**

```bash
# Install hcloud CLI, then:
hcloud server create --name manager   --type cx22 --image ubuntu-22.04 --location nbg1 --ssh-key <your-key>
hcloud server create --name worker-1  --type cx22 --image ubuntu-22.04 --location nbg1 --ssh-key <your-key>
hcloud server create --name worker-2  --type cx22 --image ubuntu-22.04 --location nbg1 --ssh-key <your-key>
```

> **Note:** `cx22` (2 vCPU, 4 GB) is sufficient for each node. Total cost ~15 EUR/month.

**Option B: Update the Vagrantfile for multi-machine:**

```ruby
Vagrant.configure("2") do |config|
  config.vm.box = "dummy"
  config.ssh.private_key_path = ENV["SSH_KEY_PATH"] || "~/.ssh/id_rsa"
  config.ssh.username = "root"

  nodes = {
    "manager"  => "cx22",
    "worker-1" => "cx22",
    "worker-2" => "cx22",
  }

  nodes.each do |name, server_type|
    config.vm.define name do |node|
      node.vm.provider :hetznercloud do |hcloud|
        hcloud.token = ENV["HCLOUD_TOKEN"]
        hcloud.image = "ubuntu-22.04"
        hcloud.location = "nbg1"
        hcloud.server_type = server_type
        hcloud.ssh_keys = ["leo", "haakon-2", "apoorva", "peter-juul", "phkt@archlinux"]
      end

      node.vm.provision "shell", inline: <<-SHELL
        apt-get update -qq
        apt-get install -y -qq docker.io > /dev/null 2>&1
        systemctl enable docker
        systemctl start docker
      SHELL
    end
  end
end
```

**On each server, install Docker and log in to ghcr.io:**

```bash
apt-get update && apt-get install -y docker.io
echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin
```

**Files changed:**
- `Vagrantfile` (if using Option B)

---

### Step 4: Initialize the Swarm

**Why:** This is the core step -- turns your standalone Docker hosts into a cluster.

**What to do:**

1. **SSH into the manager node** and initialize:

```bash
docker swarm init --advertise-addr <MANAGER_PRIVATE_IP>
```

This prints a join token. Save it.

2. **SSH into each worker** and join:

```bash
docker swarm join --token <TOKEN> <MANAGER_PRIVATE_IP>:2377
```

3. **Verify on manager:**

```bash
docker node ls
```

Expected output:
```
ID          HOSTNAME    STATUS    AVAILABILITY    MANAGER STATUS
abc123 *    manager     Ready     Active          Leader
def456      worker-1    Ready     Active
ghi789      worker-2    Ready     Active
```

**Firewall rules required** (Hetzner firewall or `ufw`):

| Port | Protocol | Purpose |
|------|----------|---------|
| 2377 | TCP | Swarm cluster management |
| 7946 | TCP + UDP | Node-to-node communication |
| 4789 | UDP | Overlay network traffic (VXLAN) |

```bash
# On ALL nodes:
ufw allow 2377/tcp
ufw allow 7946/tcp
ufw allow 7946/udp
ufw allow 4789/udp
```

---

### Step 5: Create the Swarm stack file

**Why:** The current `docker-compose.yml` uses Compose-only features (`build:`, `depends_on:`, `restart:`, `mem_limit:`). These are ignored or invalid in Swarm mode. We need a Swarm-compatible version.

**What to do:**

Create `docker-compose.swarm.yml` (keep the original for local dev/testing):

```yaml
# docker-compose.swarm.yml -- Swarm stack deployment file

services:
  traefik:
    image: traefik:v3.6
    ports:
      - "80:80"
      - "443:443"
    deploy:
      replicas: 1
      placement:
        constraints:
          - node.role == manager
      restart_policy:
        condition: on-failure
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - letsencrypt-data:/letsencrypt
    command:
      - --providers.swarm.endpoint=unix:///var/run/docker.sock
      - --providers.swarm.exposedByDefault=false
      - --entrypoints.web.address=:80
      - --entrypoints.websecure.address=:443
      - --entrypoints.web.http.redirections.entryPoint.to=websecure
      - --certificatesresolvers.myresolver.acme.tlschallenge=true
      - --certificatesresolvers.myresolver.acme.email=your-team@itu.dk
      - --certificatesresolvers.myresolver.acme.storage=/letsencrypt/acme.json
    networks:
      - minitwit-net

  webserver:
    image: ghcr.io/devtroopers-itu/minitwit:latest
    deploy:
      replicas: 3
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
      update_config:
        parallelism: 1
        delay: 10s
        failure_action: rollback
      rollback_config:
        parallelism: 1
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.minitwit.rule=Host(`your-domain.dev`)"
        - "traefik.http.routers.minitwit.entrypoints=websecure"
        - "traefik.http.routers.minitwit.tls.certresolver=myresolver"
        - "traefik.http.services.minitwit.loadbalancer.server.port=8080"
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080"]
      interval: 30s
      timeout: 10s
      retries: 3
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - SECRET_KEY=${SECRET_KEY}
    networks:
      - minitwit-net

  prometheus:
    image: ghcr.io/devtroopers-itu/minitwit-prometheus:latest
    deploy:
      replicas: 1
      placement:
        constraints:
          - node.role == manager
      restart_policy:
        condition: on-failure
      resources:
        limits:
          memory: 1G
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.retention.time=15d'
    volumes:
      - prometheus-data:/prometheus
    networks:
      - minitwit-net

  grafana:
    image: ghcr.io/devtroopers-itu/minitwit-grafana:latest
    deploy:
      replicas: 1
      placement:
        constraints:
          - node.role == manager
      restart_policy:
        condition: on-failure
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.grafana.rule=Host(`your-domain.dev`) && PathPrefix(`/grafana`)"
        - "traefik.http.routers.grafana.entrypoints=websecure"
        - "traefik.http.routers.grafana.tls.certresolver=myresolver"
        - "traefik.http.services.grafana.loadbalancer.server.port=3000"
    environment:
      - DISCORD_WEBHOOK_URL=${DISCORD_WEBHOOK_URL}
      - GF_SERVER_ROOT_URL=https://your-domain.dev/grafana
      - GF_SERVER_SERVE_FROM_SUB_PATH=true
    networks:
      - minitwit-net

  loki:
    image: ghcr.io/devtroopers-itu/minitwit-loki:latest
    deploy:
      replicas: 1
      placement:
        constraints:
          - node.role == manager
      restart_policy:
        condition: on-failure
    volumes:
      - loki-data:/loki
    networks:
      - minitwit-net

  promtail:
    image: ghcr.io/devtroopers-itu/minitwit-promtail:latest
    deploy:
      mode: global
      restart_policy:
        condition: on-failure
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
    networks:
      - minitwit-net

volumes:
  prometheus-data:
  loki-data:
  letsencrypt-data:

networks:
  minitwit-net:
    driver: overlay
```

**Key differences from the original `docker-compose.yml`:**

| What changed | Why |
|--------------|-----|
| All `build:` removed | Swarm only pulls images |
| All `depends_on:` removed | Not supported in Swarm (services must handle startup order themselves) |
| `restart: always` -> `deploy.restart_policy` | Swarm manages restarts via deploy config |
| `mem_limit: 1g` -> `deploy.resources.limits.memory` | New syntax for Swarm |
| `ports: "8080:8080"` removed from webserver | Traefik handles external routing now |
| Traefik service added | Load balancer + TLS termination |
| `mode: global` on promtail | Runs one instance per node (collects local logs) |
| Overlay network added | Enables cross-node communication |
| `update_config` + `rollback_config` on webserver | Zero-downtime rolling deployments |

**Files changed:**
- New: `docker-compose.swarm.yml`

---

### Step 6: Update Prometheus scrape config for multiple webserver replicas

**Why:** Currently, `prometheus.yml` scrapes a single target `webserver:8080`. With 3 replicas behind Swarm's internal DNS, this would round-robin and scrape a random replica each time. You need to scrape all of them.

**What to do:**

Update `monitoring/prometheus/prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'prometheus'
    scrape_interval: 15s
    static_configs:
      - targets: ['prometheus:9090']

  - job_name: 'itu-minitwit-app'
    scrape_interval: 15s
    dns_sd_configs:
      - names:
          - 'tasks.webserver'    # Swarm DNS: resolves to ALL replica IPs
        type: 'A'
        port: 8080
    metric_relabel_configs:
      - source_labels: [__name__]
        regex: 'minitwit_.*|up'
        action: keep
```

> **Key change:** `static_configs` -> `dns_sd_configs` using `tasks.webserver`. In Swarm, `tasks.<service>` resolves to the IP of every individual replica, not the load-balanced VIP.

**Files changed:**
- `monitoring/prometheus/prometheus.yml`

---

### Step 7: Use Docker secrets for sensitive env vars

**Why:** Currently, secrets are stored in a `.env` file on disk. Docker Swarm has built-in secret management that encrypts secrets at rest and only exposes them to services that need them.

**What to do:**

1. **Create secrets on the manager:**

```bash
echo "postgres://user:pass@host:5432/db" | docker secret create database_url -
echo "your-secret-key"                   | docker secret create secret_key -
echo "https://discord.com/api/..."       | docker secret create discord_webhook_url -
```

2. **Reference secrets in the stack file** (update `docker-compose.swarm.yml`):

```yaml
services:
  webserver:
    # ...
    secrets:
      - database_url
      - secret_key
    environment:
      - DATABASE_URL_FILE=/run/secrets/database_url
      - SECRET_KEY_FILE=/run/secrets/secret_key

  grafana:
    # ...
    secrets:
      - discord_webhook_url

secrets:
  database_url:
    external: true
  secret_key:
    external: true
  discord_webhook_url:
    external: true
```

> **Note:** This requires your Go app to read `DATABASE_URL_FILE` (path to a file containing the secret) instead of `DATABASE_URL` (the secret itself). Add a small helper in your Go code:
>
> ```go
> func getEnv(key string) string {
>     if filePath := os.Getenv(key + "_FILE"); filePath != "" {
>         data, err := os.ReadFile(filePath)
>         if err == nil {
>             return strings.TrimSpace(string(data))
>         }
>     }
>     return os.Getenv(key)
> }
> ```

**This step is optional.** You can continue using `.env` files by placing them on the manager and using `env_file:` in your stack. But secrets are the proper Swarm way.

**Files changed:**
- `docker-compose.swarm.yml` (if adopting secrets)
- Go application code (small env helper function)

---

### Step 8: Deploy the stack

**What to do on the manager node:**

```bash
# 1. Copy the .env file to the manager (if not using Docker secrets)
scp .env root@<MANAGER_IP>:/root/.env

# 2. Copy the stack file
scp docker-compose.swarm.yml root@<MANAGER_IP>:/root/docker-compose.swarm.yml

# 3. SSH into manager
ssh root@<MANAGER_IP>

# 4. Load env vars and deploy
export $(cat /root/.env | xargs)
docker stack deploy -c docker-compose.swarm.yml minitwit

# 5. Verify everything is running
docker service ls
docker service ps minitwit_webserver
docker service logs minitwit_webserver
```

Expected output of `docker service ls`:
```
ID        NAME                   MODE        REPLICAS   IMAGE
abc123    minitwit_traefik       replicated  1/1        traefik:v3.6
def456    minitwit_webserver     replicated  3/3        ghcr.io/devtroopers-itu/minitwit:latest
ghi789    minitwit_prometheus    replicated  1/1        ghcr.io/devtroopers-itu/minitwit-prometheus:latest
jkl012    minitwit_grafana       replicated  1/1        ghcr.io/devtroopers-itu/minitwit-grafana:latest
mno345    minitwit_loki          replicated  1/1        ghcr.io/devtroopers-itu/minitwit-loki:latest
pqr678    minitwit_promtail      global      3/3        ghcr.io/devtroopers-itu/minitwit-promtail:latest
```

---

### Step 9: Update the CD pipeline

**Why:** The current pipeline SSHs into a single server and runs `docker compose up`. With Swarm, you SSH into the manager and trigger a rolling service update.

**What to do:**

Update `.github/workflows/cd.yml`:

```yaml
name: CD

on:
  push:
    branches: [master]

jobs:
  build:
    runs-on: ubuntu-22.04
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Log in to GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and push all images
        run: |
          docker build -t ghcr.io/devtroopers-itu/minitwit:latest .
          docker build -t ghcr.io/devtroopers-itu/minitwit-prometheus:latest ./monitoring/prometheus
          docker build -t ghcr.io/devtroopers-itu/minitwit-grafana:latest ./monitoring/grafana
          docker build -t ghcr.io/devtroopers-itu/minitwit-loki:latest ./monitoring/loki
          docker build -t ghcr.io/devtroopers-itu/minitwit-promtail:latest ./monitoring/promtail
          docker push ghcr.io/devtroopers-itu/minitwit:latest
          docker push ghcr.io/devtroopers-itu/minitwit-prometheus:latest
          docker push ghcr.io/devtroopers-itu/minitwit-grafana:latest
          docker push ghcr.io/devtroopers-itu/minitwit-loki:latest
          docker push ghcr.io/devtroopers-itu/minitwit-promtail:latest

  deploy:
    runs-on: ubuntu-22.04
    needs: build
    steps:
      - name: Rolling update on Swarm
        uses: appleboy/ssh-action@v1
        with:
          host: <MANAGER_IP>
          username: root
          key: ${{ secrets.SSH_PRIVATE_KEY }}
          script: |
            # Update each service to pull the latest image
            docker service update --image ghcr.io/devtroopers-itu/minitwit:latest minitwit_webserver
            docker service update --image ghcr.io/devtroopers-itu/minitwit-prometheus:latest minitwit_prometheus
            docker service update --image ghcr.io/devtroopers-itu/minitwit-grafana:latest minitwit_grafana
            docker service update --image ghcr.io/devtroopers-itu/minitwit-loki:latest minitwit_loki
            docker service update --image ghcr.io/devtroopers-itu/minitwit-promtail:latest minitwit_promtail
```

> **What happens:** `docker service update --image` triggers a **rolling update**. Swarm pulls the new image, starts new containers one at a time, health-checks them, and stops old ones. Zero downtime.

**Files changed:**
- `.github/workflows/cd.yml`

---

### Step 10: Update CI to test with the swarm file

**Why:** You should validate the swarm compose file in CI.

**What to do:**

Add a validation step to `.github/workflows/ci.yml`:

```yaml
- name: Validate swarm stack file
  run: docker stack config -c docker-compose.swarm.yml > /dev/null
```

> **Note:** The existing E2E tests in `docker-compose.test.yml` should continue using the original `docker-compose.yml` since tests run on a single CI machine, not a swarm.

**Files changed:**
- `.github/workflows/ci.yml`

---

## Migration Checklist

| # | Task | Risk | Reversible? |
|---|------|------|-------------|
| 1 | Push all images to ghcr.io | Low | Yes |
| 2 | Bake config files into images (Loki, Promtail Dockerfiles) | Low | Yes |
| 3 | Provision 3 Hetzner VMs | Low | Yes (delete them) |
| 4 | Initialize Swarm + join workers | Low | Yes (`docker swarm leave`) |
| 5 | Create `docker-compose.swarm.yml` | Low | Yes (file in repo) |
| 6 | Update Prometheus scrape config | Low | Yes (redeploy old config) |
| 7 | Docker secrets (optional) | Medium | Yes (fall back to .env) |
| 8 | Deploy the stack | **Medium** | Yes (`docker stack rm minitwit`) |
| 9 | Update CD pipeline to use service update | **Medium** | Yes (revert commit) |
| 10 | Validate swarm file in CI | Low | Yes |

---

## Rollback Plan

If something goes wrong at any step:

1. **The old VM is still running.** Don't decommission it until the Swarm is verified.
2. **DNS/IP switch is the point of no return.** Only update DNS to point to the Swarm manager after everything is confirmed working.
3. **To tear down the Swarm completely:**

```bash
# On manager:
docker stack rm minitwit        # removes all services
docker swarm leave --force      # dissolves the swarm

# On workers:
docker swarm leave
```

4. **To go back to single-server compose:**

```bash
# On the old Hetzner VM (which you kept running):
cd /root/itu-minitwit
docker compose up -d
```

---

## Suggested Execution Order

```
Week 1:  Steps 1-2 (images + Dockerfiles) -- merge to master, no disruption
Week 2:  Steps 3-4 (provision servers + init swarm) -- infrastructure only
Week 2:  Steps 5-6 (stack file + prometheus config) -- merge to master
Week 2:  Step 8 (deploy to swarm) -- run in parallel with old server
         Verify: all services healthy, monitoring works, simulator can hit API
Week 3:  Step 9 (update CD pipeline) -- switch over
         Verify: push to master triggers rolling update
Week 3:  Step 10 (CI validation)
Week 3:  Decommission old Hetzner VM
```

> **Critical rule:** Keep the old server running until you've verified the Swarm handles at least one full CD cycle (push -> build -> deploy -> verify).
