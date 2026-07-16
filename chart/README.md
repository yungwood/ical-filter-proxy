# iCal Filter Proxy Helm Chart

This chart installs
[iCal Filter Proxy](https://github.com/yungwood/ical-filter-proxy), a service
for proxying upstream iCalendar feeds and applying filtering or transform rules
before publishing them.

Full documentation is available at
<https://yungwood.github.io/ical-filter-proxy/>.

## Install

```bash
helm repo add yungwood https://yungwood.github.io/helm-charts/
helm install ical-filter-proxy yungwood/ical-filter-proxy
```

## Basic Values

The chart can generate the application `config.yaml` from
`config.calendars`:

```yaml
config:
  enabled: true
  calendars:
    - name: example
      publish_name: "Example Calendar"
      public: true
      feed_url: "https://example.com/calendar.ics"
      filters:
        - description: "Remove cancelled events"
          remove: true
          match:
            summary:
              prefix: "Canceled: "
```

Install with:

```bash
helm install ical-filter-proxy yungwood/ical-filter-proxy -f values.yaml
```

## Management And Metrics

The management listener is enabled by default. Health and readiness probes use
that listener, and the management service defaults to port `9090`.

Enable Prometheus metrics:

```yaml
metrics:
  enabled: true
```

Enable a Prometheus Operator `ServiceMonitor`:

```yaml
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
    labels:
      release: prometheus
```

## Secret-Backed Config

Mount Kubernetes Secrets as files and reference them from calendar config:

```yaml
config:
  calendars:
    - name: private
      token_file: /run/secrets/calendar-token
      feed_url_file: /run/secrets/calendar-feed-url

volumes:
  - name: calendar-secrets
    secret:
      secretName: calendar-secrets

volumeMounts:
  - name: calendar-secrets
    mountPath: /run/secrets
    readOnly: true
```

Use `env` and `envFrom` when values should come from environment variables:

```yaml
env:
  - name: ICAL_FILTER_PROXY_TRUSTED_PROXY_CIDRS
    value: 10.0.0.0/8
envFrom:
  - secretRef:
      name: ical-filter-proxy-env
```

## Common Values

| Value                            | Default                      | Description                                  |
| -------------------------------- | ---------------------------- | -------------------------------------------- |
| `image.repository`               | `yungwood/ical-filter-proxy` | Image repository.                            |
| `image.tag`                      | chart `appVersion`           | Image tag.                                   |
| `replicaCount`                   | `1`                          | Number of replicas.                          |
| `revisionHistoryLimit`           | `3`                          | Deployment revision history limit.           |
| `app.address`                    | `:<service.port>`            | Public calendar listener address.            |
| `config.enabled`                 | `true`                       | Generate application config from values.     |
| `config.calendars`               | example public calendar      | Calendar config written to `config.yaml`.    |
| `service.type`                   | `ClusterIP`                  | Public service type.                         |
| `service.port`                   | `8080`                       | Public service port.                         |
| `management.enabled`             | `true`                       | Enable separate management listener.         |
| `management.service.enabled`     | `true`                       | Create management service.                   |
| `management.service.port`        | `9090`                       | Management service port.                     |
| `metrics.enabled`                | `false`                      | Enable Prometheus metrics endpoint.          |
| `metrics.calendarLabels`         | `false`                      | Enable per-calendar metric labels.           |
| `metrics.serviceMonitor.enabled` | `false`                      | Create Prometheus Operator `ServiceMonitor`. |
| `ingress.enabled`                | `false`                      | Create ingress.                              |
| `env`                            | `[]`                         | Extra environment variables.                 |
| `envFrom`                        | `[]`                         | Extra environment sources.                   |
| `volumes`                        | `[]`                         | Extra pod volumes.                           |
| `volumeMounts`                   | `[]`                         | Extra container volume mounts.               |
| `extraResources`                 | `[]`                         | Additional templated manifests.              |

For the full Kubernetes guide, see
<https://yungwood.github.io/ical-filter-proxy/installation/kubernetes-helm/>.
