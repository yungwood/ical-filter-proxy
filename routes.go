package main

import (
	"log/slog"
	"net/http"
)

func registerPublicRoutes(mux *http.ServeMux, config RuntimeConfig, metrics *prometheusMetrics) {
	for _, calendar := range config.Calendars {
		httpPath := "/calendars/" + calendar.Name + "/feed"
		slog.Debug("configuring endpoint", "calendar", calendar.Name, "http_path", httpPath)

		fetch := calendar.fetch
		if metrics != nil {
			fetch = metrics.instrumentFetch(calendar.Name, fetch)
		}

		mux.HandleFunc(httpPath, calendarFeedHandlerWithFetch(calendar, fetch))
	}
}

func registerInternalRoutes(mux *http.ServeMux, metrics *prometheusMetrics) {
	mux.HandleFunc("/liveness", func(_ http.ResponseWriter, _ *http.Request) {})
	mux.HandleFunc("/readiness", func(_ http.ResponseWriter, _ *http.Request) {})
	if metrics != nil {
		mux.Handle("/metrics", metrics.handler())
	}
}
