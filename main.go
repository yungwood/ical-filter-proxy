package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func main() {

	// command-line args
	var (
		configFile             string
		debugLogging           bool
		jsonLogging            bool
		listenPort             int
		validateConfig         bool
		printVersion           bool
		metricsEnabled         bool
		calendarMetricsEnabled bool
		managementAddress      string
	)
	flag.StringVar(&configFile, "config", "config.yaml", "config file")
	flag.BoolVar(&debugLogging, "debug", false, "enable debug logging")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
	flag.BoolVar(&jsonLogging, "json", false, "output logging in JSON format")
	flag.IntVar(&listenPort, "port", 8080, "listening port for api")
	flag.BoolVar(&validateConfig, "validate", false, "validate config and exit")
	flag.BoolVar(&metricsEnabled, "metrics", false, "enable prometheus metrics endpoint")
	flag.BoolVar(&calendarMetricsEnabled, "metrics-calendar-labels", false, "enable per-calendar prometheus metrics")
	flag.StringVar(&managementAddress, "management-address", "", "optional address for liveness, readiness, and metrics endpoints")
	flag.Parse()

	// print version and exit
	if printVersion {
		build := currentBuildInfo()
		fmt.Println("version:", build.Version)
		fmt.Println("revision:", build.Revision)
		fmt.Println("go version:", build.GoVersion)
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

	runtimeConfig := config.RuntimeConfig()

	var metrics *prometheusMetrics
	if metricsEnabled {
		metrics = newPrometheusMetrics(calendarMetricsEnabled)
	}

	if metricsEnabled && managementAddress == "" {
		slog.Warn("Prometheus metrics endpoint enabled on public listener; set -management-address to expose management endpoints separately")
	}

	servers := buildHTTPServers(runtimeConfig, listenPort, managementAddress, metrics)
	if err := runHTTPServers(servers...); err != nil {
		slog.Error("Error running web server", "error", err)
		os.Exit(1)
	}

}
