package main

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"net/http"
)

const (
	calendarContentType = "text/calendar; charset=utf-8"
	cacheControlHeader  = "no-store"
)

type calendarFetchFunc func(context.Context) ([]byte, error)

func calendarFeedHandler(httpPath string, calendarConfig CalendarConfig) http.HandlerFunc {
	return calendarFeedHandlerWithFetch(httpPath, calendarConfig, calendarConfig.fetch)
}

func calendarFeedHandlerWithFetch(httpPath string, calendarConfig CalendarConfig, fetch calendarFetchFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setCommonResponseHeaders(w)

		slog.Debug("Received request for calendar", "http_path", httpPath, "calendar", calendarConfig.Name, "client_ip", r.RemoteAddr)

		if !calendarConfig.Public && !tokenMatches(r.URL.Query().Get("token"), calendarConfig.Token) {
			slog.Warn("Unauthorized access attempt", "client_ip", r.RemoteAddr)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// fetch and filter upstream calendar
		feed, err := fetch(r.Context())
		if err != nil {
			slog.Error("Error fetching and filtering feed", "error", err)
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
			return
		}

		// return calendar
		w.Header().Set("Content-Type", calendarContentType)
		_, err = w.Write(feed)
		if err != nil {
			slog.Error("Error writing response", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		slog.Info("Calendar request processed", "http_path", httpPath, "calendar", calendarConfig.Name, "client_ip", r.RemoteAddr)
	}
}

func setCommonResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", cacheControlHeader)
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

func tokenMatches(token string, expectedToken string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) == 1
}
