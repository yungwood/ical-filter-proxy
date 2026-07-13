package main

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStatusRecorderCapturesExplicitStatus(t *testing.T) {
	rr := httptest.NewRecorder()
	recorder := &statusRecorder{ResponseWriter: rr}

	recorder.WriteHeader(http.StatusTeapot)

	if recorder.statusCode != http.StatusTeapot {
		t.Fatalf("statusCode = %d, want %d", recorder.statusCode, http.StatusTeapot)
	}
	if rr.Code != http.StatusTeapot {
		t.Fatalf("response code = %d, want %d", rr.Code, http.StatusTeapot)
	}
}

func TestStatusRecorderCapturesImplicitOK(t *testing.T) {
	rr := httptest.NewRecorder()
	recorder := &statusRecorder{ResponseWriter: rr}

	_, err := recorder.Write([]byte("ok"))
	if err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}

	if recorder.statusCode != http.StatusOK {
		t.Fatalf("statusCode = %d, want %d", recorder.statusCode, http.StatusOK)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("response code = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRequestLoggingMiddlewareLogsPathWithoutQuery(t *testing.T) {
	var logOutput bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logOutput, nil)))
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	handler := requestLoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}), nil)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/calendars/private/feed?token=secret", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() returned error: %v", err)
	}
	req.RemoteAddr = "192.0.2.1:12345"

	handler.ServeHTTP(httptest.NewRecorder(), req)

	got := logOutput.String()
	if !strings.Contains(got, "path=/calendars/private/feed") {
		t.Fatalf("log output = %q, want path without query", got)
	}
	if strings.Contains(got, "token=secret") {
		t.Fatalf("log output contains query token: %q", got)
	}
	if !strings.Contains(got, "status=202") {
		t.Fatalf("log output = %q, want status", got)
	}
	if !strings.Contains(got, "calendar=private") {
		t.Fatalf("log output = %q, want calendar name", got)
	}
}

func TestRequestLoggingMiddlewareSkipsHealthEndpoints(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "liveness",
			path: "/liveness",
		},
		{
			name: "readiness",
			path: "/readiness",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logOutput bytes.Buffer
			originalLogger := slog.Default()
			slog.SetDefault(slog.New(slog.NewTextHandler(&logOutput, nil)))
			t.Cleanup(func() {
				slog.SetDefault(originalLogger)
			})

			handler := requestLoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}), nil)

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tt.path, nil)
			if err != nil {
				t.Fatalf("NewRequestWithContext() returned error: %v", err)
			}

			handler.ServeHTTP(httptest.NewRecorder(), req)

			if got := logOutput.String(); got != "" {
				t.Fatalf("log output = %q, want empty", got)
			}
		})
	}
}

func TestRequestLoggingMiddlewareLogsFailedHealthEndpoints(t *testing.T) {
	var logOutput bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logOutput, nil)))
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	handler := requestLoggingMiddleware(http.NotFoundHandler(), nil)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/readiness", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() returned error: %v", err)
	}
	req.RemoteAddr = "192.0.2.1:12345"

	handler.ServeHTTP(httptest.NewRecorder(), req)

	got := logOutput.String()
	if !strings.Contains(got, "path=/readiness") {
		t.Fatalf("log output = %q, want readiness path", got)
	}
	if !strings.Contains(got, "status=404") {
		t.Fatalf("log output = %q, want 404 status", got)
	}
	if !strings.Contains(got, "client_addr=192.0.2.1") {
		t.Fatalf("log output = %q, want client address", got)
	}
}

func TestRequestLoggingMiddlewareLogsNonHealthEndpoint(t *testing.T) {
	var logOutput bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logOutput, nil)))
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	handler := requestLoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), nil)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/liveness/extra", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() returned error: %v", err)
	}

	handler.ServeHTTP(httptest.NewRecorder(), req)

	got := logOutput.String()
	if !strings.Contains(got, "path=/liveness/extra") {
		t.Fatalf("log output = %q, want non-health endpoint to be logged", got)
	}
	if strings.Contains(got, "calendar=") {
		t.Fatalf("log output = %q, did not want calendar attribute", got)
	}
}

func TestRecoveryMiddlewareReturnsInternalServerError(t *testing.T) {
	var logOutput bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logOutput, nil)))
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	handler := requestLoggingMiddleware(recoveryMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	}), nil), nil)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/calendars/private/feed", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() returned error: %v", err)
	}
	req.RemoteAddr = "192.0.2.1:12345"

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}

	got := logOutput.String()
	if !strings.Contains(got, "recovered panic while processing http request") {
		t.Fatalf("log output = %q, want recovery log", got)
	}
	if !strings.Contains(got, "panic=boom") {
		t.Fatalf("log output = %q, want panic value", got)
	}
	if !strings.Contains(got, "method=GET") {
		t.Fatalf("log output = %q, want request method", got)
	}
	if !strings.Contains(got, "calendar=private") {
		t.Fatalf("log output = %q, want calendar name", got)
	}
	if !strings.Contains(got, "status=500") {
		t.Fatalf("log output = %q, want request log status", got)
	}
}

func TestRecoveryMiddlewarePassesThrough(t *testing.T) {
	handler := recoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), nil)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/calendars/private/feed", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() returned error: %v", err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
}
