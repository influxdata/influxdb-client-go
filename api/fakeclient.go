// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"github.com/influxdata/influxdb-client-go/domain"
)

// This file contains fake client with the same interface as real client to overcome import-cycle problem
// to allow real E2E examples for apis in this package

// Fake options to satisfy Client itnerface
type Options struct {
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

//func (c *fakeClient) WriteApi(org, bucket string) WriteApi {
//	return nil
//}

//func (c *fakeClient) WriteApiBlocking(org, bucket string) WriteApiBlocking {
//	w := newWriteApiBlockingImpl(org, bucket, c.httpService, c)
//	return w
//}

func (c *fakeClient) Close() {
}

//func (c *fakeClient) QueryApi(org string) QueryApi {
//	return nil
//}

func (c *fakeClient) AuthorizationsApi() AuthorizationsApi {
	return nil
}

func (c *fakeClient) OrganizationsApi() OrganizationsApi {
	return nil
}

func (c *fakeClient) UsersApi() UsersApi {
	return nil
}

func (c *fakeClient) DeleteApi() DeleteApi {
	return nil
}

func (c *fakeClient) BucketsApi() BucketsApi {
	return nil
}
