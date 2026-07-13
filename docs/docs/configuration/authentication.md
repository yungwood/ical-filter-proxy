---
sidebar_position: 3
---

# Authentication

Calendar feeds are private by default.

## Private Token Feeds

A private calendar must define `token` or `token_file`.

```yaml title="config.yaml"
calendars:
  - name: private
    token: "changeme"
    feed_url: "https://example.com/private.ics"
```

Private feeds require the token query parameter:

```text
/calendars/private/feed?token=changeme
```

:::note
Blank tokens are not allowed for private calendars.
:::

## Public Feeds

Set `public: true` to allow access without a token:

```yaml title="config.yaml"
calendars:
  - name: public
    public: true
    feed_url: "https://example.com/public.ics"
```

:::warning
Public calendars must not define `token` or `token_file`. This is rejected at
startup to avoid ambiguous auth behavior.
:::

## Secret-Backed Tokens

Use `token_file` when the token should be read from a mounted file:

```yaml title="config.yaml"
calendars:
  - name: private
    token_file: "/run/secrets/calendar-token"
    feed_url: "https://example.com/private.ics"
```

For file-backed tokens and feed URLs, see [Secret Files](./secret-files.md).
