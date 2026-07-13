---
sidebar_position: 1
---

# Endpoints

Calendar feeds are exposed on the public listener:

```text
GET  /calendars/<calendar>/feed
HEAD /calendars/<calendar>/feed
```

Private calendars require a token query parameter:

```text
/calendars/work/feed?token=changeme
```

Management endpoints:

```text
GET /liveness
GET /readiness
GET /metrics
```

When `-management-address` is set, management endpoints are served only on the
management listener. Without `-management-address`, they are served on the
public listener for single-port operation.

:::note
The Helm chart enables a separate management listener by default.
:::

## Next Steps

- [Review HTTP behavior](./http-behavior.md)
- [Enable metrics](./metrics.md)
- [Configure logging](./logging.md)
