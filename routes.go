package main

import (
	"log/slog"
	"net/http"
)

func registerPublicRoutes(mux *http.ServeMux, config Config, metrics *prometheusMetrics) {
	for _, calendarConfig := range config.Calendars {
		httpPath := "/calendars/" + calendarConfig.Name + "/feed"
		slog.Debug("Configuring endpoint", "calendar", calendarConfig.Name, "http_path", httpPath)

		fetch := calendarConfig.fetch
		if metrics != nil {
			fetch = metrics.instrumentFetch(calendarConfig.Name, fetch)
		}

		mux.HandleFunc(httpPath, calendarFeedHandlerWithFetch(calendarConfig, fetch))
	}
}

func registerInternalRoutes(mux *http.ServeMux, metrics *prometheusMetrics) {
	mux.HandleFunc("/liveness", func(_ http.ResponseWriter, _ *http.Request) {})
	mux.HandleFunc("/readiness", func(_ http.ResponseWriter, _ *http.Request) {})
	if metrics != nil {
		mux.Handle("/metrics", metrics.handler())
	}
}
