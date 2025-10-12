package jira

import (
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	email := "user@example.com"
	token := "secret"

	auth := basicAuth(email, token)

	expected := "Basic " + base64.StdEncoding.EncodeToString([]byte(email+":"+token))
	if auth != expected {
		t.Fatalf("basicAuth() = %q, want %q", auth, expected)
	}
}

func TestCreateIssue(t *testing.T) {
	reqBody := `{"fields":{}}`

	var received struct {
		method string
		path   string
		body   string
		header http.Header
	}

	ts := newIPv4Server(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received.method = r.Method
		received.path = r.URL.Path
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		received.body = string(payload)
		received.header = r.Header.Clone()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"1000"}`))
	}))
	defer ts.Close()

	email := "user@example.com"
	token := "token123"

	if err := CreateIssue(reqBody, email, ts.URL, token); err != nil {
		t.Fatalf("CreateIssue() unexpected error: %v", err)
	}

	if received.method != http.MethodPost {
		t.Fatalf("CreateIssue did not POST: got %s", received.method)
	}

	if received.path != "/rest/api/3/issue" {
		t.Fatalf("CreateIssue path = %s, want /rest/api/3/issue", received.path)
	}

	if received.body != reqBody {
		t.Fatalf("CreateIssue body = %q, want %q", received.body, reqBody)
	}

	expectedAuth := basicAuth(email, token)
	if got := received.header.Get("Authorization"); got != expectedAuth {
		t.Fatalf("Authorization header = %q, want %q", got, expectedAuth)
	}

	if got := received.header.Get("Content-type"); got != "application/json" {
		t.Fatalf("Content-type header = %q, want application/json", got)
	}

	if got := received.header.Get("Accept"); got != "application/json" {
		t.Fatalf("Accept header = %q, want application/json", got)
	}
}

func TestCreateIssueReturnsErrorOnHTTPFailure(t *testing.T) {
	ts := newIPv4Server(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid payload"}`))
	}))
	defer ts.Close()

	email := "user@example.com"
	token := "token123"

	err := CreateIssue(`{"fields":{}}`, email, ts.URL, token)
	if err == nil {
		t.Fatalf("CreateIssue() expected error for HTTP 400")
	}
	if !strings.Contains(err.Error(), "status 400") {
		t.Fatalf("CreateIssue() error %q missing status code", err)
	}
}

func TestGetAccountId(t *testing.T) {
	const accountID = "abc-123"

	var received http.Header

	ts := newIPv4Server(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()

		if r.URL.Path != "/rest/api/3/myself" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"accountId":"` + accountID + `"}`))
	}))
	defer ts.Close()

	email := "user@example.com"
	token := "token123"

	got, err := GetAccountId(email, ts.URL, token)
	if err != nil {
		t.Fatalf("GetAccountId() unexpected error: %v", err)
	}
	if got != accountID {
		t.Fatalf("GetAccountId() = %q, want %q", got, accountID)
	}

	expectedAuth := basicAuth(email, token)
	if received.Get("Authorization") != expectedAuth {
		t.Fatalf("Authorization header = %q, want %q", received.Get("Authorization"), expectedAuth)
	}

	if received.Get("Accept") != "application/json" {
		t.Fatalf("Accept header = %q, want application/json", received.Get("Accept"))
	}
}

func TestGetAccountIdReturnsErrorOnHTTPFailure(t *testing.T) {
	ts := newIPv4Server(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer ts.Close()

	email := "user@example.com"
	token := "token123"

	_, err := GetAccountId(email, ts.URL, token)
	if err == nil {
		t.Fatalf("GetAccountId() expected error for HTTP 401")
	}
	if !strings.Contains(err.Error(), "status 401") {
		t.Fatalf("GetAccountId() error %q missing status code", err)
	}
}

func TestGetAccountIdMissingField(t *testing.T) {
	ts := newIPv4Server(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	email := "user@example.com"
	token := "token123"

	_, err := GetAccountId(email, ts.URL, token)
	if err == nil {
		t.Fatalf("GetAccountId() expected error when accountId missing")
	}
	if !strings.Contains(err.Error(), "missing accountId") {
		t.Fatalf("GetAccountId() error %q missing expected message", err)
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
