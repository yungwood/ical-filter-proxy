package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouteMetricLabel(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "calendar",
			path: "/calendars/private/feed",
			want: "calendar",
		},
		{
			name: "liveness",
			path: "/liveness",
			want: "liveness",
		},
		{
			name: "readiness",
			path: "/readiness",
			want: "readiness",
		},
		{
			name: "metrics",
			path: "/metrics",
			want: "metrics",
		},
		{
			name: "unknown",
			path: "/other",
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := routeMetricLabel(tt.path)
			if got != tt.want {
				t.Fatalf("routeMetricLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMetricsMiddlewareRecordsRequest(t *testing.T) {
	metrics := newPrometheusMetrics()
	handler := metrics.middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	req := testRequest(t, "/calendars/private/feed?token=secret")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusAccepted)
	}

	metricsRecorder := httptest.NewRecorder()
	metrics.handler().ServeHTTP(metricsRecorder, testRequest(t, "/metrics"))

	got := metricsRecorder.Body.String()
	if strings.Contains(got, "private") || strings.Contains(got, "token=secret") {
		t.Fatalf("metrics output includes sensitive route data: %q", got)
	}
	if !strings.Contains(got, `ical_filter_proxy_http_requests_total{handler="calendar",method="GET",status="202"} 1`) {
		t.Fatalf("metrics output = %q, want request counter sample", got)
	}
}

func TestMetricsInstrumentFetchRecordsSuccess(t *testing.T) {
	metrics := newPrometheusMetrics()
	fetch := metrics.instrumentFetch(func(context.Context) ([]byte, error) {
		return []byte("BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n"), nil
	})

	feed, err := fetch(context.Background())
	if err != nil {
		t.Fatalf("fetch returned error: %v", err)
	}
	if string(feed) != "BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n" {
		t.Fatalf("feed = %q, want calendar", feed)
	}

	got := scrapeMetrics(t, metrics)
	if !strings.Contains(got, `ical_filter_proxy_upstream_fetches_total{result="success"} 1`) {
		t.Fatalf("metrics output = %q, want successful upstream fetch counter", got)
	}
	if strings.Contains(got, `ical_filter_proxy_upstream_fetches_total{result="error"}`) {
		t.Fatalf("metrics output = %q, did not want error upstream fetch counter", got)
	}
}

func TestMetricsInstrumentFetchRecordsError(t *testing.T) {
	metrics := newPrometheusMetrics()
	fetch := metrics.instrumentFetch(func(context.Context) ([]byte, error) {
		return nil, errors.New("upstream failed")
	})

	_, err := fetch(context.Background())
	if err == nil {
		t.Fatal("fetch returned nil error")
	}

	got := scrapeMetrics(t, metrics)
	if !strings.Contains(got, `ical_filter_proxy_upstream_fetches_total{result="error"} 1`) {
		t.Fatalf("metrics output = %q, want failed upstream fetch counter", got)
	}
	if strings.Contains(got, `ical_filter_proxy_upstream_fetches_total{result="success"}`) {
		t.Fatalf("metrics output = %q, did not want successful upstream fetch counter", got)
	}
}

func scrapeMetrics(t *testing.T, metrics *prometheusMetrics) string {
	t.Helper()

	metricsRecorder := httptest.NewRecorder()
	metrics.handler().ServeHTTP(metricsRecorder, testRequest(t, "/metrics"))

	return metricsRecorder.Body.String()
}
