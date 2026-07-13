package main

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

func requestClientAddr(r *http.Request, trustedProxyCIDRs []netip.Prefix) string {
	remoteAddr := r.RemoteAddr
	remoteIP, ok := parseRemoteAddrIP(remoteAddr)
	if !ok || !isTrustedProxy(remoteIP, trustedProxyCIDRs) {
		return remoteAddr
	}

	forwardedAddrs, ok := parseForwardedForAddrs(r.Header.Values("X-Forwarded-For"))
	if !ok || len(forwardedAddrs) == 0 {
		return remoteAddr
	}

	for i := len(forwardedAddrs) - 1; i >= 0; i-- {
		if !isTrustedProxy(forwardedAddrs[i], trustedProxyCIDRs) {
			return forwardedAddrs[i].String()
		}
	}

	return forwardedAddrs[0].String()
}

func parseRemoteAddrIP(remoteAddr string) (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}

	addr, err := netip.ParseAddr(strings.Trim(host, "[]"))
	if err != nil {
		return netip.Addr{}, false
	}

	return addr.Unmap(), true
}

func parseForwardedForAddrs(values []string) ([]netip.Addr, bool) {
	addrs := []netip.Addr{}
	for _, value := range values {
		parts := strings.Split(value, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			addr, err := netip.ParseAddr(strings.Trim(part, "[]"))
			if err != nil {
				return nil, false
			}
			addrs = append(addrs, addr.Unmap())
		}
	}

	return addrs, true
}

func isTrustedProxy(addr netip.Addr, trustedProxyCIDRs []netip.Prefix) bool {
	for _, prefix := range trustedProxyCIDRs {
		if prefix.Contains(addr) {
			return true
		}
	}

	return false
}
