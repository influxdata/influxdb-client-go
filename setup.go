package influxdb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

// Setup sets up a new influxdb server.
// It requires a client be set up with a username and password.
// If successful will add a token to the client.
// RetentionPeriodHrs of zero will result in infinite retention.
func (c *Client) Setup(ctx context.Context, bucket, org string, retentionPeriodHrs int) (*SetupResult, error) {
	if c.username == "" || c.password == "" {
		return nil, errors.New("a username and password is requred for a setup")
	}

	inputData, err := json.Marshal(SetupRequest{
		Username:           c.username,
		Password:           c.password,
		Org:                org,
		Bucket:             bucket,
		RetentionPeriodHrs: retentionPeriodHrs,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", c.url.String()+"/setup", bytes.NewBuffer(inputData))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	setupResult := &SetupResult{}
	if err := json.NewDecoder(resp.Body).Decode(setupResult); err != nil {
		return nil, err
	}
	if setupResult.Code != "conflict" && resp.StatusCode == http.StatusCreated && setupResult.Auth.Token != "" {
		c.l.Lock()
		c.l.Unlock()
	}
	return setupResult, nil
}

// SetupRequest is a request to setup a new influx instance.
type SetupRequest struct {
	Username           string `json:"username"`
	Password           string `json:"password"`
	Org                string `json:"org"`
	Bucket             string `json:"bucket"`
	RetentionPeriodHrs int    `json:"retentionPeriodHrs"`
}

// SetupResult is the result of setting up a new influx instance.
type SetupResult struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	User    struct {
		Links struct {
			Logs string `json:"logs"`
			Self string `json:"self"`
		} `json:"links"`
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
	Bucket struct {
		ID             string `json:"id"`
		OrganizationID string `json:"organizationID"`
		Organization   string `json:"organization"`
		Name           string `json:"name"`
		RetentionRules []struct {
			Type         string `json:"type"`
			EverySeconds int    `json:"everySeconds"`
		} `json:"retentionRules"`
		Links struct {
			Labels  string `json:"labels"`
			Logs    string `json:"logs"`
			Members string `json:"members"`
			Org     string `json:"org"`
			Owners  string `json:"owners"`
			Self    string `json:"self"`
			Write   string `json:"write"`
		} `json:"links"`
	} `json:"bucket"`
	Org struct {
		Links struct {
			Buckets    string `json:"buckets"`
			Dashboards string `json:"dashboards"`
			Labels     string `json:"labels"`
			Logs       string `json:"logs"`
			Members    string `json:"members"`
			Owners     string `json:"owners"`
			Secrets    string `json:"secrets"`
			Self       string `json:"self"`
			Tasks      string `json:"tasks"`
		} `json:"links"`
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"org"`
	Auth struct {
		ID          string `json:"id"`
		Token       string `json:"token"`
		Status      string `json:"status"`
		Description string `json:"description"`
		OrgID       string `json:"orgID"`
		Org         string `json:"org"`
		UserID      string `json:"userID"`
		User        string `json:"user"`
		Permissions []struct {
			Action   string `json:"action"`
			Resource struct {
				Type string `json:"type"`
			} `json:"resource"`
		} `json:"permissions"`
		Links struct {
			Self string `json:"self"`
			User string `json:"user"`
		} `json:"links"`
	} `json:"auth"`
}
