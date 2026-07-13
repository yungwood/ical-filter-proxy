package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTokenMatches(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
		want     bool
	}{
		{
			name:     "matching token",
			token:    "secret",
			expected: "secret",
			want:     true,
		},
		{
			name:     "different token",
			token:    "wrong",
			expected: "secret",
			want:     false,
		},
		{
			name:     "empty token matches empty expected token",
			token:    "",
			expected: "",
			want:     true,
		},
		{
			name:     "empty token rejects non-empty expected token",
			token:    "",
			expected: "secret",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenMatches(tt.token, tt.expected)
			if got != tt.want {
				t.Fatalf("tokenMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalendarFeedHandlerUnauthorized(t *testing.T) {
	fetchCalled := false
	handler := calendarFeedHandlerWithFetch(
		CalendarConfig{Name: "private", Token: "secret"},
		func(context.Context) ([]byte, error) {
			fetchCalled = true
			return []byte("should not be called"), nil
		},
	)

	req := testRequest(t, "/calendars/private/feed?token=wrong")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
	assertCommonResponseHeaders(t, rr)
	if fetchCalled {
		t.Fatal("fetch was called for unauthorized request")
	}
}

func TestCalendarFeedHandlerSuccess(t *testing.T) {
	handler := calendarFeedHandlerWithFetch(
		CalendarConfig{Name: "private", Token: "secret"},
		func(context.Context) ([]byte, error) {
			return []byte("BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n"), nil
		},
	)

	req := testRequest(t, "/calendars/private/feed?token=secret")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	assertCommonResponseHeaders(t, rr)
	if got := rr.Header().Get("Content-Type"); got != calendarContentType {
		t.Fatalf("Content-Type = %q, want %q", got, calendarContentType)
	}
	if got := rr.Body.String(); got != "BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n" {
		t.Fatalf("body = %q, want calendar feed", got)
	}
}

func TestCalendarFeedHandlerHeadSuccess(t *testing.T) {
	handler := calendarFeedHandlerWithFetch(
		CalendarConfig{Name: "private", Token: "secret"},
		func(context.Context) ([]byte, error) {
			return []byte("BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n"), nil
		},
	)

	req := testRequestWithMethod(t, http.MethodHead, "/calendars/private/feed?token=secret")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	assertCommonResponseHeaders(t, rr)
	if got := rr.Header().Get("Content-Type"); got != calendarContentType {
		t.Fatalf("Content-Type = %q, want %q", got, calendarContentType)
	}
	if got := rr.Body.String(); got != "" {
		t.Fatalf("body = %q, want empty body", got)
	}
}

func TestCalendarFeedHandlerMethodNotAllowed(t *testing.T) {
	fetchCalled := false
	handler := calendarFeedHandlerWithFetch(
		CalendarConfig{Name: "private", Token: "secret"},
		func(context.Context) ([]byte, error) {
			fetchCalled = true
			return []byte("should not be called"), nil
		},
	)

	req := testRequestWithMethod(t, http.MethodPost, "/calendars/private/feed?token=secret")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
	assertCommonResponseHeaders(t, rr)
	if got := rr.Header().Get("Allow"); got != allowedMethods {
		t.Fatalf("Allow = %q, want %q", got, allowedMethods)
	}
	if fetchCalled {
		t.Fatal("fetch was called for unsupported method")
	}
}

func TestCalendarFeedHandlerPublicCalendarIgnoresToken(t *testing.T) {
	tests := []struct {
		name   string
		target string
	}{
		{
			name:   "no token",
			target: "/calendars/public/feed",
		},
		{
			name:   "wrong token",
			target: "/calendars/public/feed?token=wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := calendarFeedHandlerWithFetch(
				CalendarConfig{Name: "public", Public: true},
				func(context.Context) ([]byte, error) {
					return []byte("BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n"), nil
				},
			)

			req := testRequest(t, tt.target)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
			}
		})
	}
}

func TestCalendarFeedHandlerUpstreamError(t *testing.T) {
	handler := calendarFeedHandlerWithFetch(
		CalendarConfig{Name: "private", Token: "secret"},
		func(context.Context) ([]byte, error) {
			return nil, errors.New("upstream failed")
		},
	)

	req := testRequest(t, "/calendars/private/feed?token=secret")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadGateway)
	}
	assertCommonResponseHeaders(t, rr)
}

func testRequest(t *testing.T, target string) *http.Request {
	t.Helper()

	return testRequestWithMethod(t, http.MethodGet, target)
}

func testRequestWithMethod(t *testing.T, method string, target string) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), method, target, nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() returned error: %v", err)
	}

	return req
}

func assertCommonResponseHeaders(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	if got := rr.Header().Get("Cache-Control"); got != cacheControlHeader {
		t.Fatalf("Cache-Control = %q, want %q", got, cacheControlHeader)
	}
	if got := rr.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", got)
	}
}
