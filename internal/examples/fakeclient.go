// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package examples

import (
	"context"
	"github.com/influxdata/influxdb-client-go/api"
	"github.com/influxdata/influxdb-client-go/domain"
)

// This file contains fake client with the same interface as real client to overcome import-cycle problem
// to allow real E2E examples for apis in this package

// Fake options to satisfy Client itnerface
type Options struct {
}

func (o *Options) SetBatchSize(_ uint) *Options {
	return o
}

func DefaultOptions() *Options {
	return nil
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

func (c *fakeClient) ServerUrl() string {
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

func (c *fakeClient) WriteApi(_, _ string) api.WriteApi {
	return nil
}

func (c *fakeClient) WriteApiBlocking(_, _ string) api.WriteApiBlocking {
	return nil
}

func (c *fakeClient) Close() {
}

func (c *fakeClient) QueryApi(_ string) api.QueryApi {
	return nil
}

func (c *fakeClient) AuthorizationsApi() api.AuthorizationsApi {
	return nil
}

func (c *fakeClient) OrganizationsApi() api.OrganizationsApi {
	return nil
}

func (c *fakeClient) UsersApi() api.UsersApi {
	return nil
}

func (c *fakeClient) DeleteApi() api.DeleteApi {
	return nil
}

func (c *fakeClient) BucketsApi() api.BucketsApi {
	return nil
}
