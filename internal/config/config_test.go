package config

import "testing"

func TestValidateForAI(t *testing.T) {
	t.Parallel()

	cfg := Config{OpenAIAPIKey: "token"}
	if err := cfg.ValidateForAI(); err != nil {
		t.Fatalf("ValidateForAI() unexpected error: %v", err)
	}

	cfg = Config{}
	if err := cfg.ValidateForAI(); err == nil {
		t.Fatal("ValidateForAI() expected error when openai_api_key is empty, got nil")
	}
}

func TestValidateForJira(t *testing.T) {
	t.Parallel()

	valid := Config{
		OpenAIAPIKey: "token",
		JiraAPIKey:   "jira-token",
		JiraHost:     "https://example.atlassian.net",
		JiraEmail:    "dev@example.com",
		JiraProject:  "PCL",
	}

	if err := valid.ValidateForJira(); err != nil {
		t.Fatalf("ValidateForJira() unexpected error: %v", err)
	}

	missing := Config{
		OpenAIAPIKey: "token",
		JiraAPIKey:   "jira-token",
		JiraHost:     "https://example.atlassian.net",
		JiraProject:  "PCL",
	}

	if err := missing.ValidateForJira(); err == nil {
		t.Fatal("ValidateForJira() expected error when required fields are missing, got nil")
	}
}
