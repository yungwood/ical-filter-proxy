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
	stringTransforms := []eventStringTransform{
		{property: ics.ComponentPropertySummary, rule: filter.Transform.Summary, set: func(value string) { event.SetSummary(value) }},
		{property: ics.ComponentPropertyDescription, rule: filter.Transform.Description, set: func(value string) { event.SetDescription(value) }},
		{property: ics.ComponentPropertyLocation, rule: filter.Transform.Location, set: func(value string) { event.SetLocation(value) }},
		{property: ics.ComponentPropertyUrl, rule: filter.Transform.URL, set: func(value string) { event.SetURL(value) }},
	}
	for _, transform := range stringTransforms {
		applyEventStringTransform(*event, transform)
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

type eventStringTransform struct {
	property ics.ComponentProperty
	rule     StringTransformRule
	set      func(string)
}

func eventStringPropertyMatches(event ics.VEvent, property ics.ComponentProperty, rule StringMatchRule) bool {
	if !rule.hasConditions() {
		return true
	}

	return rule.matchesString(eventStringProperty(event, property))
}

func applyEventStringTransform(event ics.VEvent, transform eventStringTransform) {
	if !transform.rule.hasActions() {
		return
	}

	value := eventStringProperty(event, transform.property)
	transform.set(applyStringTransform(value, transform.rule))
}

func eventStringProperty(event ics.VEvent, property ics.ComponentProperty) string {
	eventProperty := event.GetProperty(property)
	if eventProperty == nil {
		return ""
	}

	return eventProperty.Value
}
