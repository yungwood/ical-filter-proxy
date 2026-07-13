package main

import (
	"testing"

	ics "github.com/arran4/golang-ical"
)

func TestCalendarConfigProcessEvent(t *testing.T) {
	tests := []struct {
		name        string
		config      CalendarConfig
		event       *ics.VEvent
		wantKeep    bool
		wantSummary string
	}{
		{
			name: "keeps event by default when no filters match",
			config: CalendarConfig{
				Filters: []Filter{
					{
						Description: "Remove canceled events",
						RemoveEvent: true,
						Match: EventMatchRules{
							Summary: StringMatchRule{Prefix: "Canceled: "},
						},
					},
				},
			},
			event:       testEvent("Team sync"),
			wantKeep:    true,
			wantSummary: "Team sync",
		},
		{
			name: "removes event when remove filter matches",
			config: CalendarConfig{
				Filters: []Filter{
					{
						Description: "Remove canceled events",
						RemoveEvent: true,
						Match: EventMatchRules{
							Summary: StringMatchRule{Prefix: "Canceled: "},
						},
					},
				},
			},
			event:       testEvent("Canceled: Team sync"),
			wantKeep:    false,
			wantSummary: "Canceled: Team sync",
		},
		{
			name: "transforms matching event and keeps it",
			config: CalendarConfig{
				Filters: []Filter{
					{
						Description: "Rename on-call events",
						Match: EventMatchRules{
							Summary: StringMatchRule{Contains: "schedule: oncall"},
						},
						Transform: EventTransformRules{
							Summary: StringTransformRule{Replace: "On-Call"},
						},
					},
				},
			},
			event:       testEvent("ops schedule: oncall"),
			wantKeep:    true,
			wantSummary: "On-Call",
		},
		{
			name: "stop prevents later matching filters",
			config: CalendarConfig{
				Filters: []Filter{
					{
						Description: "Rename and stop",
						Stop:        true,
						Transform: EventTransformRules{
							Summary: StringTransformRule{Replace: "Keep Me"},
						},
					},
					{
						Description: "Remove everything else",
						RemoveEvent: true,
					},
				},
			},
			event:       testEvent("Original"),
			wantKeep:    true,
			wantSummary: "Keep Me",
		},
		{
			name: "transform continues to later filters when stop is false",
			config: CalendarConfig{
				Filters: []Filter{
					{
						Description: "Rename without stop",
						Transform: EventTransformRules{
							Summary: StringTransformRule{Replace: "Remove Me"},
						},
					},
					{
						Description: "Remove transformed event",
						RemoveEvent: true,
						Match: EventMatchRules{
							Summary: StringMatchRule{Contains: "Remove"},
						},
					},
				},
			},
			event:       testEvent("Original"),
			wantKeep:    false,
			wantSummary: "Remove Me",
		},
		{
			name: "filter without match rules matches all events",
			config: CalendarConfig{
				Filters: []Filter{
					{
						Description: "Remove everything",
						RemoveEvent: true,
					},
				},
			},
			event:       testEvent("Original"),
			wantKeep:    false,
			wantSummary: "Original",
		},
		{
			name: "drops event without summary",
			config: CalendarConfig{
				Filters: []Filter{},
			},
			event:       ics.NewEvent("missing-summary"),
			wantKeep:    false,
			wantSummary: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKeep := tt.config.ProcessEvent(tt.event)
			if gotKeep != tt.wantKeep {
				t.Fatalf("ProcessEvent() = %v, want %v", gotKeep, tt.wantKeep)
			}

			if tt.wantSummary == "" {
				return
			}

			gotSummary := tt.event.GetProperty(ics.ComponentPropertySummary)
			if gotSummary == nil {
				t.Fatalf("summary property is nil, want %q", tt.wantSummary)
			}
			if gotSummary.Value != tt.wantSummary {
				t.Fatalf("summary = %q, want %q", gotSummary.Value, tt.wantSummary)
			}
		})
	}
}
