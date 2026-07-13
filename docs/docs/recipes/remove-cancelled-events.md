---
sidebar_position: 1
---

# Remove Cancelled Events

## Use Case

Some providers keep cancelled events in the feed and mark them with a prefix.

## Config

```yaml title="config.yaml excerpt"
filters:
  - description: "Remove cancelled events"
    remove: true
    match:
      summary:
            prefix: "Canceled: "
```

## How It Works

This removes any event whose summary starts with `Canceled: `.
