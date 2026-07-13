package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const gracefulShutdownTimeout = 10 * time.Second

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
	)
	flag.StringVar(&configFile, "config", "config.yaml", "config file")
	flag.BoolVar(&debugLogging, "debug", false, "enable debug logging")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
	flag.BoolVar(&jsonLogging, "json", false, "output logging in JSON format")
	flag.IntVar(&listenPort, "port", 8080, "listening port for api")
	flag.BoolVar(&validateConfig, "validate", false, "validate config and exit")
	flag.BoolVar(&metricsEnabled, "metrics", false, "enable prometheus metrics endpoint")
	flag.BoolVar(&calendarMetricsEnabled, "metrics-calendar-labels", false, "enable per-calendar prometheus metrics")
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

	var metrics *prometheusMetrics
	if metricsEnabled {
		metrics = newPrometheusMetrics(calendarMetricsEnabled)
	}

	mux := http.NewServeMux()
	registerPublicRoutes(mux, config, metrics)
	registerInternalRoutes(mux, metrics)

	handler := recoveryMiddleware(mux)
	if metrics != nil {
		handler = metrics.middleware(handler)
	}
	handler = requestLoggingMiddleware(handler)

	server := &http.Server{
		Addr:              ":" + strconv.Itoa(listenPort),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// start the webserver
	slog.Info("Starting web server", "port", listenPort)
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(shutdownSignal)

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Error running web server", "error", err)
		}
	case sig := <-shutdownSignal:
		slog.Info("Stopping web server", "signal", sig.String())

		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("Error stopping web server", "error", err)
			if closeErr := server.Close(); closeErr != nil && !errors.Is(closeErr, http.ErrServerClosed) {
				slog.Error("Error closing web server", "error", closeErr)
			}
		}

		if err := <-serverErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Error running web server", "error", err)
		}
	}

}
