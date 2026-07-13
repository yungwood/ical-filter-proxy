package main

import (
	"bytes"
	"context"
	"log/slog"
	"strings"

	ics "github.com/arran4/golang-ical"
)

// All structs defined in this file are used to unmarshall yaml configuration and
// provide helper functions that are used to fetch and filter events

// CalendarConfig definition
type CalendarConfig struct {
	Name        string   `yaml:"name"`
	PublishName string   `yaml:"publish_name"`
	Public      bool     `yaml:"public"`
	Token       string   `yaml:"token"`
	TokenFile   string   `yaml:"token_file"`
	FeedURL     string   `yaml:"feed_url"`
	FeedURLFile string   `yaml:"feed_url_file"`
	Filters     []Filter `yaml:"filters"`
}

// Downloads iCal feed from the URL and applies filtering rules
func (calendarConfig CalendarConfig) fetch(ctx context.Context) ([]byte, error) {

	// get the iCal feed
	slog.Debug("Fetching iCal feed", "url", calendarConfig.FeedURL)
	feedData, err := fetchUpstreamCalendar(ctx, calendarConfig.FeedURL)
	if err != nil {
		return nil, err
	}

	// parse calendar
	cal, err := ics.ParseCalendar(strings.NewReader(string(feedData)))
	if err != nil {
		return nil, err
	}

	if calendarConfig.PublishName != "" {
		cal.SetName(calendarConfig.PublishName)
	}

	// process filters
	if len(calendarConfig.Filters) > 0 {
		slog.Debug("Processing filters", "calendar", calendarConfig.Name)
		for _, event := range cal.Events() {
			if !calendarConfig.ProcessEvent(event) {
				cal.RemoveEvent(event.Id())
			}
		}
		slog.Debug("Filter processing completed", "calendar", calendarConfig.Name)
	} else {
		slog.Debug("No filters to evaluate", "calendar", calendarConfig.Name)
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
func (calendarConfig CalendarConfig) ProcessEvent(event *ics.VEvent) bool {

	// Get the Summary (the "title" of the event)
	// In case we cannot parse the event summary it should get dropped
	summary := event.GetProperty(ics.ComponentPropertySummary) // summary only for logging
	if summary == nil {
		return false
	}

	// Iterate through the Filter rules
	for id, filter := range calendarConfig.Filters {

		// Does the filter match the event?
		if filter.matchesEvent(*event) {
			slog.Debug("Filter match found", "rule_id", id, "filter_description", filter.Description, "event_summary", summary.Value)

			// The event should get dropped if RemoveEvent is set
			if filter.RemoveEvent {
				slog.Debug("Event to be removed, no more rules will be processed", "action", "DELETE", "rule_id", id, "filter_description", filter.Description, "event_summary", summary.Value)
				return false
			}

			// Apply transformation rules to event
			filter.transformEvent(event)

			// Check if we should stop processing rules
			if filter.Stop {
				slog.Debug("Stop option is set, no more rules will be processed", "rule_id", id, "filter_description", filter.Description, "event_summary", summary.Value)
				return true
			}
		}
	}

	// Keep event by default if all Filter rules are processed
	slog.Debug("Rule processing complete, event will be kept", "rule_id", nil, "event_summary", summary.Value)
	return true

}
