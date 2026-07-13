package main

import (
	"testing"

	ics "github.com/arran4/golang-ical"
)

func TestFilterMatchesEvent(t *testing.T) {
	tests := []struct {
		name  string
		event *ics.VEvent
		match EventMatchRules
		want  bool
	}{
		{
			name:  "filter without match rules matches event",
			event: testEvent("team calendar event"),
			match: EventMatchRules{},
			want:  true,
		},
		{
			name:  "summary matches",
			event: testEvent("Canceled: team calendar event"),
			match: EventMatchRules{
				Summary: StringMatchRule{Prefix: "Canceled: "},
			},
			want: true,
		},
		{
			name:  "summary mismatch rejects event",
			event: testEvent("team calendar event"),
			match: EventMatchRules{
				Summary: StringMatchRule{Prefix: "Canceled: "},
			},
			want: false,
		},
		{
			name: "description matches",
			event: testEvent("team calendar event", func(event *ics.VEvent) {
				event.SetDescription("weekday shift")
			}),
			match: EventMatchRules{
				Description: StringMatchRule{Contains: "shift"},
			},
			want: true,
		},
		{
			name:  "missing description matches empty",
			event: testEvent("team calendar event"),
			match: EventMatchRules{
				Description: StringMatchRule{Null: true},
			},
			want: true,
		},
		{
			name: "location matches",
			event: testEvent("team calendar event", func(event *ics.VEvent) {
				event.SetLocation("Adelaide")
			}),
			match: EventMatchRules{
				Location: StringMatchRule{Suffix: "laide"},
			},
			want: true,
		},
		{
			name:  "missing location matches empty",
			event: testEvent("team calendar event"),
			match: EventMatchRules{
				Location: StringMatchRule{Null: true},
			},
			want: true,
		},
		{
			name: "url matches",
			event: testEvent("team calendar event", func(event *ics.VEvent) {
				event.SetURL("https://example.com/event")
			}),
			match: EventMatchRules{
				URL: StringMatchRule{Contains: "example.com"},
			},
			want: true,
		},
		{
			name:  "missing url matches empty",
			event: testEvent("team calendar event"),
			match: EventMatchRules{
				URL: StringMatchRule{Null: true},
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
			match: EventMatchRules{
				Summary:     StringMatchRule{Contains: "calendar"},
				Description: StringMatchRule{Contains: "shift"},
				Location:    StringMatchRule{Prefix: "Adel"},
				URL:         StringMatchRule{Suffix: "/event"},
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
			match: EventMatchRules{
				Summary:     StringMatchRule{Contains: "calendar"},
				Description: StringMatchRule{Contains: "holiday"},
				Location:    StringMatchRule{Prefix: "Adel"},
				URL:         StringMatchRule{Suffix: "/event"},
			},
			want: false,
		},
		{
			name:  "event without summary never matches",
			event: ics.NewEvent("missing-summary"),
			match: EventMatchRules{},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := Filter{Match: tt.match}
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
			name: "prefix and suffix summary and description",
			transform: EventTransformRules{
				Summary:     StringTransformRule{Prefix: "[", Suffix: "]"},
				Description: StringTransformRule{Prefix: "(", Suffix: ")"},
			},
			want: map[ics.ComponentProperty]string{
				ics.ComponentPropertySummary:     "[original summary]",
				ics.ComponentPropertyDescription: "(original description)",
				ics.ComponentPropertyLocation:    "original location",
				ics.ComponentPropertyUrl:         "https://example.com/original",
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

			filter := Filter{Transform: tt.transform}
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
