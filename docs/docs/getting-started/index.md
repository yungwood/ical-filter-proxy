---
sidebar_position: 1
sidebar_label: Quick Start
---

# Quick Start

This starts a public example feed with Docker and a local `config.yaml`.

Create a config file:

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

Run the service:

```bash
docker run --rm \
  -v ./config.yaml:/app/config.yaml:ro \
  -p 8080:8080 \
  yungwood/ical-filter-proxy:latest
```

Subscribe to the filtered feed:

```text
http://localhost:8080/calendars/example/feed
```

## Next Steps

- [Install with Docker](../installation/docker.md)
- [Configure calendars](../configuration/calendars.md)
- [Learn filtering](../filtering/index.md)
