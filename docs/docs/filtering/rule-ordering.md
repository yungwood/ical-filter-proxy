---
sidebar_position: 4
---

# Rule Ordering

Filters run in the order they appear in the calendar config.

For each event:

1. The first filter is evaluated.
2. If the filter matches, its remove/transform behavior is applied.
3. If the filter has `stop: true`, no later filters run for that event.
4. The next filter is evaluated unless processing stopped or the event was removed.

A filter with no `match` block always matches.

```yaml title="config.yaml excerpt"
filters:
  - description: "Keep on-call events and rename them"
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

String transforms are deterministic:

1. `remove`
2. `replace`
3. `trim_prefix`
4. `trim_suffix`
5. `replace_text`
6. `prefix`
7. `suffix`

`remove` and `replace` are terminal. If either is set, later string transform
steps do not run for that property.
