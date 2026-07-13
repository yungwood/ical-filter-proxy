package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/netip"
	"runtime/debug"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}

	return r.ResponseWriter.Write(body)
}

func requestLoggingMiddleware(next http.Handler, trustedProxyCIDRs []netip.Prefix) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		statusCode := recorder.statusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}
		if isHealthEndpoint(r.URL.Path) && statusCode < http.StatusBadRequest {
			return
		}

		logArgs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", statusCode,
			"duration_ms", time.Since(startedAt).Milliseconds(),
			"client_addr", requestClientAddr(r, trustedProxyCIDRs),
			"user_agent", r.UserAgent(),
		}
		if calendarName, ok := calendarNameFromPath(r.URL.Path); ok {
			logArgs = append(logArgs, "calendar", calendarName)
		}

		slog.Info("http request processed", logArgs...)
	})
}

func recoveryMiddleware(next http.Handler, trustedProxyCIDRs []netip.Prefix) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logArgs := []any{
					"panic", fmt.Sprint(recovered),
					"method", r.Method,
					"path", r.URL.Path,
					"client_addr", requestClientAddr(r, trustedProxyCIDRs),
					"stack", string(debug.Stack()),
				}
				if calendarName, ok := calendarNameFromPath(r.URL.Path); ok {
					logArgs = append(logArgs, "calendar", calendarName)
				}

				slog.Error("recovered panic while processing http request", logArgs...)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func isHealthEndpoint(path string) bool {
	return path == "/liveness" || path == "/readiness"
}
