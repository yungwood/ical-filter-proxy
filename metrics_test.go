package main

import (
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
