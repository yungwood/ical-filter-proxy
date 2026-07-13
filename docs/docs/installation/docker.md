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
