package main

import (
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
	}

	options, err := parseOptions([]string{
		"-config", "/tmp/config.yaml",
		"-address", ":8082",
		"-debug=false",
		"-metrics=false",
		"-management-address", "127.0.0.1:9091",
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
