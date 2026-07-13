---
sidebar_position: 5
---

# Secret Files

Use secret files when tokens or upstream feed URLs should not be stored directly
in the YAML config.

```yaml title="config.yaml"
calendars:
  - name: outlook
    token_file: "/run/secrets/outlook-token"
    feed_url_file: "/run/secrets/outlook-feed-url"
```

:::info
`token_file` takes precedence over `token`.

`feed_url_file` takes precedence over `feed_url`.
:::

File contents are trimmed of surrounding whitespace. Missing or unreadable files
fail startup.

Secret files are useful with Docker secrets, Kubernetes projected secrets, and
other runtime secret managers that expose values as files.
