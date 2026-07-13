---
sidebar_position: 3
---

# Metrics Reference

Metrics use the `ical_filter_proxy` namespace.

Always available when metrics are enabled:

| Metric | Labels | Description |
| --- | --- | --- |
| `ical_filter_proxy_build_info` | `version`, `revision`, `goversion` | Build metadata. |
| `ical_filter_proxy_http_requests_total` | `listener`, `route`, `method`, `status` | HTTP request count. |
| `ical_filter_proxy_http_request_duration_seconds` | `listener`, `route`, `method`, `status` | Calendar HTTP request duration histogram. |
| `ical_filter_proxy_upstream_fetches_total` | `result` | Upstream fetch count. |
| `ical_filter_proxy_upstream_fetch_duration_seconds` | `result` | Upstream fetch duration histogram. |

Additional metrics when `-metrics-calendar-labels` is enabled:

| Metric | Labels | Description |
| --- | --- | --- |
| `ical_filter_proxy_calendar_requests_total` | `calendar`, `method`, `status` | Per-calendar HTTP request count. |
| `ical_filter_proxy_calendar_request_duration_seconds` | `calendar`, `method`, `status` | Per-calendar HTTP request duration histogram. |
| `ical_filter_proxy_calendar_upstream_fetches_total` | `calendar`, `result` | Per-calendar upstream fetch count. |
| `ical_filter_proxy_calendar_upstream_fetch_duration_seconds` | `calendar`, `result` | Per-calendar upstream fetch duration histogram. |

:::info
HTTP route labels are normalized to avoid exposing arbitrary paths, calendar
names, or tokens as label values.
:::
