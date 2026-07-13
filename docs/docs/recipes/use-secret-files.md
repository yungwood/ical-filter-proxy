---
sidebar_position: 7
---

# Use Secret Files

## Use Case

Use `token_file` and `feed_url_file` when secrets are mounted as files.

## Config

```yaml title="config.yaml"
calendars:
  - name: private
    token_file: "/run/secrets/calendar-token"
    feed_url_file: "/run/secrets/calendar-feed-url"
```

## How It Works

File-backed values take precedence over inline `token` and `feed_url` values.
