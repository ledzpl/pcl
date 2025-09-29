package aitool

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnalysisSendsExpectedPayload(t *testing.T) {
	diff := "diff --git a/file.txt b/file.txt\n+added line\n"
	accountID := "account-123"
	projectID := "PROJ"
	apiKey := "test-key"
	expectedContent := "{\"fields\":{\"summary\":\"Generated\"}}"

	type messageCapture struct {
		Role    string
		Content string
	}
	type capturedRequest struct {
		Method        string
		Path          string
		Authorization string
		Model         string
		Seed          int64
		Messages      []messageCapture
		Err           error
	}

	reqCh := make(chan capturedRequest, 1)

	ts := newIPv4Server(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		capture := capturedRequest{
			Method:        r.Method,
			Path:          r.URL.Path,
			Authorization: r.Header.Get("Authorization"),
		}

		var body struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string          `json:"role"`
				Content json.RawMessage `json:"content"`
			} `json:"messages"`
			Seed int64 `json:"seed"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			capture.Err = err
		} else {
			capture.Model = body.Model
			capture.Seed = body.Seed
			capture.Messages = make([]messageCapture, 0, len(body.Messages))
			for _, msg := range body.Messages {
				var content string
				if err := json.Unmarshal(msg.Content, &content); err != nil {
					capture.Err = err
					break
				}
				capture.Messages = append(capture.Messages, messageCapture{
					Role:    msg.Role,
					Content: content,
				})
			}
		}

		response := map[string]any{
			"id":      "chatcmpl-test",
			"object":  "chat.completion",
			"created": 0,
			"model":   "gpt-5",
			"choices": []any{
				map[string]any{
					"index":         0,
					"finish_reason": "stop",
					"message": map[string]any{
						"role":    "assistant",
						"content": expectedContent,
						"refusal": "",
					},
					"logprobs": map[string]any{
						"content": []any{},
						"refusal": []any{},
					},
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     1,
				"completion_tokens": 1,
				"total_tokens":      2,
				"completion_tokens_details": map[string]any{
					"accepted_prediction_tokens": 0,
					"audio_tokens":               0,
					"reasoning_tokens":           0,
					"rejected_prediction_tokens": 0,
				},
				"prompt_tokens_details": map[string]any{
					"audio_tokens":  0,
					"cached_tokens": 0,
				},
			},
		}

		reqCh <- capture

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			panic(err)
		}
	}))
	t.Cleanup(ts.Close)

	t.Setenv("OPENAI_BASE_URL", ts.URL+"/")

	got := Analysis(diff, accountID, projectID, apiKey)

	capture := <-reqCh

	if capture.Err != nil {
		t.Fatalf("decode request: %v", capture.Err)
	}
	if capture.Method != http.MethodPost {
		t.Fatalf("unexpected method: %s", capture.Method)
	}
	if capture.Path != "/chat/completions" {
		t.Fatalf("unexpected path: %s", capture.Path)
	}
	expectedAuth := fmt.Sprintf("Bearer %s", apiKey)
	if capture.Authorization != expectedAuth {
		t.Fatalf("authorization header = %q, want %q", capture.Authorization, expectedAuth)
	}
	if capture.Model != "gpt-5" {
		t.Fatalf("model = %q, want %q", capture.Model, "gpt-5")
	}
	if capture.Seed != 42 {
		t.Fatalf("seed = %d, want 42", capture.Seed)
	}

	if len(capture.Messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(capture.Messages))
	}

	if capture.Messages[0].Role != "system" || capture.Messages[0].Content != SYSPROMPT {
		t.Fatalf("system message mismatch")
	}

	expectedPrompt := fmt.Sprintf(PROMPT, projectID, accountID)
	if capture.Messages[1].Role != "user" || capture.Messages[1].Content != expectedPrompt {
		t.Fatalf("prompt message mismatch")
	}

	if capture.Messages[2].Role != "user" || capture.Messages[2].Content != diff {
		t.Fatalf("diff message mismatch")
	}

	if got != expectedContent {
		t.Fatalf("Analysis() returned %q, want %q", got, expectedContent)
	}
}

func newIPv4Server(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on IPv4: %v", err)
	}
	server := httptest.NewUnstartedServer(handler)
	server.Listener = listener
	server.Start()
	return server
}
