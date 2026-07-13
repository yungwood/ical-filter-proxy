---
sidebar_position: 2
---

# Match Rules

Match rules decide whether a filter applies to an event.

The supported event properties are:

| Property | Description |
| --- | --- |
| `summary` | Event title. |
| `description` | Event description/body. |
| `location` | Event location. |
| `url` | Event URL. |

String match conditions:

| Condition | Description |
| --- | --- |
| `empty` | Matches when the property is absent or empty. |
| `contains` | Matches when the property contains this literal text. |
| `contains_any` | Matches when the property contains at least one listed value. |
| `contains_all` | Matches when the property contains every listed value. |
| `prefix` | Matches when the property starts with this literal text. |
| `suffix` | Matches when the property ends with this literal text. |
| `regex` | Matches using a Go regular expression. |

All rules in a filter must match for the filter to apply:

```yaml title="config.yaml excerpt"
filters:
  - description: "Remove cancelled on-call events"
    remove: true
    match:
      summary:
        prefix: "Canceled: "
      description:
        contains: "on-call"
```

:::warning
Invalid regular expressions are rejected during config compilation before the
service starts.
:::
