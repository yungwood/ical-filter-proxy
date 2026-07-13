---
sidebar_position: 1
---

# Filtering

Filters are evaluated for each event in a calendar feed.

Each calendar can define an ordered list of filters:

```yaml title="config.yaml"
calendars:
  - name: example
    public: true
    feed_url: "https://example.com/calendar.ics"
    filters:
      - description: "Remove cancelled events"
        remove: true
        match:
          summary:
            prefix: "Canceled: "
```

A filter can remove an event, transform event properties, or stop later filters
from running.

If no filter removes an event, the event is kept.

## Next Steps

- [Define match rules](./match-rules.md)
- [Apply transforms](./transforms.md)
- [Browse recipes](../recipes/remove-cancelled-events.md)
