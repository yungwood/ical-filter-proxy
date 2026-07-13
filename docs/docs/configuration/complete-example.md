---
sidebar_position: 7
---

# Complete Config Example

This example combines public feeds, private token feeds, secret-backed values,
match rules, and transforms.

```yaml title="config.yaml"
calendars:
  - name: holidays
    publish_name: "Public Holidays"
    public: true
    feed_url: "https://example.com/holidays.ics"

  - name: work
    publish_name: "Work Calendar"
    token: "changeme"
    feed_url: "https://example.com/work.ics"
    user_agent: "ical-filter-proxy"
    filters:
      - description: "Remove cancelled events"
        remove: true
        match:
          summary:
            prefix: "Canceled: "

      - description: "Remove optional events"
        remove: true
        match:
          summary:
            contains_any:
              - "[Optional]"
              - "Optional:"
              - "FYI"

      - description: "Clean summary text"
        transform:
          summary:
            trim_prefix: "Roster - "
            replace_text:
              old: "Primary On Call"
              new: "On-Call"
              all: true

  - name: private
    publish_name: "Private Calendar"
    token_file: "/run/secrets/private-token"
    feed_url_file: "/run/secrets/private-feed-url"
```

In this example:

- `holidays` is public and does not require a token.
- `work` is private and requires `?token=changeme`.
- `private` reads the access token and upstream feed URL from files.
- filters run in order for each event.
