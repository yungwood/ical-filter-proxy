package main

import (
	"fmt"
	"log/slog"
	"os"
)

func main() {
	options, err := parseOptions(os.Args[1:], os.LookupEnv)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// print version and exit
	if options.printVersion {
		build := currentBuildInfo()
		fmt.Println("version:", build.Version)
		fmt.Println("revision:", build.Revision)
		fmt.Println("go version:", build.GoVersion)
		os.Exit(0)
	}

	// setup logging options
	loggingLevel := slog.LevelInfo // default loglevel
	if options.debugLogging {
		loggingLevel = slog.LevelDebug // debug logging enabled
	}
	opts := &slog.HandlerOptions{
		Level: loggingLevel,
	}

	// create json or text logger based on args
	var logger *slog.Logger
	if options.jsonLogging {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	slog.SetDefault(logger)

	// load configuration
	slog.Debug("reading config", "configFile", options.configFile)
	config, err := LoadConfig(options.configFile)
	if err != nil {
		slog.Error("Invalid configuration", "error", err)
		os.Exit(1) // fail if config is not valid
	}
	slog.Debug("loaded config")

	logConfigWarnings(config)

	runtimeConfig, err := config.Compile()
	if err != nil {
		slog.Error("Invalid configuration", "error", err)
		os.Exit(1)
	}

	// print a message and exit if validate arg was specified
	if options.validateConfig {
		slog.Info("configuration was validated successfully")
		os.Exit(0)
	}

	var metrics *prometheusMetrics
	if options.metricsEnabled {
		metrics = newPrometheusMetrics(options.calendarMetricsEnabled)
	}

	if options.metricsEnabled && options.managementAddress == "" {
		slog.Warn("Prometheus metrics endpoint enabled on public listener; set -management-address to expose management endpoints separately")
	}

	servers := buildHTTPServers(runtimeConfig, options.address, options.managementAddress, metrics)
	if err := runHTTPServers(servers...); err != nil {
		slog.Error("Error running web server", "error", err)
		os.Exit(1)
	}

}

func logConfigWarnings(config Config) {
	for _, calendar := range config.Calendars {
		if calendar.Public {
			slog.Warn("Calendar has no token set. Authentication will be disabled", "calendar", calendar.Name)
		}
		if len(calendar.Filters) == 0 {
			slog.Warn("Calendar has no filters and will be proxy-only", "calendar", calendar.Name)
		}
	}
}
