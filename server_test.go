package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPServerSetsTimeouts(t *testing.T) {
	server := newHTTPServer(":8080", http.NewServeMux())

	if server.Addr != ":8080" {
		t.Fatalf("Addr = %q, want :8080", server.Addr)
	}
	if server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("ReadHeaderTimeout = %s, want 5s", server.ReadHeaderTimeout)
	}
	if server.ReadTimeout != 10*time.Second {
		t.Fatalf("ReadTimeout = %s, want 10s", server.ReadTimeout)
	}
	if server.WriteTimeout != 30*time.Second {
		t.Fatalf("WriteTimeout = %s, want 30s", server.WriteTimeout)
	}
	if server.IdleTimeout != 60*time.Second {
		t.Fatalf("IdleTimeout = %s, want 60s", server.IdleTimeout)
	}
}

func TestBuildHTTPServersRegistersInternalRoutesOnPublicServerByDefault(t *testing.T) {
	servers := buildHTTPServers(testServerConfig(), ":8080", "", nil)

	if len(servers) != 1 {
		t.Fatalf("server count = %d, want 1", len(servers))
	}
	if servers[0].name != "public" {
		t.Fatalf("server name = %q, want public", servers[0].name)
	}
	if servers[0].server.Addr != ":8080" {
		t.Fatalf("server address = %q, want :8080", servers[0].server.Addr)
	}

	assertHandlerStatus(t, servers[0].server.Handler, "/liveness", http.StatusOK)
	assertHandlerStatus(t, servers[0].server.Handler, "/calendars/private/feed", http.StatusUnauthorized)
}

func TestBuildHTTPServersSeparatesManagementRoutes(t *testing.T) {
	servers := buildHTTPServers(testServerConfig(), ":8080", "127.0.0.1:9090", nil)

	if len(servers) != 2 {
		t.Fatalf("server count = %d, want 2", len(servers))
	}
	if servers[0].name != "public" {
		t.Fatalf("public server name = %q, want public", servers[0].name)
	}
	if servers[1].name != "management" {
		t.Fatalf("management server name = %q, want management", servers[1].name)
	}
	if servers[1].server.Addr != "127.0.0.1:9090" {
		t.Fatalf("management server address = %q, want 127.0.0.1:9090", servers[1].server.Addr)
	}

	assertHandlerStatus(t, servers[0].server.Handler, "/liveness", http.StatusNotFound)
	assertHandlerStatus(t, servers[0].server.Handler, "/calendars/private/feed", http.StatusUnauthorized)
	assertHandlerStatus(t, servers[1].server.Handler, "/liveness", http.StatusOK)
	assertHandlerStatus(t, servers[1].server.Handler, "/calendars/private/feed", http.StatusNotFound)
}

func testServerConfig() RuntimeConfig {
	return RuntimeConfig{
		Calendars: []Calendar{
			{
				Name:  "private",
				Token: "secret",
			},
		},
	}
}

func assertHandlerStatus(t *testing.T, handler http.Handler, target string, want int) {
	t.Helper()

	req := testRequest(t, target)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != want {
		t.Fatalf("%s status = %d, want %d", target, rr.Code, want)
	}
}
