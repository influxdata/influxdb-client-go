// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package examples contains fake client with the same interface as real client to overcome import-cycle problem
// to allow real E2E examples for apis in this package
package examples

import (
	"context"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

// Options is fake options to satisfy Client interface
type Options struct {
}

// SetBatchSize to emulate fake options
func (o *Options) SetBatchSize(_ uint) *Options {
	return o
}

// FakeClient emulates Client for allowing using client in api examples
type FakeClient struct {
}

// NewClient returns new FakeClient
func NewClient(_ string, _ string) *FakeClient {
	client := &FakeClient{}
	return client
}

// Options returns nil
func (c *FakeClient) Options() *Options {
	return nil
}

// ServerURL returns empty server URL
func (c *FakeClient) ServerURL() string {
	return ""
}

// Ready does nothing
func (c *FakeClient) Ready(_ context.Context) (bool, error) {
	return true, nil
}

// Setup does nothing
func (c *FakeClient) Setup(_ context.Context, _, _, _, _ string, _ int) (*domain.OnboardingResponse, error) {
	return nil, nil
}

// Health does nothing
func (c *FakeClient) Health(_ context.Context) (*domain.HealthCheck, error) {
	return nil, nil
}

// WriteAPI does nothing
func (c *FakeClient) WriteAPI(_, _ string) api.WriteAPI {
	return nil
}

// WriteAPIBlocking does nothing
func (c *FakeClient) WriteAPIBlocking(_, _ string) api.WriteAPIBlocking {
	return nil
}

// Close does nothing
func (c *FakeClient) Close() {
}

// QueryAPI returns nil
func (c *FakeClient) QueryAPI(_ string) api.QueryAPI {
	return nil
}

// AuthorizationsAPI returns nil
func (c *FakeClient) AuthorizationsAPI() api.AuthorizationsAPI {
	return nil
}

// OrganizationsAPI returns nil
func (c *FakeClient) OrganizationsAPI() api.OrganizationsAPI {
	return nil
}

// UsersAPI returns nil
func (c *FakeClient) UsersAPI() api.UsersAPI {
	return nil
}

// DeleteAPI returns nil
func (c *FakeClient) DeleteAPI() api.DeleteAPI {
	return nil
}

// BucketsAPI returns nil
func (c *FakeClient) BucketsAPI() api.BucketsAPI {
	return nil
}

// LabelsAPI returns nil
func (c *FakeClient) LabelsAPI() api.LabelsAPI {
	return nil
}

// TasksAPI returns nil
func (c *FakeClient) TasksAPI() api.TasksAPI {
	return nil
}
