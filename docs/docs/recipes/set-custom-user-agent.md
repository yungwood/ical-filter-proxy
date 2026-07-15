---
sidebar_position: 10
---

# Set A Custom User-Agent

## Use Case

Some upstream calendar providers reject generic HTTP clients. Set `user_agent`
when a provider expects a browser-like client string.

## Config

```yaml title="config.yaml excerpt"
calendars:
  - name: work
    publish_name: "Work Calendar"
    token: "work-feed-token"
    feed_url: "https://outlook.office365.com/owa/calendar/example/calendar.ics"
    user_agent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36"
```

## How It Works

`user_agent` overrides the `User-Agent` header sent only to that calendar's
upstream feed. When it is not set, iCal Filter Proxy sends
`ical-filter-proxy/<version>`.
