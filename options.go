package main

import (
	"flag"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
)

const (
	envConfig                = "ICAL_FILTER_PROXY_CONFIG"
	envAddress               = "ICAL_FILTER_PROXY_ADDRESS"
	envDebug                 = "ICAL_FILTER_PROXY_DEBUG"
	envJSON                  = "ICAL_FILTER_PROXY_JSON"
	envValidate              = "ICAL_FILTER_PROXY_VALIDATE"
	envMetrics               = "ICAL_FILTER_PROXY_METRICS"
	envMetricsCalendarLabels = "ICAL_FILTER_PROXY_METRICS_CALENDAR_LABELS"
	envManagementAddress     = "ICAL_FILTER_PROXY_MANAGEMENT_ADDRESS"
	envTrustedProxyCIDRs     = "ICAL_FILTER_PROXY_TRUSTED_PROXY_CIDRS"
)

type appOptions struct {
	configFile             string
	debugLogging           bool
	jsonLogging            bool
	address                string
	validateConfig         bool
	printVersion           bool
	metricsEnabled         bool
	calendarMetricsEnabled bool
	managementAddress      string
	trustedProxyCIDRs      []netip.Prefix
}

type envLookupFunc func(string) (string, bool)

func parseOptions(args []string, lookupEnv envLookupFunc) (appOptions, error) {
	options, err := defaultOptionsFromEnv(lookupEnv)
	if err != nil {
		return appOptions{}, err
	}

	flags := flag.NewFlagSet("ical-filter-proxy", flag.ContinueOnError)
	flags.StringVar(&options.configFile, "config", options.configFile, "config file")
	flags.BoolVar(&options.debugLogging, "debug", options.debugLogging, "enable debug logging")
	flags.BoolVar(&options.printVersion, "version", false, "print version and exit")
	flags.BoolVar(&options.jsonLogging, "json", options.jsonLogging, "output logging in JSON format")
	flags.StringVar(&options.address, "address", options.address, "address for calendar API listener")
	flags.BoolVar(&options.validateConfig, "validate", options.validateConfig, "validate config and exit")
	flags.BoolVar(&options.metricsEnabled, "metrics", options.metricsEnabled, "enable prometheus metrics endpoint")
	flags.BoolVar(&options.calendarMetricsEnabled, "metrics-calendar-labels", options.calendarMetricsEnabled, "enable per-calendar prometheus metrics")
	flags.StringVar(&options.managementAddress, "management-address", options.managementAddress, "optional address for liveness, readiness, and metrics endpoints")
	trustedProxyCIDRs := stringListFlag{}
	flags.Var(&trustedProxyCIDRs, "trusted-proxy-cidr", "trusted reverse proxy CIDR for forwarded client address handling; may be repeated")

	if err := flags.Parse(args); err != nil {
		return appOptions{}, err
	}

	if trustedProxyCIDRs.set {
		options.trustedProxyCIDRs, err = parseTrustedProxyCIDRs(trustedProxyCIDRs.values)
		if err != nil {
			return appOptions{}, err
		}
	}

	return options, nil
}

func defaultOptionsFromEnv(lookupEnv envLookupFunc) (appOptions, error) {
	options := appOptions{
		configFile: "config.yaml",
		address:    ":8080",
	}

	options.configFile = envString(lookupEnv, envConfig, options.configFile)
	options.address = envString(lookupEnv, envAddress, options.address)
	options.managementAddress = envString(lookupEnv, envManagementAddress, options.managementAddress)

	var err error
	if options.trustedProxyCIDRs, err = parseTrustedProxyCIDRs(envStringList(lookupEnv, envTrustedProxyCIDRs)); err != nil {
		return appOptions{}, err
	}
	if options.debugLogging, err = envBool(lookupEnv, envDebug, options.debugLogging); err != nil {
		return appOptions{}, err
	}
	if options.jsonLogging, err = envBool(lookupEnv, envJSON, options.jsonLogging); err != nil {
		return appOptions{}, err
	}
	if options.validateConfig, err = envBool(lookupEnv, envValidate, options.validateConfig); err != nil {
		return appOptions{}, err
	}
	if options.metricsEnabled, err = envBool(lookupEnv, envMetrics, options.metricsEnabled); err != nil {
		return appOptions{}, err
	}
	if options.calendarMetricsEnabled, err = envBool(lookupEnv, envMetricsCalendarLabels, options.calendarMetricsEnabled); err != nil {
		return appOptions{}, err
	}

	return options, nil
}

type stringListFlag struct {
	values []string
	set    bool
}

func (f *stringListFlag) String() string {
	return strings.Join(f.values, ",")
}

func (f *stringListFlag) Set(value string) error {
	f.set = true
	f.values = append(f.values, value)
	return nil
}

func envString(lookupEnv envLookupFunc, name string, fallback string) string {
	value, ok := lookupEnv(name)
	if !ok {
		return fallback
	}

	return value
}

func envStringList(lookupEnv envLookupFunc, name string) []string {
	value, ok := lookupEnv(name)
	if !ok || strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}

	return values
}

func envBool(lookupEnv envLookupFunc, name string, fallback bool) (bool, error) {
	value, ok := lookupEnv(name)
	if !ok {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", name, err)
	}

	return parsed, nil
}

func parseTrustedProxyCIDRs(cidrs []string) ([]netip.Prefix, error) {
	prefixes := make([]netip.Prefix, 0, len(cidrs))
	for _, cidr := range cidrs {
		prefix, err := netip.ParsePrefix(cidr)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR %q: %w", cidr, err)
		}
		prefixes = append(prefixes, prefix)
	}

	return prefixes, nil
}
