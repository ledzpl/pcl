package jira

import (
	"encoding/base64"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

func CreateIssue(request string, email string, host string, token string) {
	c := resty.New().
		SetBaseURL(host).
		SetTimeout(8*time.Second).
		SetHeader("Accept", "application/json").
		SetHeader("Content-type", "application/json").
		SetHeader("Authorization", basicAuth(email, token))

	var postOut map[string]any
	_, err := c.R().
		SetBody(request).
		SetResult(&postOut).
		Post("/rest/api/3/issue")

	if err != nil {
		log.Fatalln(err.Error())
	}
}

func GetAccountId(email, host, token string) string {
	c := resty.New().
		SetBaseURL(host).
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", basicAuth(email, token))

	var getOut map[string]any
	_, err := c.R().
		SetResult(&getOut).
		Get("/rest/api/3/myself")

	if err != nil {
		log.Fatal("failed to get accountid")
	}

	return getOut["accountId"].(string)
}

func basicAuth(email, token string) string {
	creds := email + ":" + token
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(creds))
}
