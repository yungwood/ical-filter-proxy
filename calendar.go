package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	ics "github.com/arran4/golang-ical"
)

// All structs defined in this file are used to unmarshall yaml configuration and
// provide helper functions that are used to fetch and filter events

// CalendarConfig definition
type CalendarConfig struct {
	Name        string   `yaml:"name"`
	PublishName string   `yaml:"publish_name"`
	Token       string   `yaml:"token"`
	FeedURL     string   `yaml:"feed_url"`
	Filters     []Filter `yaml:"filters"`
}

// Downloads iCal feed from the URL and applies filtering rules
func (calendarConfig CalendarConfig) fetch() ([]byte, error) {

	// get the iCal feed
	slog.Debug("Fetching iCal feed", "url", calendarConfig.FeedURL)
	resp, err := http.Get(calendarConfig.FeedURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	feedData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// parse calendar
	cal, err := ics.ParseCalendar(strings.NewReader(string(feedData)))
	if err != nil {
		return nil, err
	}

	if (calendarConfig.PublishName != "") {
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

// Filter definition
type Filter struct {
	Description string              `yaml:"description"`
	RemoveEvent bool                `yaml:"remove"`
	Stop        bool                `yaml:"stop"`
	Match       EventMatchRules     `yaml:"match"`
	Transform   EventTransformRules `yaml:"transform"`
}

// Returns true if a VEvent matches the Filter conditions
func (filter Filter) matchesEvent(event ics.VEvent) bool {

	// If an event property is not defined golang-ical returns a nil pointer

	// Get event Summary - only used for debug logging
	eventSummary := event.GetProperty(ics.ComponentPropertySummary)
	if eventSummary == nil {
		slog.Warn("Unable to process event summary. Event will be dropped")
		return false // never match if VEvent has no summary
	}

	// Check Summary filters against VEvent
	if filter.Match.Summary.hasConditions() {
		if !filter.Match.Summary.matchesString(eventSummary.Value) {
			return false
		}
	}

	// Check Description filters against VEvent
	if filter.Match.Description.hasConditions() {
		eventDescription := event.GetProperty(ics.ComponentPropertyDescription)
		var eventDescriptionValue string
		if eventDescription == nil {
			eventDescriptionValue = ""
		} else {
			eventDescriptionValue = eventDescription.Value
		}

		if !filter.Match.Description.matchesString(eventDescriptionValue) {
			slog.Debug("Event Description does not match filter conditions", "event_summary", eventSummary.Value, "filter", filter.Description)
			return false // event doesn't match
		}
	}

	// Check Location filters against VEvent
	if filter.Match.Location.hasConditions() {
		eventLocation := event.GetProperty(ics.ComponentPropertyLocation)
		var eventLocationValue string
		if eventLocation == nil {
			eventLocationValue = ""
		} else {
			eventLocationValue = eventLocation.Value
		}
		if !filter.Match.Location.matchesString(eventLocationValue) {
			slog.Debug("Event Location does not match filter conditions", "event_summary", eventSummary.Value, "filter", filter.Description)
			return false // event doesn't match

		}
	}

	// Check Url filters against VEvent
	if filter.Match.Url.hasConditions() {
		eventUrl := event.GetProperty(ics.ComponentPropertyUrl)
		var eventUrlValue string
		if eventUrl == nil {
			eventUrlValue = ""
		} else {
			eventUrlValue = eventUrl.Value
		}
		if !filter.Match.Url.matchesString(eventUrlValue) {
			slog.Debug("Event URL does not match filter conditions", "event_summary", eventSummary.Value, "filter", filter.Description)
			return false // event doesn't match
		}
	}

	// VEvent must match if we get here
	slog.Debug("Event matches filter conditions", "event_summary", eventSummary.Value, "filter", filter.Description)
	return true
}

// Applies filter transformations to a VEvent pointer
func (filter Filter) transformEvent(event *ics.VEvent) {

	// Summary transformations
	if filter.Transform.Summary.Remove {
		event.SetSummary("")
	} else if filter.Transform.Summary.Replace != "" {
		event.SetSummary(filter.Transform.Summary.Replace)
	}

	// Description transformations
	if filter.Transform.Description.Remove {
		event.SetDescription("")
	} else if filter.Transform.Description.Replace != "" {
		event.SetDescription(filter.Transform.Description.Replace)
	}

	// Location transformations
	if filter.Transform.Location.Remove {
		event.SetLocation("")
	} else if filter.Transform.Location.Replace != "" {
		event.SetLocation(filter.Transform.Location.Replace)
	
	// URL transformations
	if filter.Transform.Url.Remove {
		event.SetUrl("")
	} else if filter.Transform.Url.Replace != "" {
		event.SetUrl(filter.Transform.Url.Replace)
	}
}

// EventMatchRules contains VEvent properties that user can match against
type EventMatchRules struct {
	Summary     StringMatchRule `yaml:"summary"`
	Description StringMatchRule `yaml:"description"`
	Location    StringMatchRule `yaml:"location"`
	Url    		StringMatchRule `yaml:"url"`
}

// StringMatchRule defines match rules for VEvent properties with string values
type StringMatchRule struct {
	Null       bool   `yaml:"empty"`
	Contains   string `yaml:"contains"`
	Prefix     string `yaml:"prefix"`
	Suffix     string `yaml:"suffix"`
	RegexMatch string `yaml:"regex"`
}

// Returns true if StringMatchRule has any conditions
func (smr StringMatchRule) hasConditions() bool {
	return smr.Null ||
	    smr.Contains != "" ||
		smr.Prefix != "" ||
		smr.Suffix != "" ||
		smr.RegexMatch != ""
}

// Returns true if a given string (data) matches ALL StringMatchRule conditions
func (smr StringMatchRule) matchesString(data string) bool {
	// check null if set and don't process further - this condition can only be met on its own
	if smr.Null {
		return data == ""
	}
	// check contains if set
	if smr.Contains != "" {
		if data == "" || !strings.Contains(data, smr.Contains) {
			return false
		}
	}
	// check prefix if set
	if smr.Prefix != "" {
		if data == "" || !strings.HasPrefix(data, smr.Prefix) {
			return false
		}
	}
	// check suffix if set
	if smr.Suffix != "" {
		if data == "" || !strings.HasSuffix(data, smr.Suffix) {
			return false
		}
	}
	// check regex match if set
	if smr.RegexMatch != "" {
		re, err := regexp.Compile(smr.RegexMatch)
		if err != nil {
				slog.Warn("error processing regex rule", "value", smr.RegexMatch)
				return false // regex error is considered a failure to match
		}
		match := re.MatchString(data)
		if !match {
				return false // regex didn't match
		}
	}
return true
}

// EventTransformRules contains VEvent properties that user can modify
type EventTransformRules struct {
	Summary     StringTransformRule `yaml:"summary"`
	Description StringTransformRule `yaml:"description"`
	Location    StringTransformRule `yaml:"location"`
	Url    		StringTransformRule `yaml:"url"`
}

// StringTransformRule defines changes for VEvent properties with string values
type StringTransformRule struct {
	Replace string `yaml:"replace"`
	Remove  bool   `yaml:"remove"`
}
