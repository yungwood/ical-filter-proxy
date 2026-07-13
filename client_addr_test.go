package main

import (
	"net/http"
	"net/netip"
	"testing"
)

func TestRequestClientAddrUsesRemoteAddrByDefault(t *testing.T) {
	req := clientAddrRequest("198.51.100.10:12345", "203.0.113.9")

	if got, want := requestClientAddr(req, nil), "198.51.100.10:12345"; got != want {
		t.Fatalf("requestClientAddr() = %q, want %q", got, want)
	}
}

func TestRequestClientAddrIgnoresForwardedForFromUntrustedRemote(t *testing.T) {
	req := clientAddrRequest("198.51.100.10:12345", "203.0.113.9")

	if got, want := requestClientAddr(req, mustPrefixes(t, "10.0.0.0/8")), "198.51.100.10:12345"; got != want {
		t.Fatalf("requestClientAddr() = %q, want %q", got, want)
	}
}

func TestRequestClientAddrUsesForwardedForFromTrustedRemote(t *testing.T) {
	req := clientAddrRequest("10.0.0.10:12345", "203.0.113.9")

	if got, want := requestClientAddr(req, mustPrefixes(t, "10.0.0.0/8")), "203.0.113.9"; got != want {
		t.Fatalf("requestClientAddr() = %q, want %q", got, want)
	}
}

func TestRequestClientAddrSelectsNearestUntrustedForwardedFor(t *testing.T) {
	req := clientAddrRequest("10.0.0.10:12345", "198.51.100.1, 192.0.2.10")

	trustedProxyCIDRs := mustPrefixes(t, "10.0.0.0/8", "192.0.2.0/24")
	if got, want := requestClientAddr(req, trustedProxyCIDRs), "198.51.100.1"; got != want {
		t.Fatalf("requestClientAddr() = %q, want %q", got, want)
	}
}

func TestRequestClientAddrFallsBackForInvalidForwardedFor(t *testing.T) {
	req := clientAddrRequest("10.0.0.10:12345", "203.0.113.9, not-an-ip")

	if got, want := requestClientAddr(req, mustPrefixes(t, "10.0.0.0/8")), "10.0.0.10:12345"; got != want {
		t.Fatalf("requestClientAddr() = %q, want %q", got, want)
	}
}

func clientAddrRequest(remoteAddr string, forwardedFor string) *http.Request {
	req := &http.Request{
		RemoteAddr: remoteAddr,
		Header:     http.Header{},
	}
	if forwardedFor != "" {
		req.Header.Set("X-Forwarded-For", forwardedFor)
	}

	return req
}

func mustPrefixes(t *testing.T, cidrs ...string) []netip.Prefix {
	t.Helper()

	prefixes := make([]netip.Prefix, 0, len(cidrs))
	for _, cidr := range cidrs {
		prefix, err := netip.ParsePrefix(cidr)
		if err != nil {
			t.Fatalf("ParsePrefix(%q) returned error: %v", cidr, err)
		}
		prefixes = append(prefixes, prefix)
	}

	return prefixes
}
