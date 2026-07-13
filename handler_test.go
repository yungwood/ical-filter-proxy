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
		"/calendars/private/feed",
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
	if fetchCalled {
		t.Fatal("fetch was called for unauthorized request")
	}
}

func TestCalendarFeedHandlerSuccess(t *testing.T) {
	handler := calendarFeedHandlerWithFetch(
		"/calendars/private/feed",
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
	if got := rr.Header().Get("Content-Type"); got != "text/calendar" {
		t.Fatalf("Content-Type = %q, want text/calendar", got)
	}
	if got := rr.Body.String(); got != "BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n" {
		t.Fatalf("body = %q, want calendar feed", got)
	}
}

func TestCalendarFeedHandlerUpstreamError(t *testing.T) {
	handler := calendarFeedHandlerWithFetch(
		"/calendars/private/feed",
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
}

func testRequest(t *testing.T, target string) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() returned error: %v", err)
	}

	return req
}
