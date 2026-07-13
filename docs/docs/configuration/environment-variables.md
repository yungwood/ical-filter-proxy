---
sidebar_position: 4
---

# Environment Variables

Most runtime flags can be set with environment variables. CLI flags take
precedence over environment variables.

:::info
CLI flags always override environment variables.
:::

Use environment variables when runtime settings are easier to provide through a
container, systemd unit, Kubernetes manifest, or secret manager.

```bash
ICAL_FILTER_PROXY_CONFIG=/etc/ical-filter-proxy/config.yaml \
ICAL_FILTER_PROXY_ADDRESS=:8080 \
ICAL_FILTER_PROXY_METRICS=true \
ICAL_FILTER_PROXY_MANAGEMENT_ADDRESS=:9090 \
ical-filter-proxy
```

In Docker Compose:

```yaml title="docker-compose.yaml"
services:
  ical-filter-proxy:
    image: yungwood/ical-filter-proxy:latest
    environment:
      ICAL_FILTER_PROXY_METRICS: "true"
      ICAL_FILTER_PROXY_MANAGEMENT_ADDRESS: ":9090"
```

Boolean environment variables use Go boolean parsing, so values such as `true`,
`false`, `1`, and `0` are accepted.

`ICAL_FILTER_PROXY_TRUSTED_PROXY_CIDRS` accepts a comma-separated list:

```bash
ICAL_FILTER_PROXY_TRUSTED_PROXY_CIDRS=10.0.0.0/8,192.168.0.0/16
```

:::warning
Debug logging includes calendar names, request paths, event summaries, and
filter descriptions. Avoid enabling it in environments where logs are broadly
accessible or retained longer than necessary.
:::

See [CLI Flags](../reference/cli-flags.md) for the full flag and environment
variable reference.
