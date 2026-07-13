---
sidebar_position: 3
---

# Transforms

Transforms rewrite event properties when a filter matches.

The supported transform properties are:

| Property | Description |
| --- | --- |
| `summary` | Event title. |
| `description` | Event description/body. |
| `location` | Event location. |
| `url` | Event URL. |

String transform options:

| Transform | Description |
| --- | --- |
| `remove` | Sets the property to a blank string. |
| `replace` | Replaces the entire property value. |
| `trim_prefix` | Removes this text when the value starts with it. |
| `trim_suffix` | Removes this text when the value ends with it. |
| `replace_text` | Replaces literal text inside the value. |
| `prefix` | Adds text before the current value. |
| `suffix` | Adds text after the current value. |

Example:

```yaml title="config.yaml excerpt"
filters:
  - description: "Clean event summaries"
    transform:
      summary:
        trim_prefix: "Roster - "
        replace_text:
          old: "Primary On Call"
          new: "On-Call"
          all: true
        prefix: "[Team A] "
```

:::warning
`replace_text.old` must not be empty.
:::
