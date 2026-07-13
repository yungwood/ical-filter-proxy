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
		name     string
		listener string
		path     string
		want     string
	}{
		{
			name:     "calendar",
			listener: "public",
			path:     "/calendars/private/feed",
			want:     "/calendars/{calendar}/feed",
		},
		{
			name:     "liveness",
			listener: "management",
			path:     "/liveness",
			want:     "/liveness",
		},
		{
			name:     "readiness",
			listener: "management",
			path:     "/readiness",
			want:     "/readiness",
		},
		{
			name:     "metrics on management listener",
			listener: "management",
			path:     "/metrics",
			want:     "/metrics",
		},
		{
			name:     "metrics on public listener",
			listener: "public",
			path:     "/metrics",
			want:     "unknown",
		},
		{
			name:     "unknown",
			listener: "public",
			path:     "/other",
			want:     "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := routeMetricLabel(tt.listener, tt.path)
			if got != tt.want {
				t.Fatalf("routeMetricLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMetricsExposeBuildInfo(t *testing.T) {
	metrics := newPrometheusMetrics(false)

	got := scrapeMetrics(t, metrics)
	build := currentBuildInfo()
	want := `ical_filter_proxy_build_info{goversion="` + build.GoVersion + `",revision="` + build.Revision + `",version="` + build.Version + `"} 1`
	if !strings.Contains(got, want) {
		t.Fatalf("metrics output = %q, want build info sample %q", got, want)
	}
}

func TestCalendarNameFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantName string
		wantOK   bool
	}{
		{
			name:     "calendar feed",
			path:     "/calendars/private/feed",
			wantName: "private",
			wantOK:   true,
		},
		{
			name:   "not calendar route",
			path:   "/liveness",
			wantOK: false,
		},
		{
			name:   "missing feed suffix",
			path:   "/calendars/private",
			wantOK: false,
		},
		{
			name:   "extra path",
			path:   "/calendars/private/feed/extra",
			wantOK: false,
		},
		{
			name:   "empty name",
			path:   "/calendars//feed",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotOK := calendarNameFromPath(tt.path)
			if gotName != tt.wantName || gotOK != tt.wantOK {
				t.Fatalf("calendarNameFromPath() = %q, %v; want %q, %v", gotName, gotOK, tt.wantName, tt.wantOK)
			}
		})
	}
}

func TestMetricsMiddlewareRecordsRequest(t *testing.T) {
	metrics := newPrometheusMetrics(false)
	handler := metrics.middleware("public", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
	if !strings.Contains(got, `ical_filter_proxy_http_requests_total{listener="public",method="GET",route="/calendars/{calendar}/feed",status="202"} 1`) {
		t.Fatalf("metrics output = %q, want request counter sample", got)
	}
}

func TestMetricsMiddlewareRecordsPerCalendarRequestWhenEnabled(t *testing.T) {
	metrics := newPrometheusMetrics(true)
	handler := metrics.middleware("public", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	req := testRequest(t, "/calendars/private/feed?token=secret")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusAccepted)
	}

	got := scrapeMetrics(t, metrics)
	if strings.Contains(got, "token=secret") {
		t.Fatalf("metrics output includes query token: %q", got)
	}
	if !strings.Contains(got, `ical_filter_proxy_http_requests_total{listener="public",method="GET",route="/calendars/{calendar}/feed",status="202"} 1`) {
		t.Fatalf("metrics output = %q, want aggregate request counter sample", got)
	}
	if !strings.Contains(got, `ical_filter_proxy_calendar_requests_total{calendar="private",method="GET",status="202"} 1`) {
		t.Fatalf("metrics output = %q, want per-calendar request counter sample", got)
	}
}

func TestMetricsMiddlewareLabelsPublicMetricsPathAsUnknown(t *testing.T) {
	metrics := newPrometheusMetrics(false)
	handler := metrics.middleware("public", http.NotFoundHandler())

	req := testRequest(t, "/metrics")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}

	got := scrapeMetrics(t, metrics)
	if !strings.Contains(got, `ical_filter_proxy_http_requests_total{listener="public",method="GET",route="unknown",status="404"} 1`) {
		t.Fatalf("metrics output = %q, want public /metrics to be labelled unknown", got)
	}
	if strings.Contains(got, `listener="public",method="GET",route="/metrics"`) {
		t.Fatalf("metrics output = %q, did not want public /metrics labelled as metrics", got)
	}
}

func TestMetricsMiddlewareSkipsDurationForManagementRoutes(t *testing.T) {
	tests := []struct {
		name     string
		listener string
		target   string
		route    string
	}{
		{
			name:     "liveness",
			listener: "management",
			target:   "/liveness",
			route:    "/liveness",
		},
		{
			name:     "readiness",
			listener: "management",
			target:   "/readiness",
			route:    "/readiness",
		},
		{
			name:     "metrics",
			listener: "management",
			target:   "/metrics",
			route:    "/metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := newPrometheusMetrics(false)
			handler := metrics.middleware(tt.listener, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := testRequest(t, tt.target)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
			}

			got := scrapeMetrics(t, metrics)
			counterSample := `ical_filter_proxy_http_requests_total{listener="` + tt.listener + `",method="GET",route="` + tt.route + `",status="200"} 1`
			if !strings.Contains(got, counterSample) {
				t.Fatalf("metrics output = %q, want counter sample %q", got, counterSample)
			}

			durationSample := `ical_filter_proxy_http_request_duration_seconds_bucket{listener="` + tt.listener + `",method="GET",route="` + tt.route + `",status="200"`
			if strings.Contains(got, durationSample) {
				t.Fatalf("metrics output = %q, did not want duration sample %q", got, durationSample)
			}
		})
	}
}

func TestMetricsMiddlewareSkipsDurationForUnknownRoutes(t *testing.T) {
	metrics := newPrometheusMetrics(false)
	handler := metrics.middleware("public", http.NotFoundHandler())

	req := testRequest(t, "/unknown")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}

	got := scrapeMetrics(t, metrics)
	if !strings.Contains(got, `ical_filter_proxy_http_requests_total{listener="public",method="GET",route="unknown",status="404"} 1`) {
		t.Fatalf("metrics output = %q, want unknown route counter", got)
	}
	if strings.Contains(got, `ical_filter_proxy_http_request_duration_seconds_bucket{listener="public",method="GET",route="unknown",status="404"`) {
		t.Fatalf("metrics output = %q, did not want unknown route duration bucket", got)
	}
}

func TestMetricsInstrumentFetchRecordsSuccess(t *testing.T) {
	metrics := newPrometheusMetrics(false)
	fetch := metrics.instrumentFetch("private", func(context.Context) ([]byte, error) {
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

func TestMetricsInstrumentFetchRecordsPerCalendarFetchWhenEnabled(t *testing.T) {
	metrics := newPrometheusMetrics(true)
	fetch := metrics.instrumentFetch("private", func(context.Context) ([]byte, error) {
		return []byte("BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n"), nil
	})

	_, err := fetch(context.Background())
	if err != nil {
		t.Fatalf("fetch returned error: %v", err)
	}

	got := scrapeMetrics(t, metrics)
	if !strings.Contains(got, `ical_filter_proxy_upstream_fetches_total{result="success"} 1`) {
		t.Fatalf("metrics output = %q, want aggregate upstream fetch counter", got)
	}
	if !strings.Contains(got, `ical_filter_proxy_calendar_upstream_fetches_total{calendar="private",result="success"} 1`) {
		t.Fatalf("metrics output = %q, want per-calendar upstream fetch counter", got)
	}
}

func TestMetricsInstrumentFetchRecordsError(t *testing.T) {
	metrics := newPrometheusMetrics(false)
	fetch := metrics.instrumentFetch("private", func(context.Context) ([]byte, error) {
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
