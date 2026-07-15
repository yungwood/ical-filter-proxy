---
sidebar_position: 7
---

# Complete Config Example

This example combines public feeds, private token feeds, secret-backed values,
match rules, event removal, and transforms.

```yaml title="config.yaml"
calendars:
  - name: holidays
    publish_name: "Public Holidays"
    public: true
    feed_url: "https://example.com/holidays.ics"

  - name: work
    publish_name: "Work Calendar"
    token: "work-feed-token"
    feed_url: "https://outlook.office365.com/owa/calendar/example/calendar.ics"
    user_agent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36"
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

      - description: "Remove followed events"
        remove: true
        match:
          summary:
            prefix: "Following: "

      - description: "Remove private appointments"
        remove: true
        match:
          summary:
            contains: "Private Appointment"

  - name: roster
    publish_name: "On-Call Roster"
    public: true
    feed_url: "https://example.com/team-roster.ics"
    filters:
      - description: "Keep and clean on-call roster events"
        match:
          summary:
            contains_all:
              - "Roster"
            contains_any:
              - "Primary On Call"
              - "Secondary On Call"
        stop: true
        transform:
          summary:
            trim_prefix: "Roster - "
            replace_text:
              old: "Primary On Call"
              new: "On-Call"
              all: true

      - description: "Remove everything else"
        remove: true

  - name: private
    publish_name: "Private Calendar"
    token_file: "/run/secrets/private-token"
    feed_url_file: "/run/secrets/private-feed-url"
```

In this example:

- `holidays` is public and does not require a token.
- `work` is private and requires `?token=work-feed-token`.
- `roster` keeps only selected roster events and cleans their summaries.
- `private` reads the access token and upstream feed URL from files.
- filters run in order for each event.
