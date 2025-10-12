package jira

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

func CreateIssue(request string, email string, host string, token string) error {
	c := resty.New().
		SetBaseURL(host).
		SetTimeout(8*time.Second).
		SetHeader("Accept", "application/json").
		SetHeader("Content-type", "application/json").
		SetHeader("Authorization", basicAuth(email, token))

	resp, err := c.R().
		SetBody(request).
		Post("/rest/api/3/issue")

	if err != nil {
		return fmt.Errorf("jira: create issue request failed: %w", err)
	}

	if resp.IsError() {
		body := strings.TrimSpace(string(resp.Body()))
		if body == "" {
			body = resp.Status()
		}
		return fmt.Errorf("jira: create issue failed: status %d: %s", resp.StatusCode(), body)
	}

	return nil
}

func GetAccountId(email, host, token string) (string, error) {
	c := resty.New().
		SetBaseURL(host).
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", basicAuth(email, token))

	var getOut map[string]any
	resp, err := c.R().
		SetResult(&getOut).
		Get("/rest/api/3/myself")

	if err != nil {
		return "", fmt.Errorf("jira: get account id request failed: %w", err)
	}

	if resp.IsError() {
		body := strings.TrimSpace(string(resp.Body()))
		if body == "" {
			body = resp.Status()
		}
		return "", fmt.Errorf("jira: get account id failed: status %d: %s", resp.StatusCode(), body)
	}

	accountID, ok := getOut["accountId"].(string)
	if !ok || strings.TrimSpace(accountID) == "" {
		return "", fmt.Errorf("jira: get account id failed: missing accountId field")
	}

	return accountID, nil
}

func basicAuth(email, token string) string {
	creds := email + ":" + token
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(creds))
}
