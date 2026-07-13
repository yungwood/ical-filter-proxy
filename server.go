package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const gracefulShutdownTimeout = 10 * time.Second

type managedHTTPServer struct {
	name   string
	server *http.Server
}

type httpServerError struct {
	name string
	err  error
}

func newHTTPServer(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func buildHTTPHandler(listener string, handler http.Handler, metrics *prometheusMetrics) http.Handler {
	handler = recoveryMiddleware(handler)
	if metrics != nil {
		handler = metrics.middleware(listener, handler)
	}

	return requestLoggingMiddleware(handler)
}

func buildHTTPServers(config RuntimeConfig, address string, managementAddress string, metrics *prometheusMetrics) []managedHTTPServer {
	publicMux := http.NewServeMux()
	registerPublicRoutes(publicMux, config, metrics)

	servers := []managedHTTPServer{
		{
			name:   "public",
			server: newHTTPServer(address, buildHTTPHandler("public", publicMux, metrics)),
		},
	}

	if managementAddress == "" {
		// Without a separate management listener, expose health and metrics on
		// the public listener for backwards-compatible single-port operation.
		registerInternalRoutes(publicMux, metrics)
		return servers
	}

	managementMux := http.NewServeMux()
	registerInternalRoutes(managementMux, metrics)
	return append(servers, managedHTTPServer{
		name:   "management",
		server: newHTTPServer(managementAddress, buildHTTPHandler("management", managementMux, metrics)),
	})
}

func runHTTPServers(servers ...managedHTTPServer) error {
	serverErr := make(chan httpServerError, len(servers))
	for _, managedServer := range servers {
		slog.Info("Starting web server", "name", managedServer.name, "address", managedServer.server.Addr)
		go func(managedServer managedHTTPServer) {
			serverErr <- httpServerError{
				name: managedServer.name,
				err:  managedServer.server.ListenAndServe(),
			}
		}(managedServer)
	}

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(shutdownSignal)

	select {
	case err := <-serverErr:
		if err.err != nil && !errors.Is(err.err, http.ErrServerClosed) {
			return fmt.Errorf("%s server: %w", err.name, err.err)
		}
	case sig := <-shutdownSignal:
		slog.Info("Stopping web servers", "signal", sig.String())

		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()

		for _, managedServer := range servers {
			if err := managedServer.server.Shutdown(shutdownCtx); err != nil {
				slog.Error("Error stopping web server", "name", managedServer.name, "error", err)
				if closeErr := managedServer.server.Close(); closeErr != nil && !errors.Is(closeErr, http.ErrServerClosed) {
					slog.Error("Error closing web server", "name", managedServer.name, "error", closeErr)
				}
			}
		}

		for range servers {
			err := <-serverErr
			if err.err != nil && !errors.Is(err.err, http.ErrServerClosed) {
				return fmt.Errorf("%s server: %w", err.name, err.err)
			}
		}
	}

	return nil
}
