---
sidebar_position: 2
---

# Metrics

Enable Prometheus metrics with `-metrics`.

```bash
ical-filter-proxy -metrics
```

Metrics are exposed at `/metrics`.

Use a management listener to keep metrics separate from public calendar feeds:

```bash
ical-filter-proxy -metrics -management-address 127.0.0.1:9090
```

Per-calendar metric labels are disabled by default. Enable them with:

```bash
ical-filter-proxy -metrics -metrics-calendar-labels
```

Per-calendar labels are useful for debugging a small known set of calendars, but
they increase metric cardinality.

:::note
Use `-management-address` with metrics when you want `/metrics` separated from
the public calendar listener.
:::
