package main

import (
	"net/netip"
	"strings"
	"testing"
)

func TestParseOptionsDefaults(t *testing.T) {
	options, err := parseOptions(nil, emptyEnv)
	if err != nil {
		t.Fatalf("parseOptions() returned error: %v", err)
	}

	if options.configFile != "config.yaml" {
		t.Fatalf("configFile = %q, want config.yaml", options.configFile)
	}
	if options.address != ":8080" {
		t.Fatalf("address = %q, want :8080", options.address)
	}
	if options.debugLogging {
		t.Fatal("debugLogging = true, want false")
	}
	if options.printVersion {
		t.Fatal("printVersion = true, want false")
	}
}

func TestParseOptionsUsesEnvironmentDefaults(t *testing.T) {
	env := map[string]string{
		envConfig:                "/etc/ical-filter-proxy/config.yaml",
		envAddress:               "127.0.0.1:8081",
		envDebug:                 "true",
		envJSON:                  "true",
		envValidate:              "true",
		envMetrics:               "true",
		envMetricsCalendarLabels: "true",
		envManagementAddress:     "127.0.0.1:9090",
		envTrustedProxyCIDRs:     "127.0.0.1/32, 10.0.0.0/8",
	}

	options, err := parseOptions(nil, mapEnv(env))
	if err != nil {
		t.Fatalf("parseOptions() returned error: %v", err)
	}

	if options.configFile != "/etc/ical-filter-proxy/config.yaml" {
		t.Fatalf("configFile = %q, want env value", options.configFile)
	}
	if options.address != "127.0.0.1:8081" {
		t.Fatalf("address = %q, want env value", options.address)
	}
	if options.managementAddress != "127.0.0.1:9090" {
		t.Fatalf("managementAddress = %q, want env value", options.managementAddress)
	}
	if got, want := strings.Join(prefixStrings(options.trustedProxyCIDRs), ","), "127.0.0.1/32,10.0.0.0/8"; got != want {
		t.Fatalf("trustedProxyCIDRs = %q, want %q", got, want)
	}
	if !options.debugLogging || !options.jsonLogging || !options.validateConfig || !options.metricsEnabled || !options.calendarMetricsEnabled {
		t.Fatalf("boolean options = %+v, want all env booleans true", options)
	}
}

func TestParseOptionsCLIOverridesEnvironment(t *testing.T) {
	env := map[string]string{
		envConfig:            "/etc/config.yaml",
		envAddress:           ":8081",
		envDebug:             "true",
		envMetrics:           "true",
		envManagementAddress: "127.0.0.1:9090",
		envTrustedProxyCIDRs: "10.0.0.0/8",
	}

	options, err := parseOptions([]string{
		"-config", "/tmp/config.yaml",
		"-address", ":8082",
		"-debug=false",
		"-metrics=false",
		"-management-address", "127.0.0.1:9091",
		"-trusted-proxy-cidr", "127.0.0.1/32",
		"-trusted-proxy-cidr", "192.0.2.0/24",
	}, mapEnv(env))
	if err != nil {
		t.Fatalf("parseOptions() returned error: %v", err)
	}

	if options.configFile != "/tmp/config.yaml" {
		t.Fatalf("configFile = %q, want CLI value", options.configFile)
	}
	if options.address != ":8082" {
		t.Fatalf("address = %q, want CLI value", options.address)
	}
	if options.debugLogging {
		t.Fatal("debugLogging = true, want CLI false")
	}
	if options.metricsEnabled {
		t.Fatal("metricsEnabled = true, want CLI false")
	}
	if options.managementAddress != "127.0.0.1:9091" {
		t.Fatalf("managementAddress = %q, want CLI value", options.managementAddress)
	}
	if got, want := strings.Join(prefixStrings(options.trustedProxyCIDRs), ","), "127.0.0.1/32,192.0.2.0/24"; got != want {
		t.Fatalf("trustedProxyCIDRs = %q, want CLI values %q", got, want)
	}
}

func TestParseOptionsRejectsInvalidBoolEnvironment(t *testing.T) {
	_, err := parseOptions(nil, mapEnv(map[string]string{
		envDebug: "sometimes",
	}))
	if err == nil {
		t.Fatal("parseOptions() returned nil error")
	}
	if !strings.Contains(err.Error(), envDebug) {
		t.Fatalf("parseOptions() error = %q, want env var name", err)
	}
}

func TestParseOptionsRejectsInvalidTrustedProxyCIDREnvironment(t *testing.T) {
	_, err := parseOptions(nil, mapEnv(map[string]string{
		envTrustedProxyCIDRs: "not-a-cidr",
	}))
	if err == nil {
		t.Fatal("parseOptions() returned nil error")
	}
	if !strings.Contains(err.Error(), "not-a-cidr") {
		t.Fatalf("parseOptions() error = %q, want invalid CIDR value", err)
	}
}

func TestParseOptionsRejectsInvalidTrustedProxyCIDRFlag(t *testing.T) {
	_, err := parseOptions([]string{"-trusted-proxy-cidr", "not-a-cidr"}, emptyEnv)
	if err == nil {
		t.Fatal("parseOptions() returned nil error")
	}
	if !strings.Contains(err.Error(), "not-a-cidr") {
		t.Fatalf("parseOptions() error = %q, want invalid CIDR value", err)
	}
}

func TestParseOptionsVersionIsCLIOnly(t *testing.T) {
	options, err := parseOptions(nil, mapEnv(map[string]string{
		"ICAL_FILTER_PROXY_VERSION": "true",
	}))
	if err != nil {
		t.Fatalf("parseOptions() returned error: %v", err)
	}
	if options.printVersion {
		t.Fatal("printVersion = true from environment, want false")
	}

	options, err = parseOptions([]string{"-version"}, emptyEnv)
	if err != nil {
		t.Fatalf("parseOptions() returned error: %v", err)
	}
	if !options.printVersion {
		t.Fatal("printVersion = false, want CLI true")
	}
}

func emptyEnv(string) (string, bool) {
	return "", false
}

func mapEnv(values map[string]string) envLookupFunc {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}

func prefixStrings(prefixes []netip.Prefix) []string {
	values := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		values = append(values, prefix.String())
	}

	return values
}
