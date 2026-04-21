# Firewall Hardening — Worker Node Port Exposure

## Background

We use **host mode** for Traefik, pinned to the manager node. This means all public traffic (HTTP/HTTPS/simulator) enters through the manager only. Worker nodes never receive external traffic directly — they only process requests forwarded internally via the Docker Swarm overlay network.

Because of this, having ports 80, 443, and 8080 open to all IPv4/IPv6 on the worker nodes is unnecessary and increases our attack surface for no benefit.

## What Needs to Change

**Manager node** — no change, keep as is:
| Port | Protocol | Source |
|------|----------|--------|
| 22 | TCP | All |
| 80 | TCP | All |
| 443 | TCP | All |
| 8080 | TCP | All (consider restricting to simulator IP) |
| 2377, 7946, 4789 | TCP/UDP | minitwit-swarm tag only |

**Worker nodes** — remove 80, 443, 8080 from public exposure:
| Port | Protocol | Source |
|------|----------|--------|
| 22 | TCP | All |
| 2377, 7946, 4789 | TCP/UDP | minitwit-swarm tag only |

## How to Apply on DigitalOcean

1. Go to **Networking → Firewalls** in the DigitalOcean dashboard
2. Click your existing firewall → **"..." → Duplicate**
3. Delete the 80, 443, and 8080 inbound rules from the duplicate
4. Rename it `minitwit-workers` and assign it to the two worker droplets

## Why This Matters

- Workers currently accept connections on port 80, 443, and 8080 from the entire internet, even though nothing is listening there for public traffic
- Reducing open ports limits the blast radius if a worker node is ever compromised
- This is standard practice when using a dedicated ingress node in a Swarm setup

---

## Worker → Database Communication

You might wonder: if we restrict inbound traffic on the workers, how do they still talk to the managed database?

### Outbound is unrestricted

DigitalOcean firewalls allow all outbound traffic by default. Workers connect **to** the database (outbound), so no inbound rule is needed for this. The connection flow is:

```
worker:RANDOM_PORT → database:5432   (outbound, allowed)
database:5432 → worker:RANDOM_PORT   (response, automatically allowed)
```

### Stateful firewalls — how responses get back in

DigitalOcean uses a **stateful firewall**, meaning it tracks active connections. When a worker initiates an outbound connection, the firewall remembers it. When the database sends a response back, the firewall recognises it as a reply to an existing connection and allows it in automatically — no explicit inbound rule needed.

This is called **connection tracking**. The rule of thumb is:
- **Inbound rules** control who can *initiate* new connections to your node
- **Responses** to outbound connections are always allowed back in automatically

So restricting inbound on the workers has zero impact on database communication.

### Don't forget — add all three droplets to the DB trusted sources

Since all 3 webserver replicas can run on any node, make sure all three droplet IPs are listed under **Databases → your cluster → Trusted Sources** in DigitalOcean, not just the manager.
