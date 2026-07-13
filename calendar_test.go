package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ics "github.com/arran4/golang-ical"
)

func TestCalendarFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		_, _ = w.Write([]byte(testCalendarFeed))
	}))
	defer server.Close()

	config := Calendar{
		Name:        "test",
		PublishName: "Filtered Calendar",
		FeedURL:     server.URL,
		Filters: mustCompileFilters(t, []FilterConfig{
			{
				Description: "Remove canceled events",
				RemoveEvent: true,
				Match: EventMatchRules{
					Summary: StringMatchRuleConfig{Prefix: "Canceled: "},
				},
			},
			{
				Description: "Rename on-call event",
				Match: EventMatchRules{
					Summary: StringMatchRuleConfig{Contains: "schedule: oncall"},
				},
				Transform: EventTransformRules{
					Summary: StringTransformRule{Replace: "On-Call"},
				},
			},
		}),
	}

	feed, err := config.fetch(context.Background())
	if err != nil {
		t.Fatalf("fetch() returned error: %v", err)
	}

	calendar, err := ics.ParseCalendar(strings.NewReader(string(feed)))
	if err != nil {
		t.Fatalf("ParseCalendar() returned error: %v", err)
	}

	if !strings.Contains(string(feed), "X-WR-CALNAME:Filtered Calendar") {
		t.Fatalf("serialized feed does not contain published calendar name:\n%s", string(feed))
	}

	events := calendar.Events()
	if len(events) != 2 {
		t.Fatalf("len(Events()) = %d, want 2", len(events))
	}

	summaries := map[string]bool{}
	for _, event := range events {
		summary := event.GetProperty(ics.ComponentPropertySummary)
		if summary == nil {
			t.Fatal("event summary is nil")
		}
		summaries[summary.Value] = true
	}

	if !summaries["On-Call"] {
		t.Fatal("transformed On-Call event missing")
	}
	if !summaries["Team sync"] {
		t.Fatal("unmatched Team sync event missing")
	}
	if summaries["Canceled: Team sync"] {
		t.Fatal("canceled event was not removed")
	}
}

func TestCalendarProcessEvent(t *testing.T) {
	tests := []struct {
		name        string
		config      Calendar
		event       *ics.VEvent
		wantKeep    bool
		wantSummary string
	}{
		{
			name: "keeps event by default when no filters match",
			config: Calendar{
				Filters: mustCompileFilters(t, []FilterConfig{
					{
						Description: "Remove canceled events",
						RemoveEvent: true,
						Match: EventMatchRules{
							Summary: StringMatchRuleConfig{Prefix: "Canceled: "},
						},
					},
				}),
			},
			event:       testEvent("Team sync"),
			wantKeep:    true,
			wantSummary: "Team sync",
		},
		{
			name: "removes event when remove filter matches",
			config: Calendar{
				Filters: mustCompileFilters(t, []FilterConfig{
					{
						Description: "Remove canceled events",
						RemoveEvent: true,
						Match: EventMatchRules{
							Summary: StringMatchRuleConfig{Prefix: "Canceled: "},
						},
					},
				}),
			},
			event:       testEvent("Canceled: Team sync"),
			wantKeep:    false,
			wantSummary: "Canceled: Team sync",
		},
		{
			name: "transforms matching event and keeps it",
			config: Calendar{
				Filters: mustCompileFilters(t, []FilterConfig{
					{
						Description: "Rename on-call events",
						Match: EventMatchRules{
							Summary: StringMatchRuleConfig{Contains: "schedule: oncall"},
						},
						Transform: EventTransformRules{
							Summary: StringTransformRule{Replace: "On-Call"},
						},
					},
				}),
			},
			event:       testEvent("ops schedule: oncall"),
			wantKeep:    true,
			wantSummary: "On-Call",
		},
		{
			name: "stop prevents later matching filters",
			config: Calendar{
				Filters: mustCompileFilters(t, []FilterConfig{
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
				}),
			},
			event:       testEvent("Original"),
			wantKeep:    true,
			wantSummary: "Keep Me",
		},
		{
			name: "transform continues to later filters when stop is false",
			config: Calendar{
				Filters: mustCompileFilters(t, []FilterConfig{
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
							Summary: StringMatchRuleConfig{Contains: "Remove"},
						},
					},
				}),
			},
			event:       testEvent("Original"),
			wantKeep:    false,
			wantSummary: "Remove Me",
		},
		{
			name: "filter without match rules matches all events",
			config: Calendar{
				Filters: mustCompileFilters(t, []FilterConfig{
					{
						Description: "Remove everything",
						RemoveEvent: true,
					},
				}),
			},
			event:       testEvent("Original"),
			wantKeep:    false,
			wantSummary: "Original",
		},
		{
			name: "drops event without summary",
			config: Calendar{
				Filters: mustCompileFilters(t, []FilterConfig{}),
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

const testCalendarFeed = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//ical-filter-proxy//test//EN
BEGIN:VEVENT
UID:remove-1
DTSTAMP:20260713T000000Z
DTSTART:20260713T010000Z
DTEND:20260713T020000Z
SUMMARY:Canceled: Team sync
END:VEVENT
BEGIN:VEVENT
UID:transform-1
DTSTAMP:20260713T000000Z
DTSTART:20260713T030000Z
DTEND:20260713T040000Z
SUMMARY:ops schedule: oncall
END:VEVENT
BEGIN:VEVENT
UID:keep-1
DTSTAMP:20260713T000000Z
DTSTART:20260713T050000Z
DTEND:20260713T060000Z
SUMMARY:Team sync
END:VEVENT
END:VCALENDAR`
