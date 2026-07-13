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
	allowedMethods      = "GET, HEAD"
)

type calendarFetchFunc func(context.Context) ([]byte, error)

func calendarFeedHandlerWithFetch(calendar Calendar, fetch calendarFetchFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setCommonResponseHeaders(w)

		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", allowedMethods)
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		if !calendar.Public && !tokenMatches(r.URL.Query().Get("token"), calendar.Token) {
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
		if r.Method == http.MethodHead {
			return
		}

		_, err = w.Write(feed)
		if err != nil {
			slog.Error("Error writing response", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

func setCommonResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", cacheControlHeader)
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

func tokenMatches(token string, expectedToken string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) == 1
}
