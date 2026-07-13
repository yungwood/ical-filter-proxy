---
sidebar_position: 6
---

# Environment Substitution

`token` and `feed_url` can reference environment variables by using an exact
`${ENV_NAME}` value.

```yaml title="config.yaml"
calendars:
  - name: private
    token: "${CALENDAR_TOKEN}"
    feed_url: "${CALENDAR_FEED_URL}"
```

The environment variable must exist and must not be empty. Missing or empty
variables fail startup.

:::note
Environment substitution is only supported for `token` and `feed_url`.
:::

Only exact references are expanded. Partial string interpolation is not
supported:

```yaml title="config.yaml"
calendars:
  - name: example
    token: "${CALENDAR_TOKEN}"
    feed_url: "https://example.com/${TOKEN}/calendar.ics"
```

In this example, `token` is expanded but `feed_url` remains literal.

:::info
Secret files take precedence over environment substitution. If `token_file` or
`feed_url_file` is set, the file value is used.
:::
