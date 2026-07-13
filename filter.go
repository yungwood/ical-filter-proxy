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

	stringMatches := []eventStringMatch{
		{property: ics.ComponentPropertySummary, name: "summary", rule: filter.Match.Summary},
		{property: ics.ComponentPropertyDescription, name: "description", rule: filter.Match.Description},
		{property: ics.ComponentPropertyLocation, name: "location", rule: filter.Match.Location},
		{property: ics.ComponentPropertyUrl, name: "url", rule: filter.Match.URL},
	}
	for _, match := range stringMatches {
		if !eventStringPropertyMatches(event, match.property, match.rule) {
			slog.Debug("Event property does not match filter conditions", "property", match.name, "event_summary", eventSummary.Value, "filter", filter.Description)
			return false
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
	eventLocationValue := eventStringProperty(*event, ics.ComponentPropertyLocation)
	if filter.Transform.Location.hasActions() {
		event.SetLocation(applyStringTransform(eventLocationValue, filter.Transform.Location))
	}

	// URL transformations
	eventURLValue := eventStringProperty(*event, ics.ComponentPropertyUrl)
	if filter.Transform.URL.hasActions() {
		event.SetURL(applyStringTransform(eventURLValue, filter.Transform.URL))
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

type eventStringMatch struct {
	property ics.ComponentProperty
	name     string
	rule     StringMatchRule
}

func eventStringPropertyMatches(event ics.VEvent, property ics.ComponentProperty, rule StringMatchRule) bool {
	if !rule.hasConditions() {
		return true
	}

	return rule.matchesString(eventStringProperty(event, property))
}

func eventStringProperty(event ics.VEvent, property ics.ComponentProperty) string {
	eventProperty := event.GetProperty(property)
	if eventProperty == nil {
		return ""
	}

	return eventProperty.Value
}
