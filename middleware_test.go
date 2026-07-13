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
	}))

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
			}))

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

func TestRequestLoggingMiddlewareLogsNonHealthEndpoint(t *testing.T) {
	var logOutput bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logOutput, nil)))
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	handler := requestLoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/liveness/extra", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() returned error: %v", err)
	}

	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got := logOutput.String(); !strings.Contains(got, "path=/liveness/extra") {
		t.Fatalf("log output = %q, want non-health endpoint to be logged", got)
	}
}
