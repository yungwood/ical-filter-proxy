package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"

	ics "github.com/arran4/golang-ical"
)

// CalendarConfig is the YAML-backed calendar configuration. It is converted
// into Calendar before request handling so runtime code uses compiled filters
// and resolved secret values.
type CalendarConfig struct {
	Name        string         `yaml:"name"`
	PublishName string         `yaml:"publish_name"`
	Public      bool           `yaml:"public"`
	Token       string         `yaml:"token"`
	TokenFile   string         `yaml:"token_file"`
	UserAgent   string         `yaml:"user_agent"`
	FeedURL     string         `yaml:"feed_url"`
	FeedURLFile string         `yaml:"feed_url_file"`
	Filters     []FilterConfig `yaml:"filters"`
}

type RuntimeConfig struct {
	Calendars []Calendar
}

// Calendar is the runtime calendar representation used by handlers. It should
// contain values that are ready to execute, not file paths or raw rule strings
// that still need to be resolved.
type Calendar struct {
	Name        string
	PublishName string
	Public      bool
	Token       string
	UserAgent   string
	FeedURL     string
	Filters     []Filter
}

// Compile converts validated YAML config into runtime state used by the HTTP
// server. This is where raw match rules are compiled and regex syntax errors
// become startup validation failures.
func (config Config) Compile() (RuntimeConfig, error) {
	calendars := make([]Calendar, 0, len(config.Calendars))
	for _, calendarConfig := range config.Calendars {
		filters := make([]Filter, 0, len(calendarConfig.Filters))
		for filterIndex, filter := range calendarConfig.Filters {
			compiledFilter, err := filter.compile()
			if err != nil {
				return RuntimeConfig{}, fmt.Errorf("calendar %q filter %d: %w", calendarConfig.Name, filterIndex, err)
			}
			filters = append(filters, compiledFilter)
		}

		calendars = append(calendars, Calendar{
			Name:        calendarConfig.Name,
			PublishName: calendarConfig.PublishName,
			Public:      calendarConfig.Public,
			Token:       calendarConfig.Token,
			UserAgent:   calendarConfig.UserAgent,
			FeedURL:     calendarConfig.FeedURL,
			Filters:     filters,
		})
	}

	return RuntimeConfig{Calendars: calendars}, nil
}

// Downloads iCal feed from the URL and applies filtering rules
func (calendar Calendar) fetch(ctx context.Context) ([]byte, error) {

	// get the iCal feed
	slog.Debug("fetching upstream calendar", "calendar", calendar.Name, "url", redactURL(calendar.FeedURL))
	feedData, err := fetchUpstreamCalendar(ctx, calendar.FeedURL, calendar.UserAgent)
	if err != nil {
		return nil, err
	}

	// parse calendar
	cal, err := ics.ParseCalendar(strings.NewReader(string(feedData)))
	if err != nil {
		return nil, err
	}

	if calendar.PublishName != "" {
		cal.SetName(calendar.PublishName)
	}

	// process filters
	if len(calendar.Filters) > 0 {
		slog.Debug("processing filters", "calendar", calendar.Name)
		for _, event := range cal.Events() {
			if !calendar.ProcessEvent(event) {
				cal.RemoveEvent(event.Id())
			}
		}
		slog.Debug("filter processing completed", "calendar", calendar.Name)
	} else {
		slog.Debug("no filters to evaluate", "calendar", calendar.Name)
	}

	// serialize output
	var buf bytes.Buffer
	err = cal.SerializeTo(&buf)
	if err != nil {
		return nil, err
	}

	// return
	return buf.Bytes(), nil
}

// Evaluate the filters for a calendar against a given VEvent and
// perform any transformations directly to the VEvent (pointer)
// This function returns false if an event should be deleted
func (calendar Calendar) ProcessEvent(event *ics.VEvent) bool {

	// Get the Summary (the "title" of the event)
	// In case we cannot parse the event summary it should get dropped
	summary := event.GetProperty(ics.ComponentPropertySummary) // summary only for logging
	if summary == nil {
		return false
	}

	// Iterate through the Filter rules
	for id, filter := range calendar.Filters {

		// Does the filter match the event?
		if filter.matchesEvent(*event, calendar.Name) {
			slog.Debug("filter match found", "calendar", calendar.Name, "filter_index", id, "filter_description", filter.Description, "event_summary", summary.Value)

			// The event should get dropped if RemoveEvent is set
			if filter.RemoveEvent {
				slog.Debug("event removed; stopping rule processing", "calendar", calendar.Name, "action", "delete", "filter_index", id, "filter_description", filter.Description, "event_summary", summary.Value)
				return false
			}

			// Apply transformation rules to event
			filter.transformEvent(event)

			// Check if we should stop processing rules
			if filter.Stop {
				slog.Debug("filter stop set; stopping rule processing", "calendar", calendar.Name, "filter_index", id, "filter_description", filter.Description, "event_summary", summary.Value)
				return true
			}
		}
	}

	// Keep event by default if all Filter rules are processed
	slog.Debug("rule processing complete; keeping event", "calendar", calendar.Name, "filter_index", nil, "event_summary", summary.Value)
	return true

}
