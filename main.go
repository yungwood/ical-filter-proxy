package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

var version = "development"

func main() {

	// command-line args
	var (
		configFile     string
		debugLogging   bool
		jsonLogging    bool
		listenPort     int
		validateConfig bool
		printVersion   bool
	)
	flag.StringVar(&configFile, "config", "config.yaml", "config file")
	flag.BoolVar(&debugLogging, "debug", false, "enable debug logging")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
	flag.BoolVar(&jsonLogging, "json", false, "output logging in JSON format")
	flag.IntVar(&listenPort, "port", 8080, "listening port for api")
	flag.BoolVar(&validateConfig, "validate", false, "validate config and exit")
	flag.Parse()

	// print version and exit
	if printVersion {
		fmt.Println("version:", version)
		os.Exit(0)
	}

	// setup logging options
	loggingLevel := slog.LevelInfo // default loglevel
	if debugLogging {
		loggingLevel = slog.LevelDebug // debug logging enabled
	}
	opts := &slog.HandlerOptions{
		Level: loggingLevel,
	}

	// create json or text logger based on args
	var logger *slog.Logger
	if jsonLogging {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	slog.SetDefault(logger)

	// load configuration
	slog.Debug("reading config", "configFile", configFile)
	var config Config
	if !config.LoadConfig(configFile) {
		os.Exit(1) // fail if config is not valid
	}
	slog.Debug("loaded config")

	// print a message and exit if validate arg was specified
	if validateConfig {
		slog.Info("configuration was validated successfully")
		os.Exit(0)
	}

	// iterate through calendars in the config and setup a handler for each
	// todo: consider refactor to route requests dynamically?
	for _, calendarConfig := range config.Calendars {

		// configure HTTP endpoint
		httpPath := "/calendars/" + calendarConfig.Name + "/feed"
		slog.Debug("Configuring endpoint", "calendar", calendarConfig.Name, "http_path", httpPath)
		http.HandleFunc(httpPath, func(w http.ResponseWriter, r *http.Request) {

			slog.Debug("Received request for calendar", "http_path", httpPath, "calendar", calendarConfig.Name, "client_ip", r.RemoteAddr)

			// validate token
			token := r.URL.Query().Get("token")
			if token != calendarConfig.Token {
				slog.Warn("Unauthorized access attempt", "client_ip", r.RemoteAddr)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// fetch and filter upstream calendar
			feed, err := calendarConfig.fetch(r.Context())
			if err != nil {
				slog.Error("Error fetching and filtering feed", "error", err)
				http.Error(w, "Bad Gateway", http.StatusBadGateway)
				return
			}

			// return calendar
			w.Header().Set("Content-Type", "text/calendar")
			_, err = w.Write(feed)
			if err != nil {
				slog.Error("Error writing response", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			slog.Info("Calendar request processed", "http_path", httpPath, "calendar", calendarConfig.Name, "client_ip", r.RemoteAddr)
		})

	}

	// add a readiness and liveness check endpoint (return blank 200 OK response)
	http.HandleFunc("/liveness", func(_ http.ResponseWriter, _ *http.Request) {})
	http.HandleFunc("/readiness", func(_ http.ResponseWriter, _ *http.Request) {})

	// start the webserver
	slog.Info("Starting web server", "port", listenPort)
	if err := http.ListenAndServe(":"+strconv.Itoa(listenPort), nil); err != nil {
		slog.Error("Error starting web server", "error", err)
	}

}
