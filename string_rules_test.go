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

func TestStringMatchRuleConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		rule    StringMatchRuleConfig
		wantErr bool
	}{
		{
			name:    "empty rule is valid",
			rule:    StringMatchRuleConfig{},
			wantErr: false,
		},
		{
			name:    "valid regex is valid",
			rule:    StringMatchRuleConfig{RegexMatch: `cal.*event`},
			wantErr: false,
		},
		{
			name:    "invalid regex is invalid",
			rule:    StringMatchRuleConfig{RegexMatch: `[`},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("validate() error = %v, wantErr %v", err, tt.wantErr)
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

func TestStringMatchRuleConfigCompileInvalidRegex(t *testing.T) {
	_, err := (StringMatchRuleConfig{RegexMatch: `[`}).compile()
	if err == nil {
		t.Fatal("compile() returned nil error")
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
			name: "combined conditions all match",
			rule: StringMatchRuleConfig{
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
			rule: StringMatchRuleConfig{
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
