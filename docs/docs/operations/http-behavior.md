---
sidebar_position: 6
---

# HTTP Behavior

:::note[TLS termination]
iCal Filter Proxy serves plain HTTP only. Run it behind an ingress controller,
reverse proxy, or load balancer to terminate TLS when HTTPS is required.
See [Reverse Proxies](./reverse-proxies.md) for forwarded client address
configuration.
:::

:::warning
If private calendar feeds use token authentication, expose them over HTTPS at
the proxy or load balancer layer. Tokens are passed as query parameters.
:::

Calendar feed endpoints allow only `GET` and `HEAD`. `HEAD` validates access
and fetches the upstream calendar like `GET`, but returns headers without a
response body.

Unsupported methods return `405 Method Not Allowed` with:

```text
Allow: GET, HEAD
```

Private calendars return `401 Unauthorized` when the token is missing or wrong.

Upstream fetch or parse failures return `502 Bad Gateway`.

Calendar responses include:

```text
Cache-Control: no-store
X-Content-Type-Options: nosniff
Content-Type: text/calendar; charset=utf-8
```

The HTTP server uses timeouts for request headers, reads, writes, and idle
connections. It also performs graceful shutdown on `SIGINT` and `SIGTERM`.
