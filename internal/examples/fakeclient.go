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

func (o *Options) SetBatchSize(_ uint) *Options {
	return o
}

type fakeClient struct {
}

func NewClient(_ string, _ string) *fakeClient {
	client := &fakeClient{}
	return client
}

func NewClientWithOptions(_ string, _ string, _ *Options) *fakeClient {
	client := &fakeClient{}
	return client
}

func (c *fakeClient) Options() *Options {
	return nil
}

func (c *fakeClient) ServerURL() string {
	return ""
}

func (c *fakeClient) Ready(_ context.Context) (bool, error) {
	return true, nil
}

func (c *fakeClient) Setup(_ context.Context, _, _, _, _ string, _ int) (*domain.OnboardingResponse, error) {
	return nil, nil
}

func (c *fakeClient) Health(_ context.Context) (*domain.HealthCheck, error) {
	return nil, nil
}

func (c *fakeClient) WriteAPI(_, _ string) api.WriteAPI {
	return nil
}

func (c *fakeClient) WriteAPIBlocking(_, _ string) api.WriteAPIBlocking {
	return nil
}

func (c *fakeClient) Close() {
}

func (c *fakeClient) QueryAPI(_ string) api.QueryAPI {
	return nil
}

func (c *fakeClient) AuthorizationsAPI() api.AuthorizationsAPI {
	return nil
}

func (c *fakeClient) OrganizationsAPI() api.OrganizationsAPI {
	return nil
}

func (c *fakeClient) UsersAPI() api.UsersAPI {
	return nil
}

func (c *fakeClient) DeleteAPI() api.DeleteAPI {
	return nil
}

func (c *fakeClient) BucketsAPI() api.BucketsAPI {
	return nil
}

func (c *fakeClient) LabelsAPI() api.LabelsAPI {
	return nil
}

func (c *fakeClient) TasksAPI() api.TasksAPI {
	return nil
}
