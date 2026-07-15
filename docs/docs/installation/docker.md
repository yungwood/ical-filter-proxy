---
sidebar_position: 1
---

# Docker

Docker images are published to Docker Hub:

```text
yungwood/ical-filter-proxy
```

View the image repository on
[Docker Hub](https://hub.docker.com/r/yungwood/ical-filter-proxy).

## Docker Run

Run the container with a config file mounted at `/app/config.yaml`:

```bash
docker run -d \
  --name ical-filter-proxy \
  -v ./config.yaml:/app/config.yaml:ro \
  -p 8080:8080 \
  --restart unless-stopped \
  yungwood/ical-filter-proxy:latest
```

The public listener defaults to `:8080`. Set `-address` to listen on a
different address.

iCal Filter Proxy serves plain HTTP. Use a reverse proxy or load balancer for
HTTPS when needed. See [HTTP Behavior](../operations/http-behavior.md).

## Docker Compose

Create a config file next to your Compose file:

```yaml title="config.yaml"
calendars:
  - name: example
    publish_name: "Example Calendar"
    public: true
    feed_url: "https://example.com/calendar.ics"
    filters:
      - description: "Remove cancelled events"
        remove: true
        match:
          summary:
            prefix: "Canceled: "
```

Create the Compose file:

```yaml title="docker-compose.yaml"
services:
  ical-filter-proxy:
    image: yungwood/ical-filter-proxy:latest
    container_name: ical-filter-proxy
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/app/config.yaml:ro
    restart: unless-stopped
```

Start the service:

```bash
docker compose up -d
```

The filtered feed is available at:

```text
http://localhost:8080/calendars/example/feed
```

Pass runtime flags with `command`:

```yaml title="docker-compose.yaml"
services:
  ical-filter-proxy:
    image: yungwood/ical-filter-proxy:latest
    command:
      - -metrics
      - -management-address
      - :9090
    ports:
      - "8080:8080"
      - "9090:9090"
    volumes:
      - ./config.yaml:/app/config.yaml:ro
```

## Docker Compose With Caddy HTTPS

Use Caddy when you want automatic Let's Encrypt certificates for a public FQDN
such as `ical.example.com`.

Before starting, make sure:

- `ical.example.com` has DNS `A` or `AAAA` records pointing to the host.
- Ports `80` and `443` are reachable from the internet.
- No other service is already bound to ports `80` or `443` on the host.

Create the application config:

```yaml title="config.yaml"
calendars:
  - name: work
    publish_name: "Work Calendar"
    token: "work-feed-token"
    feed_url: "https://outlook.office365.com/owa/calendar/example/calendar.ics"
    filters:
      - description: "Remove cancelled events"
        remove: true
        match:
          summary:
            prefix: "Canceled: "
```

Create the Caddy config:

```caddyfile title="Caddyfile"
{
	email admin@example.com
}

ical.example.com {
	reverse_proxy ical-filter-proxy:8080
}
```

Create the Compose file:

```yaml title="docker-compose.yaml"
services:
  ical-filter-proxy:
    image: yungwood/ical-filter-proxy:latest
    container_name: ical-filter-proxy
    environment:
      ICAL_FILTER_PROXY_TRUSTED_PROXY_CIDRS: 172.30.0.0/24
    volumes:
      - ./config.yaml:/app/config.yaml:ro
    networks:
      ical-proxy:
    restart: unless-stopped

  caddy:
    image: caddy:2-alpine
    container_name: ical-filter-proxy-caddy
    depends_on:
      - ical-filter-proxy
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
    networks:
      ical-proxy:
    restart: unless-stopped

networks:
  ical-proxy:
    ipam:
      config:
        - subnet: 172.30.0.0/24

volumes:
  caddy_data:
  caddy_config:
```

Start both containers:

```bash
docker compose up -d
```

The filtered feed is available at:

```text
https://ical.example.com/calendars/work/feed?token=work-feed-token
```

Caddy terminates HTTPS and forwards requests to iCal Filter Proxy over the
private Compose network. The app container does not publish its own port to the
host. `ICAL_FILTER_PROXY_TRUSTED_PROXY_CIDRS` allows logs to use the client
address from Caddy's `X-Forwarded-For` header for requests arriving from that
Compose network.

## Docker Compose Secrets

Compose secrets can be used with `token_file` and `feed_url_file`.

```yaml title="config.yaml"
calendars:
  - name: private
    token_file: /run/secrets/calendar_token
    feed_url_file: /run/secrets/calendar_feed_url
```

```yaml title="docker-compose.yaml"
services:
  ical-filter-proxy:
    image: yungwood/ical-filter-proxy:latest
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/app/config.yaml:ro
    secrets:
      - calendar_token
      - calendar_feed_url

secrets:
  calendar_token:
    file: ./secrets/calendar-token
  calendar_feed_url:
    file: ./secrets/calendar-feed-url
```

Compose mounts each secret at `/run/secrets/<secret_name>`. The files should
contain only the raw token or upstream feed URL.

:::info
Secret-file config takes precedence over inline `token` and `feed_url` values.
:::

If you prefer environment variables, see
[Environment Substitution](../configuration/environment-substitution.md).
