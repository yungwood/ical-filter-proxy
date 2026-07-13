---
sidebar_position: 1
---

# Configuration

iCal Filter Proxy is configured with a YAML file. By default, the service reads
`config.yaml` from the working directory.

Use `-config` to choose a different file:

```bash
ical-filter-proxy -config /etc/ical-filter-proxy/config.yaml
```

The config must contain at least one calendar:

```yaml title="config.yaml"
calendars:
  - name: example
    public: true
    feed_url: "https://example.com/calendar.ics"
```

Configuration is validated before the service starts. Invalid calendar settings,
missing secret files, missing environment references, invalid feed URLs, and
invalid regex match rules fail startup.

## Next Steps

- [Configure calendars](./calendars.md)
- [Choose an authentication mode](./authentication.md)
- [Use secret files](./secret-files.md)
