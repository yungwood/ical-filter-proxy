package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want bool
	}{
		{
			name: "valid public calendar",
			yaml: `
calendars:
  - name: public
    public: true
    feed_url: https://example.com/feed.ics
`,
			want: true,
		},
		{
			name: "valid token calendar",
			yaml: `
calendars:
  - name: private
    token: secret
    feed_url: https://example.com/feed.ics
`,
			want: true,
		},
		{
			name: "empty config",
			yaml: `{}`,
			want: false,
		},
		{
			name: "invalid feed url",
			yaml: `
calendars:
  - name: invalid-url
    public: true
    feed_url: ftp://example.com/feed.ics
`,
			want: false,
		},
		{
			name: "tokenless non-public calendar",
			yaml: `
calendars:
  - name: private
    feed_url: https://example.com/feed.ics
`,
			want: false,
		},
		{
			name: "public calendar with token",
			yaml: `
calendars:
  - name: public
    public: true
    token: secret
    feed_url: https://example.com/feed.ics
`,
			want: false,
		},
		{
			name: "public calendar with token_file",
			yaml: `
calendars:
  - name: public
    public: true
    token_file: /run/secrets/token
    feed_url: https://example.com/feed.ics
`,
			want: false,
		},
		{
			name: "invalid yaml",
			yaml: `
calendars:
  - name: invalid
    public: true
    feed_url: [
`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile := writeTempFile(t, "config-*.yaml", tt.yaml)

			var config Config
			got := config.LoadConfig(configFile)
			if got != tt.want {
				t.Fatalf("LoadConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	var config Config
	got := config.LoadConfig(filepath.Join(t.TempDir(), "missing.yaml"))
	if got {
		t.Fatal("LoadConfig() = true, want false")
	}
}

func TestLoadConfigLoadsSecretsFromFiles(t *testing.T) {
	tokenFile := writeTempFile(t, "token-*", "file-token\n")
	feedURLFile := writeTempFile(t, "feed-url-*", "https://example.com/from-file.ics\n")
	configFile := writeTempFile(t, "config-*.yaml", `
calendars:
  - name: private
    token: ignored-token
    token_file: `+tokenFile+`
    feed_url: https://example.com/ignored.ics
    feed_url_file: `+feedURLFile+`
`)

	var config Config
	if !config.LoadConfig(configFile) {
		t.Fatal("LoadConfig() = false, want true")
	}

	if got := config.Calendars[0].Token; got != "file-token" {
		t.Fatalf("Token = %q, want file-token", got)
	}
	if got := config.Calendars[0].FeedURL; got != "https://example.com/from-file.ics" {
		t.Fatalf("FeedURL = %q, want https://example.com/from-file.ics", got)
	}
}

func TestConfigRuntimeConfig(t *testing.T) {
	config := Config{
		Calendars: []CalendarConfig{
			{
				Name:        "private",
				PublishName: "Published Calendar",
				Public:      false,
				Token:       "secret",
				TokenFile:   "/run/secrets/token",
				FeedURL:     "https://example.com/feed.ics",
				FeedURLFile: "/run/secrets/feed-url",
				Filters: []Filter{
					{
						Description: "rename event",
						Transform: EventTransformRules{
							Summary: StringTransformRule{Replace: "renamed"},
						},
					},
				},
			},
		},
	}

	runtimeConfig, err := config.RuntimeConfig()
	if err != nil {
		t.Fatalf("RuntimeConfig() returned error: %v", err)
	}

	if len(runtimeConfig.Calendars) != 1 {
		t.Fatalf("len(Calendars) = %d, want 1", len(runtimeConfig.Calendars))
	}

	calendar := runtimeConfig.Calendars[0]
	if calendar.Name != "private" {
		t.Fatalf("Name = %q, want private", calendar.Name)
	}
	if calendar.PublishName != "Published Calendar" {
		t.Fatalf("PublishName = %q, want Published Calendar", calendar.PublishName)
	}
	if calendar.Public {
		t.Fatal("Public = true, want false")
	}
	if calendar.Token != "secret" {
		t.Fatalf("Token = %q, want secret", calendar.Token)
	}
	if calendar.FeedURL != "https://example.com/feed.ics" {
		t.Fatalf("FeedURL = %q, want https://example.com/feed.ics", calendar.FeedURL)
	}
	if len(calendar.Filters) != 1 {
		t.Fatalf("len(Filters) = %d, want 1", len(calendar.Filters))
	}
}

func TestConfigRuntimeConfigRejectsInvalidRegex(t *testing.T) {
	config := Config{
		Calendars: []CalendarConfig{
			{
				Name:    "public",
				Public:  true,
				FeedURL: "https://example.com/feed.ics",
				Filters: []Filter{
					{
						Description: "invalid regex",
						Match: EventMatchRules{
							Summary: StringMatchRule{RegexMatch: "["},
						},
					},
				},
			},
		},
	}

	_, err := config.RuntimeConfig()
	if err == nil {
		t.Fatal("RuntimeConfig() returned nil error")
	}
}

func TestLoadConfigRejectsMissingSecretFiles(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "missing feed_url_file",
			yaml: `
calendars:
  - name: private
    token: secret
    feed_url_file: /missing/feed-url
`,
		},
		{
			name: "missing token_file without fallback token",
			yaml: `
calendars:
  - name: private
    token_file: /missing/token
    feed_url: https://example.com/feed.ics
`,
		},
		{
			name: "missing token_file with fallback token",
			yaml: `
calendars:
  - name: private
    token: ignored-token
    token_file: /missing/token
    feed_url: https://example.com/feed.ics
`,
		},
		{
			name: "missing token_file on public calendar",
			yaml: `
calendars:
  - name: public
    public: true
    token_file: /missing/token
    feed_url: https://example.com/feed.ics
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile := writeTempFile(t, "config-*.yaml", tt.yaml)

			var config Config
			got := config.LoadConfig(configFile)
			if got {
				t.Fatal("LoadConfig() = true, want false")
			}
		})
	}
}

func TestReadSecretFile(t *testing.T) {
	secretFile := writeTempFile(t, "secret-*", " secret value \n")

	got, err := readSecretFile(secretFile)
	if err != nil {
		t.Fatalf("readSecretFile() returned error: %v", err)
	}
	if got != "secret value" {
		t.Fatalf("readSecretFile() = %q, want secret value", got)
	}
}

func TestReadSecretFileMissingFile(t *testing.T) {
	_, err := readSecretFile(filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("readSecretFile() returned nil error")
	}
}

func writeTempFile(t *testing.T, pattern string, data string) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), pattern)
	if err != nil {
		t.Fatalf("CreateTemp() returned error: %v", err)
	}
	if _, err := file.WriteString(data); err != nil {
		t.Fatalf("WriteString() returned error: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}

	return file.Name()
}
