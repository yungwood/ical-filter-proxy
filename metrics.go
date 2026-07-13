package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	prometheusNamespace         = "ical_filter_proxy"
	prometheusHTTPSubsystem     = "http"
	prometheusUpstreamSubsystem = "upstream"
	prometheusCalendarSubsystem = "calendar"
)

type prometheusMetrics struct {
	registry                      *prometheus.Registry
	buildInfo                     *prometheus.GaugeVec
	requestsTotal                 *prometheus.CounterVec
	requestDuration               *prometheus.HistogramVec
	upstreamFetchesTotal          *prometheus.CounterVec
	upstreamFetchDuration         *prometheus.HistogramVec
	calendarMetricsEnabled        bool
	calendarRequestsTotal         *prometheus.CounterVec
	calendarRequestDuration       *prometheus.HistogramVec
	calendarUpstreamFetchesTotal  *prometheus.CounterVec
	calendarUpstreamFetchDuration *prometheus.HistogramVec
}

func newPrometheusMetrics(calendarMetricsEnabled bool) *prometheusMetrics {
	registry := prometheus.NewRegistry()
	metrics := &prometheusMetrics{
		registry:               registry,
		calendarMetricsEnabled: calendarMetricsEnabled,
		buildInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: prometheusNamespace,
				Name:      "build_info",
				Help:      "Build information for ical-filter-proxy.",
			},
			[]string{"version", "revision", "goversion"},
		),
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: prometheusNamespace,
				Subsystem: prometheusHTTPSubsystem,
				Name:      "requests_total",
				Help:      "Total number of HTTP requests.",
			},
			[]string{"handler", "method", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: prometheusNamespace,
				Subsystem: prometheusHTTPSubsystem,
				Name:      "request_duration_seconds",
				Help:      "Duration of HTTP requests in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"handler", "method", "status"},
		),
		upstreamFetchesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: prometheusNamespace,
				Subsystem: prometheusUpstreamSubsystem,
				Name:      "fetches_total",
				Help:      "Total number of upstream calendar fetches.",
			},
			[]string{"result"},
		),
		upstreamFetchDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: prometheusNamespace,
				Subsystem: prometheusUpstreamSubsystem,
				Name:      "fetch_duration_seconds",
				Help:      "Duration of upstream calendar fetches in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"result"},
		),
	}

	build := currentBuildInfo()
	metrics.buildInfo.WithLabelValues(build.Version, build.Revision, build.GoVersion).Set(1)

	collectors := []prometheus.Collector{
		metrics.buildInfo,
		metrics.requestsTotal,
		metrics.requestDuration,
		metrics.upstreamFetchesTotal,
		metrics.upstreamFetchDuration,
	}

	if calendarMetricsEnabled {
		metrics.calendarRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: prometheusNamespace,
				Subsystem: prometheusCalendarSubsystem,
				Name:      "requests_total",
				Help:      "Total number of per-calendar HTTP requests.",
			},
			[]string{"calendar", "method", "status"},
		)
		metrics.calendarRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: prometheusNamespace,
				Subsystem: prometheusCalendarSubsystem,
				Name:      "request_duration_seconds",
				Help:      "Duration of per-calendar HTTP requests in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"calendar", "method", "status"},
		)
		metrics.calendarUpstreamFetchesTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: prometheusNamespace,
				Subsystem: prometheusCalendarSubsystem,
				Name:      "upstream_fetches_total",
				Help:      "Total number of per-calendar upstream fetches.",
			},
			[]string{"calendar", "result"},
		)
		metrics.calendarUpstreamFetchDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: prometheusNamespace,
				Subsystem: prometheusCalendarSubsystem,
				Name:      "upstream_fetch_duration_seconds",
				Help:      "Duration of per-calendar upstream fetches in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"calendar", "result"},
		)

		collectors = append(
			collectors,
			metrics.calendarRequestsTotal,
			metrics.calendarRequestDuration,
			metrics.calendarUpstreamFetchesTotal,
			metrics.calendarUpstreamFetchDuration,
		)
	}

	registry.MustRegister(collectors...)

	return metrics
}

func (m *prometheusMetrics) handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *prometheusMetrics) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		statusCode := recorder.statusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		labels := prometheus.Labels{
			"handler": routeMetricLabel(r.URL.Path),
			"method":  r.Method,
			"status":  strconv.Itoa(statusCode),
		}
		m.requestsTotal.With(labels).Inc()
		m.requestDuration.With(labels).Observe(time.Since(startedAt).Seconds())

		if calendarName, ok := calendarNameFromPath(r.URL.Path); ok && m.calendarMetricsEnabled {
			calendarLabels := prometheus.Labels{
				"calendar": calendarName,
				"method":   r.Method,
				"status":   strconv.Itoa(statusCode),
			}
			m.calendarRequestsTotal.With(calendarLabels).Inc()
			m.calendarRequestDuration.With(calendarLabels).Observe(time.Since(startedAt).Seconds())
		}
	})
}

func (m *prometheusMetrics) instrumentFetch(calendarName string, fetch calendarFetchFunc) calendarFetchFunc {
	return func(ctx context.Context) ([]byte, error) {
		startedAt := time.Now()
		feed, err := fetch(ctx)

		result := "success"
		if err != nil {
			result = "error"
		}

		labels := prometheus.Labels{"result": result}
		m.upstreamFetchesTotal.With(labels).Inc()
		m.upstreamFetchDuration.With(labels).Observe(time.Since(startedAt).Seconds())

		if m.calendarMetricsEnabled {
			calendarLabels := prometheus.Labels{
				"calendar": calendarName,
				"result":   result,
			}
			m.calendarUpstreamFetchesTotal.With(calendarLabels).Inc()
			m.calendarUpstreamFetchDuration.With(calendarLabels).Observe(time.Since(startedAt).Seconds())
		}

		return feed, err
	}
}

func routeMetricLabel(path string) string {
	switch {
	case strings.HasPrefix(path, "/calendars/"):
		return "calendar"
	case path == "/liveness":
		return "liveness"
	case path == "/readiness":
		return "readiness"
	case path == "/metrics":
		return "metrics"
	default:
		return "unknown"
	}
}

func calendarNameFromPath(path string) (string, bool) {
	const prefix = "/calendars/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}

	remainingPath := strings.TrimPrefix(path, prefix)
	calendarName, remainingPath, found := strings.Cut(remainingPath, "/")
	if !found || calendarName == "" || remainingPath != "feed" {
		return "", false
	}

	return calendarName, true
}
