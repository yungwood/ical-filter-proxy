---
sidebar_position: 2
---

# Remove Optional Events

## Use Case

Use `contains_any` when a provider marks optional events with more than one
phrase.

## Config

```yaml title="config.yaml excerpt"
filters:
  - description: "Remove optional events"
    remove: true
    match:
      summary:
        contains_any:
          - "[Optional]"
          - "Optional:"
          - "FYI"
```

## How It Works

The filter matches when the summary contains at least one listed value.
