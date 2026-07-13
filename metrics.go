package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	prometheusNamespace     = "ical_filter_proxy"
	prometheusHTTPSubsystem = "http"
)

type prometheusMetrics struct {
	registry        *prometheus.Registry
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func newPrometheusMetrics() *prometheusMetrics {
	registry := prometheus.NewRegistry()
	metrics := &prometheusMetrics{
		registry: registry,
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
	}

	registry.MustRegister(metrics.requestsTotal, metrics.requestDuration)

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
	})
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
