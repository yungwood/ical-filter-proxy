---
sidebar_position: 1
---

# CLI Flags

CLI flags take precedence over environment variables.

| Flag | Environment variable | Default | Description |
| --- | --- | --- | --- |
| `-config` | `ICAL_FILTER_PROXY_CONFIG` | `config.yaml` | Config file path. |
| `-address` | `ICAL_FILTER_PROXY_ADDRESS` | `:8080` | Public calendar listener address. |
| `-management-address` | `ICAL_FILTER_PROXY_MANAGEMENT_ADDRESS` | empty | Optional management listener address for liveness, readiness, and metrics. |
| `-debug` | `ICAL_FILTER_PROXY_DEBUG` | `false` | Enable debug logging. |
| `-json` | `ICAL_FILTER_PROXY_JSON` | `false` | Emit logs as JSON. |
| `-validate` | `ICAL_FILTER_PROXY_VALIDATE` | `false` | Validate config and exit. |
| `-metrics` | `ICAL_FILTER_PROXY_METRICS` | `false` | Enable Prometheus metrics. |
| `-metrics-calendar-labels` | `ICAL_FILTER_PROXY_METRICS_CALENDAR_LABELS` | `false` | Enable per-calendar metric labels. |
| `-trusted-proxy-cidr` | `ICAL_FILTER_PROXY_TRUSTED_PROXY_CIDRS` | empty | Trusted reverse proxy CIDR for forwarded client addresses. The flag may be repeated; the environment variable is comma-separated. |
| `-version` | none | `false` | Print version and exit. |

Boolean environment variables use Go boolean parsing.
