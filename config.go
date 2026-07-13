package main

import (
	"log/slog"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is used to parse config.yaml.
type Config struct {
	Calendars []CalendarConfig `yaml:"calendars"`
}

// LoadConfig loads the configuration file and does basic validation.
func (config *Config) LoadConfig(file string) bool {
	data, err := os.ReadFile(file)
	if err != nil {
		slog.Error("Unable to open config file! You can use -config to specify a different file", "file", file)
		return false
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		slog.Error("Error while unmarshalling yaml! Check config file is valid", "file", file)
		return false
	}

	// ensure calendars exist
	if len(config.Calendars) == 0 {
		slog.Error("No calendars found! Configuration should define at least one calendar")
		return false
	}

	// validate calendar configs and load secrets
	for i := range config.Calendars {

		// grab pointer so we can mutate values when loading from file
		calendarConfig := &config.Calendars[i]

		if calendarConfig.Public && (calendarConfig.Token != "" || calendarConfig.TokenFile != "") {
			slog.Error("Public calendar cannot define token or token_file", "calendar", calendarConfig.Name)
			return false
		}

		// check if url should be loaded from file
		if calendarConfig.FeedURLFile != "" {
			calendarConfig.FeedURL, err = readSecretFile(calendarConfig.FeedURLFile)
			if err != nil {
				slog.Error("Unable to read feed_url_file", "calendar", calendarConfig.Name, "feed_url_file", calendarConfig.FeedURLFile)
				return false
			}
		}

		// check if url seems valid
		if !strings.HasPrefix(calendarConfig.FeedURL, "http://") && !strings.HasPrefix(calendarConfig.FeedURL, "https://") {
			slog.Debug("Calendar URL must begin with http:// or https://", "calendar", calendarConfig.Name, "feed_url", len(calendarConfig.Filters))
			return false
		}

		// check if token should be loaded from file
		if calendarConfig.TokenFile != "" {
			calendarConfig.Token, err = readSecretFile(calendarConfig.TokenFile)
			if err != nil {
				slog.Error("Unable to read token_file", "calendar", calendarConfig.Name, "token_file", calendarConfig.TokenFile)
				return false
			}
		}

		if calendarConfig.Public {
			slog.Warn("Calendar has no token set. Authentication will be disabled", "calendar", calendarConfig.Name)
		} else if calendarConfig.Token == "" {
			slog.Error("Private calendar must define token or token_file", "calendar", calendarConfig.Name)
			return false
		}

		// Print a warning if the calendar has no filters
		if len(calendarConfig.Filters) == 0 {
			slog.Warn("Calendar has no filters and will be proxy-only", "calendar", calendarConfig.Name)
			continue
		}

		if !calendarConfig.validateFilters() {
			return false
		}
	}

	return true // config is parsed successfully
}

func (calendarConfig CalendarConfig) validateFilters() bool {
	for filterIndex, filter := range calendarConfig.Filters {
		matchRules := []struct {
			name string
			rule StringMatchRule
		}{
			{name: "summary", rule: filter.Match.Summary},
			{name: "description", rule: filter.Match.Description},
			{name: "location", rule: filter.Match.Location},
			{name: "url", rule: filter.Match.URL},
		}

		for _, matchRule := range matchRules {
			if err := matchRule.rule.validate(); err != nil {
				slog.Error("Invalid string match rule", "calendar", calendarConfig.Name, "filter_index", filterIndex, "property", matchRule.name, "error", err)
				return false
			}
		}
	}

	return true
}

func readSecretFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
