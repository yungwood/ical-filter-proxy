package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchUpstreamCalendarSuccess(t *testing.T) {
	body := "BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "ical-filter-proxy/"+version {
			t.Fatalf("User-Agent = %q, want %q", got, "ical-filter-proxy/"+version)
		}
		if got := r.Header.Get("Accept"); got != "text/calendar, text/plain;q=0.8, */*;q=0.1" {
			t.Fatalf("Accept = %q, want text/calendar preference", got)
		}
		w.Header().Set("Content-Type", "text/calendar")
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	got, err := fetchUpstreamCalendar(context.Background(), server.URL, "")
	if err != nil {
		t.Fatalf("fetchUpstreamCalendar returned error: %v", err)
	}
	if string(got) != body {
		t.Fatalf("body = %q, want %q", string(got), body)
	}
}

func TestFetchUpstreamCalendarUsesCustomUserAgent(t *testing.T) {
	body := "BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "custom-calendar-client/1.0" {
			t.Fatalf("User-Agent = %q, want custom-calendar-client/1.0", got)
		}
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	got, err := fetchUpstreamCalendar(context.Background(), server.URL, "custom-calendar-client/1.0")
	if err != nil {
		t.Fatalf("fetchUpstreamCalendar returned error: %v", err)
	}
	if string(got) != body {
		t.Fatalf("body = %q, want %q", string(got), body)
	}
}

func TestUpstreamUserAgent(t *testing.T) {
	if got, want := upstreamUserAgent("custom-calendar-client/1.0"), "custom-calendar-client/1.0"; got != want {
		t.Fatalf("upstreamUserAgent(custom) = %q, want %q", got, want)
	}
	if got, want := upstreamUserAgent(""), "ical-filter-proxy/"+version; got != want {
		t.Fatalf("upstreamUserAgent(empty) = %q, want %q", got, want)
	}
}

func TestFetchUpstreamCalendarRejectsNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "upstream unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, err := fetchUpstreamCalendar(context.Background(), server.URL, "")
	if err == nil {
		t.Fatal("fetchUpstreamCalendar returned nil error")
	}
	if !strings.Contains(err.Error(), "upstream returned 503 Service Unavailable") {
		t.Fatalf("error = %q, want upstream status", err.Error())
	}
}

func TestFetchUpstreamCalendarRejectsOversizedBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(bytes.Repeat([]byte("x"), maxCalendarBytes+1))
	}))
	defer server.Close()

	_, err := fetchUpstreamCalendar(context.Background(), server.URL, "")
	if err == nil {
		t.Fatal("fetchUpstreamCalendar returned nil error")
	}
	if !strings.Contains(err.Error(), "upstream calendar exceeds") {
		t.Fatalf("error = %q, want size limit error", err.Error())
	}
}

func TestFetchUpstreamCalendarUsesContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("server should not receive request for canceled context")
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := fetchUpstreamCalendar(ctx, server.URL, "")
	if err == nil {
		t.Fatal("fetchUpstreamCalendar returned nil error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestRedactURL(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   string
	}{
		{
			name:   "redacts path token",
			rawURL: "https://calendar.example.com/secret/path/feed.ics",
			want:   "https://calendar.example.com/...",
		},
		{
			name:   "preserves query presence",
			rawURL: "https://calendar.example.com/secret/path/feed.ics?token=secret",
			want:   "https://calendar.example.com/...?REDACTED",
		},
		{
			name:   "preserves fragment presence",
			rawURL: "https://calendar.example.com/secret/path/feed.ics#secret",
			want:   "https://calendar.example.com/...#REDACTED",
		},
		{
			name:   "preserves query and fragment presence",
			rawURL: "https://calendar.example.com/secret/path/feed.ics?token=secret#secret",
			want:   "https://calendar.example.com/...?REDACTED#REDACTED",
		},
		{
			name:   "preserves host port",
			rawURL: "http://localhost:8080/secret",
			want:   "http://localhost:8080/...",
		},
		{
			name:   "rejects relative url",
			rawURL: "/secret/path",
			want:   "<invalid-url>",
		},
		{
			name:   "rejects malformed url",
			rawURL: "://bad-url",
			want:   "<invalid-url>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactURL(tt.rawURL)
			if got != tt.want {
				t.Fatalf("redactURL(%q) = %q, want %q", tt.rawURL, got, tt.want)
			}
		})
	}
}
