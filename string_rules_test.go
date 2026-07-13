package main

import "testing"

func TestStringMatchRuleConfigHasConditions(t *testing.T) {
	tests := []struct {
		name string
		rule StringMatchRuleConfig
		want bool
	}{
		{
			name: "empty rule",
			rule: StringMatchRuleConfig{},
			want: false,
		},
		{
			name: "empty condition",
			rule: StringMatchRuleConfig{Null: true},
			want: true,
		},
		{
			name: "contains condition",
			rule: StringMatchRuleConfig{Contains: "event"},
			want: true,
		},
		{
			name: "contains_any condition",
			rule: StringMatchRuleConfig{ContainsAny: []string{"event"}},
			want: true,
		},
		{
			name: "contains_all condition",
			rule: StringMatchRuleConfig{ContainsAll: []string{"event"}},
			want: true,
		},
		{
			name: "prefix condition",
			rule: StringMatchRuleConfig{Prefix: "event"},
			want: true,
		},
		{
			name: "suffix condition",
			rule: StringMatchRuleConfig{Suffix: "event"},
			want: true,
		},
		{
			name: "regex condition",
			rule: StringMatchRuleConfig{RegexMatch: "event"},
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

func TestStringMatchRuleConfigCompile(t *testing.T) {
	rule, err := (StringMatchRuleConfig{
		Contains:   "calendar",
		Prefix:     "team",
		Suffix:     "event",
		RegexMatch: `team .* event`,
	}).compile()
	if err != nil {
		t.Fatalf("compile() returned error: %v", err)
	}

	if !rule.matchesString("team calendar event") {
		t.Fatal("compiled rule did not match expected value")
	}
	if rule.matchesString("team calendar shift") {
		t.Fatal("compiled rule matched unexpected value")
	}
}

func TestStringMatchRuleConfigCompileCopiesContainsSlices(t *testing.T) {
	config := StringMatchRuleConfig{
		ContainsAny: []string{"calendar"},
		ContainsAll: []string{"team", "event"},
	}

	rule, err := config.compile()
	if err != nil {
		t.Fatalf("compile() returned error: %v", err)
	}

	config.ContainsAny[0] = "changed"
	config.ContainsAll[0] = "changed"

	if !rule.matchesString("team calendar event") {
		t.Fatal("compiled rule changed after config slice mutation")
	}
}

func TestStringMatchRuleConfigCompileInvalidRegex(t *testing.T) {
	_, err := (StringMatchRuleConfig{RegexMatch: `[`}).compile()
	if err == nil {
		t.Fatal("compile() returned nil error")
	}
}

func TestStringMatchRuleConfigCompileRejectsEmptyContainsListValues(t *testing.T) {
	tests := []struct {
		name    string
		rule    StringMatchRuleConfig
		wantErr string
	}{
		{
			name:    "contains_any empty value",
			rule:    StringMatchRuleConfig{ContainsAny: []string{"calendar", ""}},
			wantErr: "contains_any[1] must not be empty",
		},
		{
			name:    "contains_all empty value",
			rule:    StringMatchRuleConfig{ContainsAll: []string{"", "event"}},
			wantErr: "contains_all[0] must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.rule.compile()
			if err == nil {
				t.Fatal("compile() returned nil error")
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestStringMatchRuleConfigMatchesString(t *testing.T) {
	tests := []struct {
		name string
		rule StringMatchRuleConfig
		data string
		want bool
	}{
		{
			name: "empty rule matches non-empty string",
			rule: StringMatchRuleConfig{},
			data: "calendar event",
			want: true,
		},
		{
			name: "empty rule matches empty string",
			rule: StringMatchRuleConfig{},
			data: "",
			want: true,
		},
		{
			name: "empty condition matches empty string",
			rule: StringMatchRuleConfig{Null: true},
			data: "",
			want: true,
		},
		{
			name: "empty condition rejects non-empty string",
			rule: StringMatchRuleConfig{Null: true},
			data: "calendar event",
			want: false,
		},
		{
			name: "contains matches",
			rule: StringMatchRuleConfig{Contains: "event"},
			data: "calendar event",
			want: true,
		},
		{
			name: "contains rejects missing value",
			rule: StringMatchRuleConfig{Contains: "shift"},
			data: "calendar event",
			want: false,
		},
		{
			name: "contains rejects empty data",
			rule: StringMatchRuleConfig{Contains: "event"},
			data: "",
			want: false,
		},
		{
			name: "contains_any matches first value",
			rule: StringMatchRuleConfig{ContainsAny: []string{"calendar", "holiday"}},
			data: "calendar event",
			want: true,
		},
		{
			name: "contains_any matches later value",
			rule: StringMatchRuleConfig{ContainsAny: []string{"holiday", "event"}},
			data: "calendar event",
			want: true,
		},
		{
			name: "contains_any rejects missing values",
			rule: StringMatchRuleConfig{ContainsAny: []string{"shift", "holiday"}},
			data: "calendar event",
			want: false,
		},
		{
			name: "contains_any rejects empty data",
			rule: StringMatchRuleConfig{ContainsAny: []string{"event"}},
			data: "",
			want: false,
		},
		{
			name: "empty contains_any list is ignored",
			rule: StringMatchRuleConfig{ContainsAny: []string{}},
			data: "calendar event",
			want: true,
		},
		{
			name: "contains_all matches all values",
			rule: StringMatchRuleConfig{ContainsAll: []string{"calendar", "event"}},
			data: "calendar event",
			want: true,
		},
		{
			name: "contains_all rejects missing value",
			rule: StringMatchRuleConfig{ContainsAll: []string{"calendar", "shift"}},
			data: "calendar event",
			want: false,
		},
		{
			name: "contains_all rejects empty data",
			rule: StringMatchRuleConfig{ContainsAll: []string{"event"}},
			data: "",
			want: false,
		},
		{
			name: "empty contains_all list is ignored",
			rule: StringMatchRuleConfig{ContainsAll: []string{}},
			data: "calendar event",
			want: true,
		},
		{
			name: "prefix matches",
			rule: StringMatchRuleConfig{Prefix: "calendar"},
			data: "calendar event",
			want: true,
		},
		{
			name: "prefix rejects missing prefix",
			rule: StringMatchRuleConfig{Prefix: "event"},
			data: "calendar event",
			want: false,
		},
		{
			name: "suffix matches",
			rule: StringMatchRuleConfig{Suffix: "event"},
			data: "calendar event",
			want: true,
		},
		{
			name: "suffix rejects missing suffix",
			rule: StringMatchRuleConfig{Suffix: "calendar"},
			data: "calendar event",
			want: false,
		},
		{
			name: "regex matches",
			rule: StringMatchRuleConfig{RegexMatch: `cal.*event`},
			data: "calendar event",
			want: true,
		},
		{
			name: "regex rejects missing match",
			rule: StringMatchRuleConfig{RegexMatch: `^event`},
			data: "calendar event",
			want: false,
		},
		{
			name: "invalid regex rejects",
			rule: StringMatchRuleConfig{RegexMatch: `[`},
			data: "calendar event",
			want: false,
		},
		{
			name: "contains_any with empty value rejects",
			rule: StringMatchRuleConfig{ContainsAny: []string{""}},
			data: "calendar event",
			want: false,
		},
		{
			name: "contains_all with empty value rejects",
			rule: StringMatchRuleConfig{ContainsAll: []string{""}},
			data: "calendar event",
			want: false,
		},
		{
			name: "combined conditions all match",
			rule: StringMatchRuleConfig{
				Contains:    "calendar",
				ContainsAny: []string{"meeting", "event"},
				ContainsAll: []string{"team", "calendar"},
				Prefix:      "team",
				Suffix:      "event",
				RegexMatch:  `team .* event`,
			},
			data: "team calendar event",
			want: true,
		},
		{
			name: "combined conditions reject partial match",
			rule: StringMatchRuleConfig{
				Contains:    "calendar",
				ContainsAny: []string{"meeting", "event"},
				ContainsAll: []string{"team", "calendar"},
				Prefix:      "team",
				Suffix:      "shift",
				RegexMatch:  `team .* event`,
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
