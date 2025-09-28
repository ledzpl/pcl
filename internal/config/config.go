package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	OpenAIAPIKey string `json:"openai_api_key"`
	JiraAPIKey   string `json:"jira_api_key"`
	JiraHost     string `json:"jira_host"`
	JiraEmail    string `json:"jira_email"`
	JiraProject  string `json:"jira_project"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %q: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c Config) Validate() error {
	missing := make([]string, 0, 5)

	if isBlank(c.OpenAIAPIKey) {
		missing = append(missing, "openai_api_key")
	}
	if isBlank(c.JiraAPIKey) {
		missing = append(missing, "jira_api_key")
	}
	if isBlank(c.JiraHost) {
		missing = append(missing, "jira_host")
	}
	if isBlank(c.JiraEmail) {
		missing = append(missing, "jira_email")
	}
	if isBlank(c.JiraProject) {
		missing = append(missing, "jira_project")
	}

	if len(missing) > 0 {
		return fmt.Errorf("config: missing required keys: %s", strings.Join(missing, ", "))
	}

	return nil
}

func isBlank(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
