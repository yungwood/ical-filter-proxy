---
sidebar_position: 4
---

# Health and Readiness

The service exposes:

```text
/liveness
/readiness
```

Both endpoints return success when the process is running and the HTTP handler
is available.

Set `-management-address` to move these endpoints to a separate listener:

```bash
ical-filter-proxy -management-address 127.0.0.1:9090
```

This is recommended for Kubernetes and other environments where probes should
not share the public calendar listener.
