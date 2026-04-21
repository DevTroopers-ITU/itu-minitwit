# Ops CLI Reference

Quick-lookup commands for operating the Swarm cluster, debugging a service, and recovering from common issues. Keep this page short — it's a cheat-sheet, not a manual.

## Hosts

| Role | IP | Notes |
|------|----|----|
| Manager | 64.226.116.162 | Traefik, Grafana, Prometheus, Loki live here |
| Worker-1 | 134.122.90.176 | Webserver replica |
| Worker-2 | 206.189.59.60 | Webserver replica |
| Legacy | 46.224.144.214 | Old Hetzner single-server deploy |
| Domain | devtroopersminitwit.codes | A-record → Manager IP |

```bash
ssh root@64.226.116.162     # manager
ssh root@134.122.90.176     # worker-1
ssh root@206.189.59.60      # worker-2
```

All Swarm commands below must run on the **manager**.

## Fast health check

Run this first when something looks off:

```bash
ssh root@64.226.116.162 '
  echo "== NODES ==";    docker node ls
  echo "== SERVICES =="; docker service ls
  echo "== TASKS ==";    docker stack ps minitwit -f "desired-state=running" \
    --format "{{.Name}}|{{.Node}}|{{.CurrentState}}"
  echo "== /latest =="; curl -s https://devtroopersminitwit.codes/latest
'
```

All services should be `N/N`, all tasks `Running`, `/latest` should return a number that increments every few seconds (simulator is live).

## Swarm — cluster

```bash
docker node ls                                  # all nodes + leader
docker node inspect Manager --pretty            # details on one node
docker node ps Worker-1                         # what's running on a node
```

## Swarm — services

```bash
docker service ls                               # all services, replicas desired/running
docker stack ls                                 # stacks and their service count
docker stack ps minitwit                        # tasks across the whole stack
docker stack ps minitwit --no-trunc \
  -f "desired-state=running"                    # only live tasks

docker service ps minitwit_webserver            # task history for one service
docker service ps minitwit_webserver --no-trunc # full error messages on Rejected tasks
docker service inspect minitwit_webserver --pretty
docker service inspect minitwit_webserver \
  --format '{{.Spec.TaskTemplate.ContainerSpec.Image}}'   # running digest
```

Tasks you'll see in `service ps`:
- `Running` — healthy
- `Shutdown` — previous replica being replaced (normal during rolling update)
- `Rejected` — task couldn't start, check the error column (usually image pull)
- `Preparing` / `Starting` — in-flight during deploy

## Logs

```bash
# All replicas of a service, follow:
docker service logs --tail 100 --follow minitwit_webserver

# Since a time, filter errors:
docker service logs --since 10m minitwit_webserver | grep -iE "error|panic"

# Traefik access logs — every HTTP request going through prod:
docker service logs --tail 200 minitwit_traefik

# For older/persistent logs, use Loki via Grafana (see Monitoring below).
```

## Containers (when you need to go deeper than services)

Run on the node the task is placed on:

```bash
docker ps                                       # containers on THIS node
docker ps --format 'table {{.Names}}\t{{.Status}}'
docker logs <container-name> --tail 100         # container stdout/stderr
docker exec -it <container-name> sh             # shell into a container
docker stats --no-stream                        # CPU/mem per container
```

Find which node a task is on via `docker service ps <svc>` from the manager first.

## Deploy, redeploy, rollback

Normal path: push to `master` → GitHub Actions runs CD → stack redeploys. Manual controls:

```bash
# Force a fresh pull + redeploy without a commit:
docker service update --force minitwit_webserver

# Roll one service back to its previous spec (no git needed):
docker service rollback minitwit_webserver

# Full stack redeploy from the current checkout on the manager:
cd /root/itu-minitwit
docker stack deploy -c docker-stack.yml minitwit --with-registry-auth

# Scale a service up/down at runtime (not persisted to the stack file):
docker service scale minitwit_webserver=2
```

If CD reports "success" but prod doesn't reflect your changes, the manager's working tree is probably dirty (see Troubleshooting).

## Troubleshooting

**Task stuck `Rejected` with "No such image"**
The node can't pull the image digest — either GHCR credentials expired or the digest was deleted. Fix on the affected node:

```bash
ssh root@<node-ip>
echo "$GHCR_PAT" | docker login ghcr.io -u <gh-user> --password-stdin
docker pull ghcr.io/devtroopers-itu/minitwit:latest
# Then on manager:
docker service update --force minitwit_webserver
```

**CD workflow says "success" but nothing changed in prod**
The `git pull` on the manager silently failed because the working tree is dirty. Check and fix:

```bash
ssh root@64.226.116.162
cd /root/itu-minitwit
git status                              # any "modified:" files?
git stash                               # or: git checkout -- <file>
git pull origin master
docker stack deploy -c docker-stack.yml minitwit --with-registry-auth
```

**Placement not matching docker-stack.yml** (e.g. Grafana running on a worker)
Usually means the deployed spec is stale. Redeploy with the current file:

```bash
ssh root@64.226.116.162
cd /root/itu-minitwit && git pull origin master
docker stack deploy -c docker-stack.yml minitwit --with-registry-auth
# Confirm placement after the update settles:
docker service ps minitwit_grafana -f "desired-state=running"
```

**HTTP 504 from the domain**
Check Traefik is up and reachable, then that webserver replicas are healthy:

```bash
docker service ps minitwit_traefik
docker service ps minitwit_webserver -f "desired-state=running"
docker service logs --tail 50 minitwit_traefik | grep -i error
```

**Node fell out of the Swarm** (`Down` in `docker node ls`)
SSH into that node, check the daemon, rejoin if needed:

```bash
ssh root@<node-ip>
systemctl status docker
docker info | grep -i swarm
# If it left the swarm, get a fresh worker join token from manager:
ssh root@64.226.116.162 'docker swarm join-token worker'
```

## Monitoring

| Tool | URL | What for |
|------|-----|----|
| Grafana | https://grafana.devtroopersminitwit.codes/ | Dashboards + alerts |
| App | https://devtroopersminitwit.codes/ | Production UI |
| Simulator endpoint | https://devtroopersminitwit.codes/latest | Sim health, grader poll target |

Prometheus and Loki aren't directly exposed — query them through Grafana.

Quick Prometheus query from the manager (useful when Grafana is down):

```bash
ssh root@64.226.116.162 \
  'docker exec $(docker ps -qf name=minitwit_prometheus) \
   wget -qO- "http://localhost:9090/api/v1/query?query=up"' | jq .
```

## Database (DigitalOcean managed Postgres)

The connection string lives in the `database_url` Swarm secret, not in the repo.

```bash
# Read the secret file from inside a running webserver container:
ssh root@<node-running-webserver>
docker exec <container> cat /run/secrets/database_url

# Open psql from the manager using that URL:
ssh root@64.226.116.162
apt-get install -y postgresql-client    # once
psql "$(docker exec $(docker ps -qf name=minitwit_webserver) cat /run/secrets/database_url)"
```

Useful queries:

```sql
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM messages;
SELECT latest FROM sim_states WHERE id = 1;   -- simulator counter (post-PR#138)
```

## Registry (ghcr.io)

```bash
# From your laptop — log in to push by hand if CD is broken:
echo "$GHCR_PAT" | docker login ghcr.io -u <gh-user> --password-stdin

# List tags for our images:
curl -s -H "Authorization: Bearer $(echo $GHCR_PAT | base64)" \
  https://ghcr.io/v2/devtroopers-itu/minitwit/tags/list

# Pull and retag a specific digest locally for debugging:
docker pull ghcr.io/devtroopers-itu/minitwit:latest
```

The PAT ("classic" token with `read:packages` + `write:packages`) is stored in GitHub Actions as the default `GITHUB_TOKEN` for the org's builds. Node-local logins use the same PAT — it's in `~/.docker/config.json` on each node.

## CI/CD (GitHub Actions)

```bash
gh run list --workflow=ci.yml --limit 5         # recent CI runs
gh run list --workflow=cd.yml --limit 5
gh run view <run-id> --log                      # full log of a run
gh run watch                                    # tail the in-progress run
gh pr checks <pr-number>                        # check gate status on a PR
```

## Firewall / UFW

```bash
ssh root@<node-ip> 'ufw status verbose'         # current rules
# Rule changes: edit manually, never via the DO firewall panel alone
# (see ufw-setup-recap.md in this folder for the baseline).
```

## Common gotchas

- **Never edit files on the manager** (`/root/itu-minitwit/`). Anything you change there blocks the next CD run. All edits go through git.
- **`docker-compose.yml` is for local dev only**. Prod uses `docker-stack.yml`. They look similar and drift matters.
- **Webserver replicas share Postgres but nothing else**. No local file state in prod. If a "fix" involves writing to disk on the webserver, it'll disappear on the next rolling update.
- **Traefik runs in host mode**, not Swarm ingress. All public traffic enters on the manager only. Workers have 80/443/8080 closed to the public (see `firewall-changes.md`).

## See also

- [`docker-swarm.md`](docker-swarm.md) — why Swarm, initial setup, service distribution
- [`firewall-changes.md`](firewall-changes.md) — manager vs worker port exposure
- [`ufw-setup-recap.md`](ufw-setup-recap.md) — UFW baseline on all nodes
- [`../incidents/session11-ops-debug.md`](../incidents/session11-ops-debug.md) — Traefik v3.6 debug diary (host mode fix)
- [`../architecture/latest-counter-in-db.md`](../architecture/latest-counter-in-db.md) — why `/latest` is backed by Postgres
- [`../monitoring/alerting.md`](../monitoring/alerting.md) — Grafana + Discord alert rules
