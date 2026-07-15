---
sidebar_position: 6
---

# Remove Outlook Private Appointments

## Use Case

Outlook can publish private events with a generic `Private Appointment` summary.
Remove them when they are not useful in the shared calendar feed.

## Config

```yaml title="config.yaml excerpt"
filters:
  - description: "Remove private appointments"
    remove: true
    match:
      summary:
        contains: "Private Appointment"
```

## How It Works

This removes events whose summary contains `Private Appointment`.
