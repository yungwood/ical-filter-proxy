---
sidebar_position: 2
---

# Config Reference

Top-level config:

| Field | Type | Description |
| --- | --- | --- |
| `calendars` | list | Calendar definitions. At least one calendar is required. |

Calendar fields:

| Field | Type | Description |
| --- | --- | --- |
| `name` | string | Calendar slug used in the feed URL. |
| `publish_name` | string | Calendar name written to the output feed. |
| `public` | boolean | Allows access without a token. |
| `token` | string | Token required for private calendar access. |
| `token_file` | string | File containing the access token. |
| `feed_url` | string | Upstream iCalendar URL. |
| `feed_url_file` | string | File containing the upstream iCalendar URL. |
| `user_agent` | string | User-Agent override for upstream requests. |
| `filters` | list | Ordered filter definitions. |

Filter fields:

| Field | Type | Description |
| --- | --- | --- |
| `description` | string | Human-readable filter description used in logs. |
| `remove` | boolean | Removes matching events. |
| `stop` | boolean | Stops later filters after this filter matches. |
| `match` | object | Event match rules. |
| `transform` | object | Event transform rules. |

Match and transform rules currently support `summary`, `description`,
`location`, and `url`.

Match target fields:

| Field | Type | Description |
| --- | --- | --- |
| `summary` | object | Match against the event summary/title. |
| `description` | object | Match against the event description. |
| `location` | object | Match against the event location. |
| `url` | object | Match against the event URL. |

String match conditions:

| Field | Type | Description |
| --- | --- | --- |
| `empty` | boolean | Matches when the property is absent or empty. When set, this condition is evaluated on its own. |
| `contains` | string | Matches when the property contains this literal text. |
| `contains_any` | list of strings | Matches when the property contains at least one listed value. Empty list values are rejected. |
| `contains_all` | list of strings | Matches when the property contains every listed value. Empty list values are rejected. |
| `prefix` | string | Matches when the property starts with this literal text. |
| `suffix` | string | Matches when the property ends with this literal text. |
| `regex` | string | Matches using a Go regular expression. Invalid regexes fail config compilation. |

Transform target fields:

| Field | Type | Description |
| --- | --- | --- |
| `summary` | object | Transform the event summary/title. |
| `description` | object | Transform the event description. |
| `location` | object | Transform the event location. |
| `url` | object | Transform the event URL. |

String transform operations:

| Field | Type | Description |
| --- | --- | --- |
| `remove` | boolean | Sets the property to a blank string. This takes precedence over other transform operations. |
| `replace` | string | Replaces the entire property value. This takes precedence over trim, text replacement, prefix, and suffix operations. |
| `trim_prefix` | string | Removes this text when the property starts with it. |
| `trim_suffix` | string | Removes this text when the property ends with it. |
| `replace_text` | object | Replaces literal text inside the property value. |
| `prefix` | string | Adds text before the current property value. |
| `suffix` | string | Adds text after the current property value. |

`replace_text` fields:

| Field | Type | Description |
| --- | --- | --- |
| `old` | string | Literal text to replace. Required when `replace_text` is used. |
| `new` | string | Replacement text. |
| `all` | boolean | Replaces every occurrence when `true`; otherwise only the first occurrence is replaced. |

When `remove` and `replace` are not set, string transforms run in this order:
`trim_prefix`, `trim_suffix`, `replace_text`, `prefix`, `suffix`.
