---
sidebar_position: 3
---

# Remove Public Holidays

## Use Case

Use a regex matcher when provider wording varies.

## Config

```yaml title="config.yaml excerpt"
filters:
  - description: "Remove public holidays"
    remove: true
    match:
      summary:
        regex: ".*[Pp]ublic [Hh]oliday.*"
```

## How It Works

Invalid regexes fail config compilation before the service starts.
