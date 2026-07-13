---
sidebar_position: 5
---

# Clean OpsGenie Feed

## Use Case

This pattern keeps on-call schedule events, renames them, and removes everything
else.

## Config

```yaml title="config.yaml excerpt"
filters:
  - description: "Keep on-call schedule events and fix names"
    match:
      summary:
        contains: "schedule: oncall"
    stop: true
    transform:
      summary:
        replace: "On-Call"

  - description: "Remove all other events"
    remove: true
```

## How It Works

The first filter uses `stop: true` so kept events are not removed by the catch-all
filter that follows.
