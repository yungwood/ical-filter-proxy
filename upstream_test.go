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

	got, err := fetchUpstreamCalendar(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("fetchUpstreamCalendar returned error: %v", err)
	}
	if string(got) != body {
		t.Fatalf("body = %q, want %q", string(got), body)
	}
}

func TestFetchUpstreamCalendarRejectsNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "upstream unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, err := fetchUpstreamCalendar(context.Background(), server.URL)
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

	_, err := fetchUpstreamCalendar(context.Background(), server.URL)
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

	_, err := fetchUpstreamCalendar(ctx, server.URL)
	if err == nil {
		t.Fatal("fetchUpstreamCalendar returned nil error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}
