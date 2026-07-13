package main

import "testing"

func TestStringMatchRuleHasConditions(t *testing.T) {
	tests := []struct {
		name string
		rule StringMatchRule
		want bool
	}{
		{
			name: "empty rule",
			rule: StringMatchRule{},
			want: false,
		},
		{
			name: "empty condition",
			rule: StringMatchRule{Null: true},
			want: true,
		},
		{
			name: "contains condition",
			rule: StringMatchRule{Contains: "event"},
			want: true,
		},
		{
			name: "prefix condition",
			rule: StringMatchRule{Prefix: "event"},
			want: true,
		},
		{
			name: "suffix condition",
			rule: StringMatchRule{Suffix: "event"},
			want: true,
		},
		{
			name: "regex condition",
			rule: StringMatchRule{RegexMatch: "event"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rule.hasConditions()
			if got != tt.want {
				t.Fatalf("hasConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringMatchRuleMatchesString(t *testing.T) {
	tests := []struct {
		name string
		rule StringMatchRule
		data string
		want bool
	}{
		{
			name: "empty rule matches non-empty string",
			rule: StringMatchRule{},
			data: "calendar event",
			want: true,
		},
		{
			name: "empty rule matches empty string",
			rule: StringMatchRule{},
			data: "",
			want: true,
		},
		{
			name: "empty condition matches empty string",
			rule: StringMatchRule{Null: true},
			data: "",
			want: true,
		},
		{
			name: "empty condition rejects non-empty string",
			rule: StringMatchRule{Null: true},
			data: "calendar event",
			want: false,
		},
		{
			name: "contains matches",
			rule: StringMatchRule{Contains: "event"},
			data: "calendar event",
			want: true,
		},
		{
			name: "contains rejects missing value",
			rule: StringMatchRule{Contains: "shift"},
			data: "calendar event",
			want: false,
		},
		{
			name: "contains rejects empty data",
			rule: StringMatchRule{Contains: "event"},
			data: "",
			want: false,
		},
		{
			name: "prefix matches",
			rule: StringMatchRule{Prefix: "calendar"},
			data: "calendar event",
			want: true,
		},
		{
			name: "prefix rejects missing prefix",
			rule: StringMatchRule{Prefix: "event"},
			data: "calendar event",
			want: false,
		},
		{
			name: "suffix matches",
			rule: StringMatchRule{Suffix: "event"},
			data: "calendar event",
			want: true,
		},
		{
			name: "suffix rejects missing suffix",
			rule: StringMatchRule{Suffix: "calendar"},
			data: "calendar event",
			want: false,
		},
		{
			name: "regex matches",
			rule: StringMatchRule{RegexMatch: `cal.*event`},
			data: "calendar event",
			want: true,
		},
		{
			name: "regex rejects missing match",
			rule: StringMatchRule{RegexMatch: `^event`},
			data: "calendar event",
			want: false,
		},
		{
			name: "invalid regex rejects",
			rule: StringMatchRule{RegexMatch: `[`},
			data: "calendar event",
			want: false,
		},
		{
			name: "combined conditions all match",
			rule: StringMatchRule{
				Contains:   "calendar",
				Prefix:     "team",
				Suffix:     "event",
				RegexMatch: `team .* event`,
			},
			data: "team calendar event",
			want: true,
		},
		{
			name: "combined conditions reject partial match",
			rule: StringMatchRule{
				Contains:   "calendar",
				Prefix:     "team",
				Suffix:     "shift",
				RegexMatch: `team .* event`,
			},
			data: "team calendar event",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rule.matchesString(tt.data)
			if got != tt.want {
				t.Fatalf("matchesString(%q) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestStringTransformRuleHasActions(t *testing.T) {
	tests := []struct {
		name string
		rule StringTransformRule
		want bool
	}{
		{
			name: "empty rule",
			rule: StringTransformRule{},
			want: false,
		},
		{
			name: "replace action",
			rule: StringTransformRule{Replace: "new"},
			want: true,
		},
		{
			name: "remove action",
			rule: StringTransformRule{Remove: true},
			want: true,
		},
		{
			name: "prefix action",
			rule: StringTransformRule{Prefix: "new "},
			want: true,
		},
		{
			name: "suffix action",
			rule: StringTransformRule{Suffix: " new"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rule.hasActions()
			if got != tt.want {
				t.Fatalf("hasActions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyStringTransform(t *testing.T) {
	tests := []struct {
		name  string
		value string
		rule  StringTransformRule
		want  string
	}{
		{
			name:  "empty rule keeps value",
			value: "calendar event",
			rule:  StringTransformRule{},
			want:  "calendar event",
		},
		{
			name:  "remove wins",
			value: "calendar event",
			rule:  StringTransformRule{Remove: true, Replace: "replacement", Prefix: "[", Suffix: "]"},
			want:  "",
		},
		{
			name:  "replace wins over prefix and suffix",
			value: "calendar event",
			rule:  StringTransformRule{Replace: "replacement", Prefix: "[", Suffix: "]"},
			want:  "replacement",
		},
		{
			name:  "prefix and suffix apply together",
			value: "calendar event",
			rule:  StringTransformRule{Prefix: "[", Suffix: "]"},
			want:  "[calendar event]",
		},
		{
			name:  "prefix applies to empty value",
			value: "",
			rule:  StringTransformRule{Prefix: "prefix"},
			want:  "prefix",
		},
		{
			name:  "suffix applies to empty value",
			value: "",
			rule:  StringTransformRule{Suffix: "suffix"},
			want:  "suffix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyStringTransform(tt.value, tt.rule)
			if got != tt.want {
				t.Fatalf("applyStringTransform() = %q, want %q", got, tt.want)
			}
		})
	}
}
