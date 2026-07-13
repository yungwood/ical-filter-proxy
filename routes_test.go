package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterPublicRoutesRegistersCalendarFeeds(t *testing.T) {
	mux := http.NewServeMux()
	registerPublicRoutes(mux, Config{
		Calendars: []CalendarConfig{
			{
				Name:  "private",
				Token: "secret",
			},
		},
	}, nil)

	req := testRequest(t, "/calendars/private/feed")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestRegisterInternalRoutesRegistersHealthEndpoints(t *testing.T) {
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

	mux := http.NewServeMux()
	registerInternalRoutes(mux, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := testRequest(t, tt.path)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
			}
		})
	}
}

func TestPublicAndInternalRoutesCanUseSeparateMuxes(t *testing.T) {
	publicMux := http.NewServeMux()
	registerPublicRoutes(publicMux, Config{
		Calendars: []CalendarConfig{
			{
				Name:  "private",
				Token: "secret",
			},
		},
	}, nil)

	internalMux := http.NewServeMux()
	registerInternalRoutes(internalMux, nil)

	req := testRequest(t, "/liveness")
	rr := httptest.NewRecorder()
	publicMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("public mux status = %d, want %d", rr.Code, http.StatusNotFound)
	}

	req = testRequest(t, "/calendars/private/feed")
	rr = httptest.NewRecorder()
	internalMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("internal mux status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestRegisterInternalRoutesMetricsDisabled(t *testing.T) {
	mux := http.NewServeMux()
	registerInternalRoutes(mux, nil)

	req := testRequest(t, "/metrics")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestRegisterInternalRoutesMetricsEnabled(t *testing.T) {
	mux := http.NewServeMux()
	registerInternalRoutes(mux, newPrometheusMetrics(false))

	req := testRequest(t, "/metrics")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}
