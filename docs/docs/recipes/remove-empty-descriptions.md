---
sidebar_position: 4
---

# Remove Empty Descriptions

## Use Case

Use `empty: true` to remove events where a field is absent or blank.

## Config

```yaml title="config.yaml excerpt"
filters:
  - description: "Remove events without descriptions"
    remove: true
    match:
      description:
        empty: true
```

## How It Works

The filter matches when the event description is missing or empty, then removes
the event.
