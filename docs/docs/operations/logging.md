---
sidebar_position: 3
---

# Logging

Logs are text by default. Use `-json` for JSON logs:

```bash
ical-filter-proxy -json
```

Use `-debug` to log filter-level details:

```bash
ical-filter-proxy -debug
```

:::warning
Debug logging can include calendar names, request paths, event summaries, and
filter descriptions. Avoid enabling it where logs are broadly accessible or
retained longer than necessary.
:::

Successful health and readiness probes are not logged. Failed health/readiness
requests are logged.
