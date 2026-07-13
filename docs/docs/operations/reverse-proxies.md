---
sidebar_position: 5
---

# Reverse Proxies

iCal Filter Proxy serves plain HTTP. Use an ingress controller, reverse proxy,
or load balancer to terminate TLS when HTTPS is required.

:::warning
Use HTTPS at the proxy or load balancer layer for private calendar feeds. Tokens
are passed as query parameters.
:::

## Forwarded Client Addresses

By default, logs use the direct remote address from the HTTP connection.

When iCal Filter Proxy runs behind a trusted reverse proxy, configure the proxy
CIDR so `X-Forwarded-For` can be used for client address logging:

```bash
ical-filter-proxy -trusted-proxy-cidr 10.0.0.0/8
```

Multiple trusted proxy CIDRs can be configured by repeating the flag:

```bash
ical-filter-proxy \
  -trusted-proxy-cidr 10.0.0.0/8 \
  -trusted-proxy-cidr 192.168.0.0/16
```

Or with an environment variable:

```bash
ICAL_FILTER_PROXY_TRUSTED_PROXY_CIDRS=10.0.0.0/8,192.168.0.0/16
```

## X-Forwarded-For Behavior

`X-Forwarded-For` is only used when the direct remote address belongs to a
trusted proxy CIDR.

When trusted, iCal Filter Proxy reads `X-Forwarded-For` from right to left and
uses the nearest untrusted address as the client address.

If the direct remote address is not trusted, or the `X-Forwarded-For` header is
invalid, the direct remote address is used instead.

:::note
Do not configure broad trusted proxy CIDRs unless all clients in that range are
actually trusted proxies. Trusting the wrong source allows clients to control
the logged client address.
:::
