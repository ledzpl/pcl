package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigValidateSuccess(t *testing.T) {
	cfg := Config{
		OpenAIAPIKey: "openai",
		JiraAPIKey:   "jira",
		JiraHost:     "https://example.atlassian.net",
		JiraEmail:    "user@example.com",
		JiraProject:  "TEST",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
}

func TestConfigValidateMissingFields(t *testing.T) {
	cfg := Config{
		JiraAPIKey:  "jira",
		JiraProject: "TEST",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("Validate() expected error, got nil")
	}

	errMsg := err.Error()
	wantSubstrings := []string{
		"openai_api_key",
		"jira_host",
		"jira_email",
	}

	for _, substring := range wantSubstrings {
		if !strings.Contains(errMsg, substring) {
			t.Fatalf("Validate() error missing %q in %q", substring, errMsg)
		}
	}
}

func TestLoadSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	data := []byte(`{
		"openai_api_key": "openai",
		"jira_api_key": "jira",
		"jira_host": "https://example.atlassian.net",
		"jira_email": "user@example.com",
		"jira_project": "TEST"
	}`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile() failed: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.OpenAIAPIKey != "openai" || cfg.JiraProject != "TEST" {
		t.Fatalf("Load() returned unexpected config: %+v", cfg)
	}
}

func TestLoadValidationError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	data := []byte(`{
		"openai_api_key": "",
		"jira_api_key": "jira"
	}`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile() failed: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatalf("Load() expected error, got nil")
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load("does-not-exist.json"); err == nil {
		t.Fatalf("Load() expected error for missing file, got nil")
	}
}

func TestConfigValidateWhitespace(t *testing.T) {
	cfg := Config{
		OpenAIAPIKey: "   ",
		JiraAPIKey:   "jira",
		JiraHost:     " ",
		JiraEmail:    "user@example.com",
		JiraProject:  "",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("Validate() expected error for whitespace fields")
	}

	errMsg := err.Error()
	for _, key := range []string{"openai_api_key", "jira_host", "jira_project"} {
		if !strings.Contains(errMsg, key) {
			t.Fatalf("Validate() missing key %q in error: %s", key, errMsg)
		}
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := os.WriteFile(path, []byte("not-json"), 0o600); err != nil {
		t.Fatalf("WriteFile() failed: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("Load() expected parse error, got nil")
	}

	if !strings.Contains(err.Error(), "parse") {
		t.Fatalf("Load() error %q did not mention parse", err)
	}
}
