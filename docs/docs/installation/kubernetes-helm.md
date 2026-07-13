---
sidebar_position: 3
---

# Kubernetes / Helm

iCal Filter Proxy can be installed with the published Helm chart:

```bash
helm repo add yungwood https://yungwood.github.io/helm-charts/
helm install ical-filter-proxy yungwood/ical-filter-proxy
```

The source chart lives in the main repository under `chart/`. Tagged releases
publish chart updates to the `yungwood/helm-charts` repository.

iCal Filter Proxy serves plain HTTP. Use an ingress controller or load balancer
for HTTPS when needed. See [HTTP Behavior](../operations/http-behavior.md).

## Calendar Config

The chart can generate `config.yaml` from values:

In this chart, `config.calendars` is written into the generated application
`config.yaml`; other top-level keys configure Kubernetes resources.

```yaml title="values.yaml"
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

Install with the values file:

```bash
helm install ical-filter-proxy yungwood/ical-filter-proxy -f values.yaml
```

## Metrics

The chart enables the management listener by default. Health and readiness
probes use that listener, and the management service defaults to port `9090`.

:::note
The examples below rely on the default management listener settings.
:::

Enable Prometheus metrics and, if the Prometheus Operator is installed, create a
`ServiceMonitor`:

```yaml title="values.yaml"
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
    labels:
      release: prometheus
    interval: 30s
    scrapeTimeout: 10s
```

This exposes `/metrics` on the management service. The `ServiceMonitor` also
targets the management service.

## Secret-Backed Config Values

Use mounted secret files when upstream feed URLs or tokens should not be stored
directly in Helm values.

First create a Secret, or create an equivalent Secret with your usual secret
manager:

```bash
kubectl create secret generic calendar-secrets \
  --from-literal=calendar-token='changeme' \
  --from-literal=calendar-feed-url='https://example.com/private.ics'
```

Then mount that Secret and point the app config at the mounted file paths:

```yaml title="values.yaml"
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

:::info
The Secret keys are mounted as files, so this example creates:

```text
/run/secrets/calendar-token
/run/secrets/calendar-feed-url
```

Those paths are read by iCal Filter Proxy at startup.
:::

If you prefer environment variables, see
[Environment Substitution](../configuration/environment-substitution.md).

## Next Steps

- [Use secret files](../configuration/secret-files.md)
- [Enable metrics](../operations/metrics.md)
- [Review HTTP behavior](../operations/http-behavior.md)
