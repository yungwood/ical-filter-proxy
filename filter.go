package main

import (
	"fmt"
	"log/slog"

	ics "github.com/arran4/golang-ical"
)

// FilterConfig is the YAML-backed filter configuration. It is compiled into
// Filter before runtime event processing.
type FilterConfig struct {
	Description string                `yaml:"description"`
	RemoveEvent bool                  `yaml:"remove"`
	Stop        bool                  `yaml:"stop"`
	Match       EventMatchRulesConfig `yaml:"match"`
	Transform   EventTransformRules   `yaml:"transform"`
}

// Filter is the runtime filter representation. Match rules are compiled
// once during config preparation; transforms remain raw because they are already
// directly executable.
type Filter struct {
	Description string
	RemoveEvent bool
	Stop        bool
	Match       EventMatchRules
	Transform   EventTransformRules
}

func (filter FilterConfig) compile() (Filter, error) {
	match, err := filter.Match.compile()
	if err != nil {
		return Filter{}, err
	}

	return Filter{
		Description: filter.Description,
		RemoveEvent: filter.RemoveEvent,
		Stop:        filter.Stop,
		Match:       match,
		Transform:   filter.Transform,
	}, nil
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

// EventMatchRulesConfig contains YAML-backed match rules for VEvent properties.
type EventMatchRulesConfig struct {
	Summary     StringMatchRuleConfig `yaml:"summary"`
	Description StringMatchRuleConfig `yaml:"description"`
	Location    StringMatchRuleConfig `yaml:"location"`
	URL         StringMatchRuleConfig `yaml:"url"`
}

// EventMatchRules contains runtime-ready match rules for VEvent
// properties.
type EventMatchRules struct {
	Summary     StringMatchRule
	Description StringMatchRule
	Location    StringMatchRule
	URL         StringMatchRule
}

func (rules EventMatchRulesConfig) compile() (EventMatchRules, error) {
	summary, err := rules.Summary.compile()
	if err != nil {
		return EventMatchRules{}, fmt.Errorf("summary: %w", err)
	}
	description, err := rules.Description.compile()
	if err != nil {
		return EventMatchRules{}, fmt.Errorf("description: %w", err)
	}
	location, err := rules.Location.compile()
	if err != nil {
		return EventMatchRules{}, fmt.Errorf("location: %w", err)
	}
	url, err := rules.URL.compile()
	if err != nil {
		return EventMatchRules{}, fmt.Errorf("url: %w", err)
	}

	return EventMatchRules{
		Summary:     summary,
		Description: description,
		Location:    location,
		URL:         url,
	}, nil
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
