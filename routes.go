package main

import (
	"log/slog"
	"net/http"
)

func registerPublicRoutes(mux *http.ServeMux, config Config) {
	for _, calendarConfig := range config.Calendars {
		httpPath := "/calendars/" + calendarConfig.Name + "/feed"
		slog.Debug("Configuring endpoint", "calendar", calendarConfig.Name, "http_path", httpPath)
		mux.HandleFunc(httpPath, calendarFeedHandler(calendarConfig))
	}
}

func registerInternalRoutes(mux *http.ServeMux, metrics *prometheusMetrics) {
	mux.HandleFunc("/liveness", func(_ http.ResponseWriter, _ *http.Request) {})
	mux.HandleFunc("/readiness", func(_ http.ResponseWriter, _ *http.Request) {})
	if metrics != nil {
		mux.Handle("/metrics", metrics.handler())
	}
}
