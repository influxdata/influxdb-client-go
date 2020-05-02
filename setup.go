// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/influxdata/influxdb-client-go/domain"
	"log"
	"net/http"
	"net/url"
	"path"
)

func (c *clientImpl) Setup(ctx context.Context, username, password, org, bucket string, retentionPeriodHours int) (*domain.OnboardingResponse, error) {
	if username == "" || password == "" {
		return nil, errors.New("a username and password is required for a setup")
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	var setupResult *domain.OnboardingResponse
	inputData, err := json.Marshal(domain.OnboardingRequest{
		Username:           username,
		Password:           password,
		Org:                org,
		Bucket:             bucket,
		RetentionPeriodHrs: &retentionPeriodHours,
	})
	if err != nil {
		return nil, err
	}
	if c.options.LogLevel() > 2 {
		log.Printf("D! Request:\n%s\n", string(inputData))
	}
	u, err := url.Parse(c.httpService.ServerApiUrl())
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "setup")

	perror := c.httpService.PostRequest(ctx, u.String(), bytes.NewReader(inputData), func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json; charset=utf-8")
	},
		func(resp *http.Response) error {
			defer func() {
				_ = resp.Body.Close()
			}()
			setupResponse := &domain.OnboardingResponse{}
			if err := json.NewDecoder(resp.Body).Decode(setupResponse); err != nil {
				return err
			}
			setupResult = setupResponse
			if setupResponse.Auth != nil && *setupResponse.Auth.Token != "" {
				c.httpService.SetAuthorization("Token " + *setupResponse.Auth.Token)
			}
			return nil
		},
	)
	if perror != nil {
		return nil, perror
	}
	return setupResult, nil
}
