package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is used to parse config.yaml.
type Config struct {
	Calendars []CalendarConfig `yaml:"calendars"`
}

// LoadConfig loads YAML config, resolves secret files, and validates
// calendar-level settings. Rule compilation and regex validation happen when
// building RuntimeConfig.
func LoadConfig(file string) (Config, error) {
	return loadConfig(file, os.LookupEnv)
}

func loadConfig(file string, lookupEnv envLookupFunc) (Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Config{}, fmt.Errorf("open config file %q: %w; use -config to specify a different file", file, err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, fmt.Errorf("parse config file %q: %w", file, err)
	}

	// ensure calendars exist
	if len(config.Calendars) == 0 {
		return Config{}, errors.New("no calendars configured; configuration must define at least one calendar")
	}

	// validate calendar configs and load secrets
	for i := range config.Calendars {

		// grab pointer so we can mutate values when loading from file
		calendarConfig := &config.Calendars[i]

		if calendarConfig.Public && (calendarConfig.Token != "" || calendarConfig.TokenFile != "") {
			return Config{}, fmt.Errorf("calendar %q: public calendar cannot define token or token_file", calendarConfig.Name)
		}

		// check if url should be loaded from file
		if calendarConfig.FeedURLFile != "" {
			calendarConfig.FeedURL, err = readSecretFile(calendarConfig.FeedURLFile)
			if err != nil {
				return Config{}, fmt.Errorf("calendar %q: read feed_url_file %q: %w", calendarConfig.Name, calendarConfig.FeedURLFile, err)
			}
		} else {
			calendarConfig.FeedURL, err = resolveConfigEnvReference(calendarConfig.FeedURL, lookupEnv)
			if err != nil {
				return Config{}, fmt.Errorf("calendar %q: feed_url: %w", calendarConfig.Name, err)
			}
		}

		// check if url seems valid
		if !strings.HasPrefix(calendarConfig.FeedURL, "http://") && !strings.HasPrefix(calendarConfig.FeedURL, "https://") {
			return Config{}, fmt.Errorf("calendar %q: feed_url must begin with http:// or https://", calendarConfig.Name)
		}

		// check if token should be loaded from file
		if calendarConfig.TokenFile != "" {
			calendarConfig.Token, err = readSecretFile(calendarConfig.TokenFile)
			if err != nil {
				return Config{}, fmt.Errorf("calendar %q: read token_file %q: %w", calendarConfig.Name, calendarConfig.TokenFile, err)
			}
		} else {
			calendarConfig.Token, err = resolveConfigEnvReference(calendarConfig.Token, lookupEnv)
			if err != nil {
				return Config{}, fmt.Errorf("calendar %q: token: %w", calendarConfig.Name, err)
			}
		}

		if !calendarConfig.Public && calendarConfig.Token == "" {
			return Config{}, fmt.Errorf("calendar %q: private calendar must define token or token_file", calendarConfig.Name)
		}
		calendarConfig.UserAgent = strings.TrimSpace(calendarConfig.UserAgent)

	}

	return config, nil
}

var exactEnvReferencePattern = regexp.MustCompile(`^\$\{([A-Za-z_][A-Za-z0-9_]*)\}$`)

func resolveConfigEnvReference(value string, lookupEnv envLookupFunc) (string, error) {
	matches := exactEnvReferencePattern.FindStringSubmatch(value)
	if matches == nil {
		return value, nil
	}

	envName := matches[1]
	envValue, ok := lookupEnv(envName)
	if !ok || envValue == "" {
		return "", fmt.Errorf("environment variable %s is not set or is empty", envName)
	}

	return envValue, nil
}

func readSecretFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
