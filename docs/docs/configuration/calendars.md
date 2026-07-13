---
sidebar_position: 2
---

# Calendars

Each item under `calendars` defines one published feed.

```yaml title="config.yaml"
calendars:
  - name: outlook
    publish_name: "Work Calendar"
    token: "changeme"
    feed_url: "https://outlook.office365.com/owa/calendar/example.ics"
    user_agent: "ical-filter-proxy"
```

Common fields:

| Field | Required | Description |
| --- | --- | --- |
| `name` | Yes | Calendar slug used in the feed URL. |
| `feed_url` | Yes, unless `feed_url_file` is set | Upstream iCalendar URL. Must start with `http://` or `https://`. |
| `publish_name` | No | Calendar name written to the published feed. Uses the upstream value when omitted. |
| `public` | No | Allows access without a token when `true`. |
| `token` | Required unless `public: true` or `token_file` is set | Token required to access the published feed. |
| `token_file` | No | File containing the access token. Takes precedence over `token`. |
| `feed_url_file` | No | File containing the upstream feed URL. Takes precedence over `feed_url`. |
| `user_agent` | No | Overrides the User-Agent header sent to this upstream feed. |
| `filters` | No | Ordered filter and transform rules applied to events. |

Published feeds are exposed at:

```text
/calendars/<name>/feed
/calendars/<name>/feed?token=<token>
```
