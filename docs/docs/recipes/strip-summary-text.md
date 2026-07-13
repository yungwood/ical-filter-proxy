---
sidebar_position: 6
---

# Strip Summary Text

## Use Case

Use string transforms to remove provider noise while keeping the useful part of
the event summary.

## Config

```yaml title="config.yaml excerpt"
filters:
  - description: "Clean summary text"
    transform:
      summary:
        trim_prefix: "Roster - "
        trim_suffix: " - confirmed"
        replace_text:
          old: "Primary On Call"
          new: "On-Call"
          all: true
```

## How It Works

Transforms compose in a fixed order, so this trims first, replaces text next,
and then applies any configured prefix or suffix.
