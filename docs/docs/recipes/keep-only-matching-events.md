---
sidebar_position: 7
---

# Keep Only Matching Events

## Use Case

Use a keep-list pattern when a feed contains many events but you only want to
publish events that match specific rules.

## Config

```yaml title="config.yaml excerpt"
filters:
  - description: "Keep engineering on-call events"
    match:
      summary:
        contains: "Roster"
        contains_any:
          - "Primary On Call"
          - "Secondary On Call"
    stop: true

  - description: "Remove everything else"
    remove: true
```

## How It Works

String match conditions are combined with AND logic. In this example, the event
summary must contain `Roster` and must also contain either `Primary On Call` or
`Secondary On Call`.

The first filter uses `stop: true` so matching events are kept and later filters
do not run for them. The final filter has no match rules, so it matches any
event that reaches it and removes it.
