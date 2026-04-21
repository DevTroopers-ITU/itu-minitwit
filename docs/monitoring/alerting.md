# Alerting

## Overview

Alerts are evaluated by Prometheus and routed through Grafana to a Discord channel. This keeps the team informed about production issues without requiring anyone to watch dashboards.

## How it works

```
Prometheus (evaluates rules every 15s)
    │
    │  alert fires
    ▼
Grafana (receives alert, applies notification policy)
    │
    │  groups alerts, respects repeat_interval
    ▼
Discord (#alerts channel via webhook)
```

Prometheus continuously evaluates the alert rules defined in `monitoring/prometheus/prometheus.rules.yml`. When a condition is met for the specified duration (`for`), it fires an alert. Grafana picks this up and routes it to the Discord webhook based on the notification policy.

## Alert rules

All rules are defined in `monitoring/prometheus/prometheus.rules.yml`.

| Alert | Condition | Duration | Severity |
|-------|-----------|----------|----------|
| **WebserverDown** | `up{job="itu-minitwit-app"} == 0` | 1 minute | critical |
| **HighErrorRate** | >10% of responses are 5xx over 5 min | 2 minutes | warning |
| **SlowResponses** | P95 response time > 1 second | 5 minutes | warning |

### Why these thresholds?

- **WebserverDown** fires after 1 minute to avoid false alarms from brief restarts (e.g., during deploys).
- **HighErrorRate** uses a 5-minute rate window and waits 2 minutes, so a single failed request won't trigger it.
- **SlowResponses** requires 5 minutes of sustained slow performance. Brief spikes are normal and ignored.

## Notification policy

Defined in `monitoring/grafana/alerting/policies.yml`.

- **group_by: alertname** — alerts of the same type are grouped into one notification.
- **group_wait: 30s** — waits 30 seconds before sending, so alerts that fire together arrive as one message.
- **group_interval: 5m** — if new alerts join an existing group, wait 5 minutes before updating.
- **repeat_interval: 4h** — a still-firing alert only re-notifies every 4 hours. This prevents Discord spam.

## Discord setup

The webhook URL is stored as a GitHub secret (`DISCORD_WEBHOOK_URL`) and passed to the Grafana container via the `DISCORD_WEBHOOK_URL` environment variable. The contact point config in `monitoring/grafana/alerting/contactpoints.yml` references it as `${DISCORD_WEBHOOK_URL}`.

## Files involved

| File | Role |
|------|------|
| `monitoring/prometheus/prometheus.rules.yml` | Alert rule definitions (PromQL expressions) |
| `monitoring/grafana/alerting/contactpoints.yml` | Discord webhook contact point |
| `monitoring/grafana/alerting/policies.yml` | Notification routing and rate limiting |
| `monitoring/grafana/Dockerfile` | Copies alerting config into the Grafana image |
| `docker-compose.yml` | Passes `DISCORD_WEBHOOK_URL` env to Grafana |
| `.github/workflows/cd.yml` | Writes `DISCORD_WEBHOOK_URL` to `.env` on deploy |

## Testing alerts

To verify alerts work after deployment:

1. Check Prometheus: `http://<server>:9090/alerts` — all rules should show as "inactive" (green).
2. Check Grafana: Alerting > Alert rules — same rules should appear.
3. Check Grafana: Alerting > Contact points — "discord" should be listed.
4. To test the Discord webhook manually, use "Test" button in Grafana's contact point UI.
