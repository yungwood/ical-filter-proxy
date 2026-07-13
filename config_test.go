package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "valid public calendar",
			yaml: `
calendars:
  - name: public
    public: true
    feed_url: https://example.com/feed.ics
`,
			wantErr: false,
		},
		{
			name: "valid token calendar",
			yaml: `
calendars:
  - name: private
    token: secret
    feed_url: https://example.com/feed.ics
`,
			wantErr: false,
		},
		{
			name:    "empty config",
			yaml:    `{}`,
			wantErr: true,
		},
		{
			name: "invalid feed url",
			yaml: `
calendars:
  - name: invalid-url
    public: true
    feed_url: ftp://example.com/feed.ics
`,
			wantErr: true,
		},
		{
			name: "tokenless non-public calendar",
			yaml: `
calendars:
  - name: private
    feed_url: https://example.com/feed.ics
`,
			wantErr: true,
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
			wantErr: true,
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
			wantErr: true,
		},
		{
			name: "invalid yaml",
			yaml: `
calendars:
  - name: invalid
    public: true
    feed_url: [
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile := writeTempFile(t, "config-*.yaml", tt.yaml)

			_, err := LoadConfig(configFile)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := LoadConfig(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil {
		t.Fatal("LoadConfig() returned nil error")
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

	config, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}

	if got := config.Calendars[0].Token; got != "file-token" {
		t.Fatalf("Token = %q, want file-token", got)
	}
	if got := config.Calendars[0].FeedURL; got != "https://example.com/from-file.ics" {
		t.Fatalf("FeedURL = %q, want https://example.com/from-file.ics", got)
	}
}

func TestLoadConfigResolvesEnvironmentReferences(t *testing.T) {
	configFile := writeTempFile(t, "config-*.yaml", `
calendars:
  - name: private
    token: ${CALENDAR_TOKEN}
    feed_url: ${CALENDAR_FEED_URL}
`)

	config, err := loadConfig(configFile, mapEnv(map[string]string{
		"CALENDAR_TOKEN":    "env-token",
		"CALENDAR_FEED_URL": "https://example.com/from-env.ics",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}

	if got := config.Calendars[0].Token; got != "env-token" {
		t.Fatalf("Token = %q, want env-token", got)
	}
	if got := config.Calendars[0].FeedURL; got != "https://example.com/from-env.ics" {
		t.Fatalf("FeedURL = %q, want https://example.com/from-env.ics", got)
	}
}

func TestLoadConfigRejectsMissingEnvironmentReference(t *testing.T) {
	configFile := writeTempFile(t, "config-*.yaml", `
calendars:
  - name: private
    token: ${MISSING_TOKEN}
    feed_url: https://example.com/feed.ics
`)

	_, err := loadConfig(configFile, emptyEnv)
	if err == nil {
		t.Fatal("LoadConfig() returned nil error")
	}
	if !strings.Contains(err.Error(), "MISSING_TOKEN") {
		t.Fatalf("LoadConfig() error = %q, want missing env name", err)
	}
}

func TestLoadConfigRejectsEmptyEnvironmentReference(t *testing.T) {
	configFile := writeTempFile(t, "config-*.yaml", `
calendars:
  - name: private
    token: ${EMPTY_TOKEN}
    feed_url: https://example.com/feed.ics
`)

	_, err := loadConfig(configFile, mapEnv(map[string]string{
		"EMPTY_TOKEN": "",
	}))
	if err == nil {
		t.Fatal("LoadConfig() returned nil error")
	}
	if !strings.Contains(err.Error(), "EMPTY_TOKEN") {
		t.Fatalf("LoadConfig() error = %q, want empty env name", err)
	}
}

func TestLoadConfigDoesNotExpandPartialEnvironmentReferences(t *testing.T) {
	configFile := writeTempFile(t, "config-*.yaml", `
calendars:
  - name: private
    token: ${TOKEN}
    feed_url: https://example.com/${CALENDAR_TOKEN}/feed.ics
`)

	config, err := loadConfig(configFile, mapEnv(map[string]string{
		"TOKEN":          "env-token",
		"CALENDAR_TOKEN": "secret-path",
	}))
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}

	if got := config.Calendars[0].FeedURL; got != "https://example.com/${CALENDAR_TOKEN}/feed.ics" {
		t.Fatalf("FeedURL = %q, want literal partial env reference", got)
	}
}

func TestLoadConfigSecretFilesTakePrecedenceOverEnvironmentReferences(t *testing.T) {
	tokenFile := writeTempFile(t, "token-*", "file-token\n")
	feedURLFile := writeTempFile(t, "feed-url-*", "https://example.com/from-file.ics\n")
	configFile := writeTempFile(t, "config-*.yaml", `
calendars:
  - name: private
    token: ${CALENDAR_TOKEN}
    token_file: `+tokenFile+`
    feed_url: ${CALENDAR_FEED_URL}
    feed_url_file: `+feedURLFile+`
`)

	config, err := loadConfig(configFile, emptyEnv)
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}

	if got := config.Calendars[0].Token; got != "file-token" {
		t.Fatalf("Token = %q, want file-token", got)
	}
	if got := config.Calendars[0].FeedURL; got != "https://example.com/from-file.ics" {
		t.Fatalf("FeedURL = %q, want https://example.com/from-file.ics", got)
	}
}

func TestConfigCompile(t *testing.T) {
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
				Filters: []FilterConfig{
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

	runtimeConfig, err := config.Compile()
	if err != nil {
		t.Fatalf("Compile() returned error: %v", err)
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

func TestConfigCompileRejectsInvalidRegex(t *testing.T) {
	config := Config{
		Calendars: []CalendarConfig{
			{
				Name:    "public",
				Public:  true,
				FeedURL: "https://example.com/feed.ics",
				Filters: []FilterConfig{
					{
						Description: "invalid regex",
						Match: EventMatchRulesConfig{
							Summary: StringMatchRuleConfig{RegexMatch: "["},
						},
					},
				},
			},
		},
	}

	_, err := config.Compile()
	if err == nil {
		t.Fatal("Compile() returned nil error")
	}
	if !strings.Contains(err.Error(), `calendar "public" filter 0`) || !strings.Contains(err.Error(), `summary: invalid regex`) {
		t.Fatalf("Compile() error = %q, want calendar/filter/property context", err)
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

			_, err := LoadConfig(configFile)
			if err == nil {
				t.Fatal("LoadConfig() returned nil error")
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
