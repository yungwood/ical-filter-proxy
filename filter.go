package main

import (
	"log/slog"

	ics "github.com/arran4/golang-ical"
)

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

	// Check URL filters against VEvent
	if filter.Match.URL.hasConditions() {
		eventURL := event.GetProperty(ics.ComponentPropertyUrl)
		var eventURLValue string
		if eventURL == nil {
			eventURLValue = ""
		} else {
			eventURLValue = eventURL.Value
		}
		if !filter.Match.URL.matchesString(eventURLValue) {
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
	eventSummary := event.GetProperty(ics.ComponentPropertySummary)
	var eventSummaryValue string
	if eventSummary == nil {
		eventSummaryValue = ""
	} else {
		eventSummaryValue = eventSummary.Value
	}
	if filter.Transform.Summary.Remove {
		event.SetSummary("")
	}
	if filter.Transform.Summary.Replace != "" {
		event.SetSummary(filter.Transform.Summary.Replace)
	}
	if filter.Transform.Summary.Prefix != "" {
		event.SetSummary(filter.Transform.Summary.Prefix + eventSummaryValue)
	}
	if filter.Transform.Summary.Suffix != "" {
		event.SetSummary(eventSummaryValue + filter.Transform.Summary.Suffix)
	}

	// Description transformations
	eventDescription := event.GetProperty(ics.ComponentPropertyDescription)
	var eventDescriptionValue string
	if eventDescription == nil {
		eventDescriptionValue = ""
	} else {
		eventDescriptionValue = eventDescription.Value
	}
	if filter.Transform.Description.Remove {
		event.SetDescription("")
	}
	if filter.Transform.Description.Replace != "" {
		event.SetDescription(filter.Transform.Description.Replace)
	}
	if filter.Transform.Description.Prefix != "" {
		event.SetDescription(filter.Transform.Description.Prefix + eventDescriptionValue)
	}
	if filter.Transform.Description.Suffix != "" {
		event.SetDescription(eventDescriptionValue + filter.Transform.Description.Suffix)
	}

	// Location transformations
	if filter.Transform.Location.Remove {
		event.SetLocation("")
	} else if filter.Transform.Location.Replace != "" {
		event.SetLocation(filter.Transform.Location.Replace)
	}

	// URL transformations
	if filter.Transform.URL.Remove {
		event.SetURL("")
	} else if filter.Transform.URL.Replace != "" {
		event.SetURL(filter.Transform.URL.Replace)
	}
}

// EventMatchRules contains VEvent properties that user can match against
type EventMatchRules struct {
	Summary     StringMatchRule `yaml:"summary"`
	Description StringMatchRule `yaml:"description"`
	Location    StringMatchRule `yaml:"location"`
	URL         StringMatchRule `yaml:"url"`
}

// EventTransformRules contains VEvent properties that user can modify
type EventTransformRules struct {
	Summary     StringTransformRule `yaml:"summary"`
	Description StringTransformRule `yaml:"description"`
	Location    StringTransformRule `yaml:"location"`
	URL         StringTransformRule `yaml:"url"`
}
