package main

import (
	"testing"

	ics "github.com/arran4/golang-ical"
)

func TestFilterMatchesEvent(t *testing.T) {
	tests := []struct {
		name  string
		event *ics.VEvent
		match EventMatchRulesConfig
		want  bool
	}{
		{
			name:  "filter without match rules matches event",
			event: testEvent("team calendar event"),
			match: EventMatchRulesConfig{},
			want:  true,
		},
		{
			name:  "summary matches",
			event: testEvent("Canceled: team calendar event"),
			match: EventMatchRulesConfig{
				Summary: StringMatchRuleConfig{Prefix: "Canceled: "},
			},
			want: true,
		},
		{
			name:  "summary mismatch rejects event",
			event: testEvent("team calendar event"),
			match: EventMatchRulesConfig{
				Summary: StringMatchRuleConfig{Prefix: "Canceled: "},
			},
			want: false,
		},
		{
			name: "description matches",
			event: testEvent("team calendar event", func(event *ics.VEvent) {
				event.SetDescription("weekday shift")
			}),
			match: EventMatchRulesConfig{
				Description: StringMatchRuleConfig{Contains: "shift"},
			},
			want: true,
		},
		{
			name:  "missing description matches empty",
			event: testEvent("team calendar event"),
			match: EventMatchRulesConfig{
				Description: StringMatchRuleConfig{Null: true},
			},
			want: true,
		},
		{
			name: "location matches",
			event: testEvent("team calendar event", func(event *ics.VEvent) {
				event.SetLocation("Adelaide")
			}),
			match: EventMatchRulesConfig{
				Location: StringMatchRuleConfig{Suffix: "laide"},
			},
			want: true,
		},
		{
			name:  "missing location matches empty",
			event: testEvent("team calendar event"),
			match: EventMatchRulesConfig{
				Location: StringMatchRuleConfig{Null: true},
			},
			want: true,
		},
		{
			name: "url matches",
			event: testEvent("team calendar event", func(event *ics.VEvent) {
				event.SetURL("https://example.com/event")
			}),
			match: EventMatchRulesConfig{
				URL: StringMatchRuleConfig{Contains: "example.com"},
			},
			want: true,
		},
		{
			name:  "missing url matches empty",
			event: testEvent("team calendar event"),
			match: EventMatchRulesConfig{
				URL: StringMatchRuleConfig{Null: true},
			},
			want: true,
		},
		{
			name: "all configured properties must match",
			event: testEvent("team calendar event", func(event *ics.VEvent) {
				event.SetDescription("weekday shift")
				event.SetLocation("Adelaide")
				event.SetURL("https://example.com/event")
			}),
			match: EventMatchRulesConfig{
				Summary:     StringMatchRuleConfig{Contains: "calendar"},
				Description: StringMatchRuleConfig{Contains: "shift"},
				Location:    StringMatchRuleConfig{Prefix: "Adel"},
				URL:         StringMatchRuleConfig{Suffix: "/event"},
			},
			want: true,
		},
		{
			name: "one configured property mismatch rejects event",
			event: testEvent("team calendar event", func(event *ics.VEvent) {
				event.SetDescription("weekday shift")
				event.SetLocation("Adelaide")
				event.SetURL("https://example.com/event")
			}),
			match: EventMatchRulesConfig{
				Summary:     StringMatchRuleConfig{Contains: "calendar"},
				Description: StringMatchRuleConfig{Contains: "holiday"},
				Location:    StringMatchRuleConfig{Prefix: "Adel"},
				URL:         StringMatchRuleConfig{Suffix: "/event"},
			},
			want: false,
		},
		{
			name:  "event without summary never matches",
			event: ics.NewEvent("missing-summary"),
			match: EventMatchRulesConfig{},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := mustCompileFilter(t, FilterConfig{Match: tt.match})
			got := filter.matchesEvent(*tt.event)
			if got != tt.want {
				t.Fatalf("matchesEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventStringProperty(t *testing.T) {
	event := testEvent("team calendar event")

	if got := eventStringProperty(*event, ics.ComponentPropertySummary); got != "team calendar event" {
		t.Fatalf("summary = %q, want team calendar event", got)
	}
	if got := eventStringProperty(*event, ics.ComponentPropertyDescription); got != "" {
		t.Fatalf("missing description = %q, want empty string", got)
	}
}

func TestEventStringPropertyMatches(t *testing.T) {
	tests := []struct {
		name     string
		event    *ics.VEvent
		property ics.ComponentProperty
		rule     StringMatchRule
		want     bool
	}{
		{
			name:     "empty rule matches",
			event:    testEvent("team calendar event"),
			property: ics.ComponentPropertySummary,
			rule:     mustCompileStringMatchRule(t, StringMatchRuleConfig{}),
			want:     true,
		},
		{
			name:     "property value matches",
			event:    testEvent("team calendar event"),
			property: ics.ComponentPropertySummary,
			rule:     mustCompileStringMatchRule(t, StringMatchRuleConfig{Contains: "calendar"}),
			want:     true,
		},
		{
			name:     "property value mismatch",
			event:    testEvent("team calendar event"),
			property: ics.ComponentPropertySummary,
			rule:     mustCompileStringMatchRule(t, StringMatchRuleConfig{Contains: "holiday"}),
			want:     false,
		},
		{
			name:     "missing property matches empty rule",
			event:    testEvent("team calendar event"),
			property: ics.ComponentPropertyDescription,
			rule:     mustCompileStringMatchRule(t, StringMatchRuleConfig{Null: true}),
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := eventStringPropertyMatches(*tt.event, tt.property, tt.rule)
			if got != tt.want {
				t.Fatalf("eventStringPropertyMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyEventStringTransform(t *testing.T) {
	t.Run("empty transform does not call setter", func(t *testing.T) {
		called := false
		event := testEvent("team calendar event")

		applyEventStringTransform(*event, eventStringTransform{
			property: ics.ComponentPropertySummary,
			rule:     StringTransformRule{},
			set: func(_ string) {
				called = true
			},
		})

		if called {
			t.Fatal("setter called for empty transform")
		}
	})

	t.Run("transform uses existing property value", func(t *testing.T) {
		var got string
		event := testEvent("team calendar event")

		applyEventStringTransform(*event, eventStringTransform{
			property: ics.ComponentPropertySummary,
			rule:     StringTransformRule{Prefix: "[", Suffix: "]"},
			set: func(value string) {
				got = value
			},
		})

		if got != "[team calendar event]" {
			t.Fatalf("setter value = %q, want %q", got, "[team calendar event]")
		}
	})

	t.Run("missing property transforms empty value", func(t *testing.T) {
		var got string
		event := testEvent("team calendar event")

		applyEventStringTransform(*event, eventStringTransform{
			property: ics.ComponentPropertyDescription,
			rule:     StringTransformRule{Prefix: "missing: "},
			set: func(value string) {
				got = value
			},
		})

		if got != "missing: " {
			t.Fatalf("setter value = %q, want %q", got, "missing: ")
		}
	})
}

func TestFilterTransformEvent(t *testing.T) {
	tests := []struct {
		name      string
		transform EventTransformRules
		want      map[ics.ComponentProperty]string
	}{
		{
			name: "replace fields",
			transform: EventTransformRules{
				Summary:     StringTransformRule{Replace: "new summary"},
				Description: StringTransformRule{Replace: "new description"},
				Location:    StringTransformRule{Replace: "new location"},
				URL:         StringTransformRule{Replace: "https://example.com/new"},
			},
			want: map[ics.ComponentProperty]string{
				ics.ComponentPropertySummary:     "new summary",
				ics.ComponentPropertyDescription: "new description",
				ics.ComponentPropertyLocation:    "new location",
				ics.ComponentPropertyUrl:         "https://example.com/new",
			},
		},
		{
			name: "remove fields",
			transform: EventTransformRules{
				Summary:     StringTransformRule{Remove: true},
				Description: StringTransformRule{Remove: true},
				Location:    StringTransformRule{Remove: true},
				URL:         StringTransformRule{Remove: true},
			},
			want: map[ics.ComponentProperty]string{
				ics.ComponentPropertySummary:     "",
				ics.ComponentPropertyDescription: "",
				ics.ComponentPropertyLocation:    "",
				ics.ComponentPropertyUrl:         "",
			},
		},
		{
			name: "prefix and suffix fields",
			transform: EventTransformRules{
				Summary:     StringTransformRule{Prefix: "[", Suffix: "]"},
				Description: StringTransformRule{Prefix: "(", Suffix: ")"},
				Location:    StringTransformRule{Prefix: "Location: "},
				URL:         StringTransformRule{Suffix: "?tracked=true"},
			},
			want: map[ics.ComponentProperty]string{
				ics.ComponentPropertySummary:     "[original summary]",
				ics.ComponentPropertyDescription: "(original description)",
				ics.ComponentPropertyLocation:    "Location: original location",
				ics.ComponentPropertyUrl:         "https://example.com/original?tracked=true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := testEvent("original summary", func(event *ics.VEvent) {
				event.SetDescription("original description")
				event.SetLocation("original location")
				event.SetURL("https://example.com/original")
			})

			filter := mustCompileFilter(t, FilterConfig{Transform: tt.transform})
			filter.transformEvent(event)

			for property, want := range tt.want {
				got := event.GetProperty(property)
				if got == nil {
					t.Fatalf("%s property is nil, want %q", property, want)
				}
				if got.Value != want {
					t.Fatalf("%s = %q, want %q", property, got.Value, want)
				}
			}
		})
	}
}

func testEvent(summary string, options ...func(*ics.VEvent)) *ics.VEvent {
	event := ics.NewEvent(summary)
	event.SetSummary(summary)

	for _, option := range options {
		option(event)
	}

	return event
}

func mustCompileFilter(t *testing.T, filter FilterConfig) Filter {
	t.Helper()

	compiledFilter, err := filter.compile()
	if err != nil {
		t.Fatalf("compile() returned error: %v", err)
	}

	return compiledFilter
}

func mustCompileFilters(t *testing.T, filters []FilterConfig) []Filter {
	t.Helper()

	compiledFilters := make([]Filter, 0, len(filters))
	for _, filter := range filters {
		compiledFilters = append(compiledFilters, mustCompileFilter(t, filter))
	}

	return compiledFilters
}

func mustCompileStringMatchRule(t *testing.T, rule StringMatchRuleConfig) StringMatchRule {
	t.Helper()

	compiledRule, err := rule.compile()
	if err != nil {
		t.Fatalf("compile() returned error: %v", err)
	}

	return compiledRule
}
