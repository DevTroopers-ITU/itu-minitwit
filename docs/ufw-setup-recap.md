# UFW Setup — Recap

## What is UFW?

UFW (Uncomplicated Firewall) is a host-level firewall that runs directly on each Linux server. It is a user-friendly interface for **iptables**, which is the actual firewall engine built into the Linux kernel.

```
Internet
   ↓
DO Firewall        ← perimeter, stops traffic before it hits the droplet
   ↓
Your Droplet
   ↓
UFW/iptables       ← host-level, stops traffic if DO firewall is misconfigured
   ↓
Your App
```

Having both the DigitalOcean perimeter firewall and UFW is **defense in depth** — if one layer is misconfigured, the other catches it.

---

## How UFW Works

UFW works like a shopping list. You add rules one by one and nothing is active yet:

```bash
ufw allow 22/tcp      # add to list
ufw allow 80/tcp      # add to list
ufw allow 443/tcp     # add to list
```

When you're ready, you flip the switch and all rules go live at once:

```bash
ufw enable
```

Once enabled, UFW also starts automatically on system reboot — so rules persist across restarts without any manual intervention.

---

## The Docker/UFW Bypass Problem

When Docker maps ports using `-p`, it writes rules directly to **iptables** and bypasses UFW entirely. This means a port blocked in UFW can still be open if Docker published it.

Since we use **Traefik in host mode** rather than Docker's `-p` port mapping, we are less exposed to this than a typical Docker setup. Traefik handles all public traffic on the manager node directly.

### The Fix — Docker USER Chain

To make Docker respect UFW rules, we add a `DOCKER-USER` iptables chain on every node. Docker checks this chain before processing traffic, giving UFW a chance to intercept it.

On **each node**, add the following to the bottom of `/etc/ufw/after.rules`:

```
# BEGIN UFW AND DOCKER
*filter
:ufw-user-forward - [0:0]
:DOCKER-USER - [0:0]
-A DOCKER-USER -j ufw-user-forward
-A DOCKER-USER -j RETURN
COMMIT
# END UFW AND DOCKER
```

Then reload UFW:
```bash
ufw reload
```

**Why we added this on every node:** Even though only the manager runs Traefik, Docker is running on all three nodes and could in theory open ports via iptables on any of them — for example if a service is accidentally deployed without a placement constraint. Adding the `DOCKER-USER` chain on all nodes ensures Docker traffic is routed through UFW everywhere, not just on the manager.

---

## Why Worker Nodes Have Different Rules

We use host mode with Traefik pinned to the manager node. This means:
- All public traffic (HTTP/HTTPS/simulator) enters through the **manager only**
- Worker nodes never receive external traffic directly
- Workers only need SSH and Docker Swarm internal ports

Opening 80, 443, and 8080 on workers would be unnecessary attack surface.

---

## Rules Applied

### Manager Node
```bash
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp                                        # SSH
ufw allow 80/tcp                                        # HTTP (redirects to HTTPS)
ufw allow 443/tcp                                       # HTTPS
ufw allow 8080/tcp                                      # Simulator endpoint
ufw allow from <worker1-ip> to any port 2377 proto tcp  # Swarm management
ufw allow from <worker1-ip> to any port 7946 proto tcp  # Node communication
ufw allow from <worker1-ip> to any port 7946 proto udp  # Node communication
ufw allow from <worker1-ip> to any port 4789 proto udp  # Overlay network
# repeat for worker2
ufw enable
```

### Worker Nodes
```bash
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp                                        # SSH
ufw allow from <manager-ip> to any port 2377 proto tcp  # Swarm management
ufw allow from <manager-ip> to any port 7946 proto tcp  # Node communication
ufw allow from <manager-ip> to any port 7946 proto udp  # Node communication
ufw allow from <manager-ip> to any port 4789 proto udp  # Overlay network
# repeat for other worker
ufw enable
```

---

## ICMP (Ping)

UFW allows ICMP from everyone by default via `/etc/ufw/before.rules`. We left this as is because:
- The DO perimeter firewall already restricts ICMP to the `minitwit-swarm` tag
- ICMP is low risk — it only tells someone the server is alive
- UFW's ICMP syntax is limited and not worth fighting for minimal gain

---

## Does UFW Affect the Monitoring Stack?

No. Prometheus, Grafana, Loki, and Promtail all communicate over the Docker **backend overlay network**, which is internal to the swarm. UFW only sees traffic coming in from outside the overlay network — container-to-container traffic never touches the host's network interface that UFW guards.

---

## Does UFW Affect the Database Connection?

No. UFW only restricts **inbound** traffic — who can initiate new connections to your node. The database connection is **outbound** from the workers, which is allowed by default.

DigitalOcean uses a **stateful firewall**, meaning it tracks active connections. When a worker initiates a connection to the database, the firewall remembers it and automatically allows the response back in — no explicit inbound rule needed.

---

## Verification

After enabling, verify with:
```bash
ufw status verbose
```

Expected output on a worker node:
```
Status: active
Default: deny (incoming), allow (outgoing)

22/tcp                ALLOW IN    Anywhere
2377/tcp              ALLOW IN    <manager-ip>
7946/tcp              ALLOW IN    <manager-ip>
7946/udp              ALLOW IN    <manager-ip>
4789/udp              ALLOW IN    <manager-ip>
2377/tcp              ALLOW IN    <other-worker-ip>
7946/tcp              ALLOW IN    <other-worker-ip>
7946/udp              ALLOW IN    <other-worker-ip>
4789/udp              ALLOW IN    <other-worker-ip>
```
