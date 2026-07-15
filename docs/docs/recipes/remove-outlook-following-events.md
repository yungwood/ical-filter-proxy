---
sidebar_position: 5
---

# Remove Outlook Following Events

## Use Case

Outlook can include followed events that are useful in Outlook but noisy in a
published calendar feed.

## Config

```yaml title="config.yaml excerpt"
filters:
  - description: "Remove followed events"
    remove: true
    match:
      summary:
        prefix: "Following: "
```

## How It Works

This removes any event whose summary starts with `Following: `.
