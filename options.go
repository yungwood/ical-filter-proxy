package main

import (
	"flag"
	"fmt"
	"strconv"
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

	if err := flags.Parse(args); err != nil {
		return appOptions{}, err
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

func envString(lookupEnv envLookupFunc, name string, fallback string) string {
	value, ok := lookupEnv(name)
	if !ok {
		return fallback
	}

	return value
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
