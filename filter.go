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
		eventDescriptionValue := eventStringProperty(event, ics.ComponentPropertyDescription)
		if !filter.Match.Description.matchesString(eventDescriptionValue) {
			slog.Debug("Event Description does not match filter conditions", "event_summary", eventSummary.Value, "filter", filter.Description)
			return false // event doesn't match
		}
	}

	// Check Location filters against VEvent
	if filter.Match.Location.hasConditions() {
		eventLocationValue := eventStringProperty(event, ics.ComponentPropertyLocation)
		if !filter.Match.Location.matchesString(eventLocationValue) {
			slog.Debug("Event Location does not match filter conditions", "event_summary", eventSummary.Value, "filter", filter.Description)
			return false // event doesn't match

		}
	}

	// Check URL filters against VEvent
	if filter.Match.URL.hasConditions() {
		eventURLValue := eventStringProperty(event, ics.ComponentPropertyUrl)
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
	eventSummaryValue := eventStringProperty(*event, ics.ComponentPropertySummary)
	if filter.Transform.Summary.hasActions() {
		event.SetSummary(applyStringTransform(eventSummaryValue, filter.Transform.Summary))
	}

	// Description transformations
	eventDescriptionValue := eventStringProperty(*event, ics.ComponentPropertyDescription)
	if filter.Transform.Description.hasActions() {
		event.SetDescription(applyStringTransform(eventDescriptionValue, filter.Transform.Description))
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

func eventStringProperty(event ics.VEvent, property ics.ComponentProperty) string {
	eventProperty := event.GetProperty(property)
	if eventProperty == nil {
		return ""
	}

	return eventProperty.Value
}
