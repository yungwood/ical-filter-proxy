---
sidebar_position: 9
---

# Hide Event Descriptions

## Use Case

Use a transform to keep events visible while removing description text from the
published feed.

## Config

```yaml title="config.yaml excerpt"
filters:
  - description: "Remove descriptions from shared events"
    transform:
      description:
        remove: true
```

## How It Works

The filter has no match rules, so it applies to every event. The description
transform sets the description property to a blank string without removing the
event.
