//go:build e2e

// Copyright 2024 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE fil
package api_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

var msg = "Check: ${ r._check_name } is: ${ r._level }"
var flux = `from(bucket: "foo") |> range(start: -1d, stop: now()) |> aggregateWindow(every: 1m, fn: mean) |> filter(fn: (r) => r._field == "usage_user") |> yield()`
var every = "1h"
var offset = "0s"
var timeSince = "90m"
var staleTime = "30m"
var level = domain.CheckStatusLevelCRIT

func TestCreateGetDeleteThresholdCheck(t *testing.T) {
	ctx := context.Background()
	client := influxdb2.NewClient(serverURL, authToken)

	greater := domain.GreaterThreshold{}
	greater.Value = 10.0
	lc := domain.CheckStatusLevelCRIT
	greater.Level = &lc
	greater.AllValues = &[]bool{true}[0]

	lesser := domain.LesserThreshold{}
	lesser.Value = 1.0
	lo := domain.CheckStatusLevelOK
	lesser.Level = &lo

	rang := domain.RangeThreshold{}
	rang.Min = 3.0
	rang.Max = 8.0
	lw := domain.CheckStatusLevelWARN
	rang.Level = &lw

	org, err := client.OrganizationsAPI().FindOrganizationByName(ctx, "my-org")
	require.Nil(t, err)

	thresholds := []domain.Threshold{&greater, &lesser, &rang}

	check := domain.ThresholdCheck{
		CheckBaseExtend: domain.CheckBaseExtend{
			CheckBase: domain.CheckBase{
				Name:   "ThresholdCheck test",
				OrgID:  *org.Id,
				Query:  domain.DashboardQuery{Text: &flux},
				Status: domain.TaskStatusTypeActive,
			},
			Every:                 &every,
			Offset:                &offset,
			StatusMessageTemplate: &msg,
		},
		Thresholds: &thresholds,
	}
	params := domain.CreateCheckAllParams{
		Body: domain.CreateCheckJSONRequestBody(&check),
	}
	nc, err := client.APIClient().CreateCheck(context.Background(), &params)
	require.Nil(t, err)
	tc := validateTC(t, nc, *org.Id)

	gp := domain.GetChecksIDAllParams{CheckID: *tc.Id}

	c, err := client.APIClient().GetChecksID(context.Background(), &gp)
	require.Nil(t, err)
	tc = validateTC(t, c, *org.Id)

	dp := domain.DeleteChecksIDAllParams{
		CheckID: *tc.Id,
	}

	err = client.APIClient().DeleteChecksID(context.Background(), &dp)
	require.Nil(t, err)

	_, err = client.APIClient().GetChecksID(context.Background(), &gp)
	require.NotNil(t, err)
}

func TestCreateGetDeleteDeadmanCheck(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)

	ctx := context.Background()

	org, err := client.OrganizationsAPI().FindOrganizationByName(ctx, "my-org")
	require.Nil(t, err)

	check := domain.DeadmanCheck{
		CheckBaseExtend: domain.CheckBaseExtend{
			CheckBase: domain.CheckBase{
				Name:   "DeadmanCheck test",
				OrgID:  *org.Id,
				Query:  domain.DashboardQuery{Text: &flux},
				Status: domain.TaskStatusTypeActive,
			},
			Every:                 &every,
			Offset:                &offset,
			StatusMessageTemplate: &msg,
		},
		TimeSince: &timeSince,
		StaleTime: &staleTime,
		Level:     &level,
	}
	params := domain.CreateCheckAllParams{
		Body: domain.CreateCheckJSONRequestBody(&check),
	}
	nc, err := client.APIClient().CreateCheck(context.Background(), &params)
	require.Nil(t, err)
	dc := validateDC(t, nc, *org.Id)

	gp := domain.GetChecksIDAllParams{CheckID: *dc.Id}

	c, err := client.APIClient().GetChecksID(context.Background(), &gp)
	require.Nil(t, err)
	dc = validateDC(t, c, *org.Id)

	dp := domain.DeleteChecksIDAllParams{
		CheckID: *dc.Id,
	}

	err = client.APIClient().DeleteChecksID(context.Background(), &dp)
	require.Nil(t, err)

}

func TestUpdateThresholdCheck(t *testing.T) {
	ctx := context.Background()
	client := influxdb2.NewClient(serverURL, authToken)

	greater := domain.GreaterThreshold{}
	greater.Value = 10.0
	lc := domain.CheckStatusLevelCRIT
	greater.Level = &lc
	greater.AllValues = &[]bool{true}[0]

	org, err := client.OrganizationsAPI().FindOrganizationByName(ctx, "my-org")
	require.Nil(t, err)

	thresholds := []domain.Threshold{&greater}

	ev := "1m"

	check := domain.ThresholdCheck{
		CheckBaseExtend: domain.CheckBaseExtend{
			CheckBase: domain.CheckBase{
				Name:   "ThresholdCheck update test",
				OrgID:  *org.Id,
				Query:  domain.DashboardQuery{Text: &flux},
				Status: domain.TaskStatusTypeActive,
			},
			Every:                 &ev,
			Offset:                &offset,
			StatusMessageTemplate: &msg,
		},
		Thresholds: &thresholds,
	}
	params := domain.CreateCheckAllParams{
		Body: domain.CreateCheckJSONRequestBody(&check),
	}
	nc, err := client.APIClient().CreateCheck(context.Background(), &params)
	require.Nil(t, err)
	require.NotNil(t, nc)
	tc := nc.(*domain.ThresholdCheck)

	lesser := domain.LesserThreshold{}
	lesser.Value = 1.0
	lo := domain.CheckStatusLevelOK
	lesser.Level = &lo

	rang := domain.RangeThreshold{}
	rang.Min = 3.0
	rang.Max = 8.0
	lw := domain.CheckStatusLevelWARN
	rang.Level = &lw

	thresholds = []domain.Threshold{&greater, &lesser, &rang}
	tc.Thresholds = &thresholds
	tc.Every = &every
	tc.Name = "ThresholdCheck test"

	updateParams := domain.PutChecksIDAllParams{
		CheckID: *tc.Id,
		Body:    tc,
	}
	nc, err = client.APIClient().PutChecksID(context.Background(), &updateParams)
	require.Nil(t, err)
	require.NotNil(t, nc)
	tc = validateTC(t, nc, *org.Id)

	dp := domain.DeleteChecksIDAllParams{
		CheckID: *tc.Id,
	}

	err = client.APIClient().DeleteChecksID(context.Background(), &dp)
	require.Nil(t, err)
}

func TestGetChecks(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)

	ctx := context.Background()

	org, err := client.OrganizationsAPI().FindOrganizationByName(ctx, "my-org")
	require.Nil(t, err)

	check := domain.DeadmanCheck{
		CheckBaseExtend: domain.CheckBaseExtend{
			CheckBase: domain.CheckBase{
				Name:   "DeadmanCheck test",
				OrgID:  *org.Id,
				Query:  domain.DashboardQuery{Text: &flux},
				Status: domain.TaskStatusTypeActive,
			},
			Every:                 &every,
			Offset:                &offset,
			StatusMessageTemplate: &msg,
		},
		TimeSince: &timeSince,
		StaleTime: &staleTime,
		Level:     &level,
	}
	params := domain.CreateCheckAllParams{
		Body: domain.CreateCheckJSONRequestBody(&check),
	}
	nc, err := client.APIClient().CreateCheck(context.Background(), &params)
	require.Nil(t, err)
	validateDC(t, nc, *org.Id)

	greater := domain.GreaterThreshold{}
	greater.Value = 10.0
	lc := domain.CheckStatusLevelCRIT
	greater.Level = &lc
	greater.AllValues = &[]bool{true}[0]

	lesser := domain.LesserThreshold{}
	lesser.Value = 1.0
	lo := domain.CheckStatusLevelOK
	lesser.Level = &lo

	rang := domain.RangeThreshold{}
	rang.Min = 3.0
	rang.Max = 8.0
	lw := domain.CheckStatusLevelWARN
	rang.Level = &lw

	thresholds := []domain.Threshold{&greater, &lesser, &rang}

	check2 := domain.ThresholdCheck{
		CheckBaseExtend: domain.CheckBaseExtend{
			CheckBase: domain.CheckBase{
				Name:   "ThresholdCheck test",
				OrgID:  *org.Id,
				Query:  domain.DashboardQuery{Text: &flux},
				Status: domain.TaskStatusTypeActive,
			},
			Every:                 &every,
			Offset:                &offset,
			StatusMessageTemplate: &msg,
		},
		Thresholds: &thresholds,
	}
	params2 := domain.CreateCheckAllParams{
		Body: domain.CreateCheckJSONRequestBody(&check2),
	}
	nc, err = client.APIClient().CreateCheck(context.Background(), &params2)
	require.Nil(t, err)
	validateTC(t, nc, *org.Id)

	gp := domain.GetChecksParams{
		OrgID: *org.Id,
	}
	checks, err := client.APIClient().GetChecks(context.Background(), &gp)
	require.Nil(t, err)
	require.NotNil(t, checks)
	require.NotNil(t, checks.Checks)
	assert.Len(t, *checks.Checks, 2)
	dc := validateDC(t, (*checks.Checks)[0], *org.Id)
	tc := validateTC(t, (*checks.Checks)[1], *org.Id)

	dp := domain.DeleteChecksIDAllParams{
		CheckID: *dc.Id,
	}
	err = client.APIClient().DeleteChecksID(context.Background(), &dp)
	require.Nil(t, err)

	dp2 := domain.DeleteChecksIDAllParams{
		CheckID: *tc.Id,
	}
	err = client.APIClient().DeleteChecksID(context.Background(), &dp2)
	require.Nil(t, err)

	checks, err = client.APIClient().GetChecks(context.Background(), &gp)
	require.Nil(t, err)
	require.NotNil(t, checks)
	assert.Nil(t, checks.Checks)
}

func validateDC(t *testing.T, nc domain.Check, orgID string) *domain.DeadmanCheck {
	require.NotNil(t, nc)
	require.Equal(t, "deadman", nc.Type())
	dc := nc.(*domain.DeadmanCheck)
	require.NotNil(t, dc)
	assert.Equal(t, "DeadmanCheck test", dc.Name)
	assert.Equal(t, orgID, dc.OrgID)
	assert.Equal(t, msg, *dc.StatusMessageTemplate)
	assert.Equal(t, flux, *dc.Query.Text)
	assert.Equal(t, every, *dc.Every)
	assert.Equal(t, offset, *dc.Offset)
	assert.Equal(t, domain.TaskStatusTypeActive, dc.Status)
	assert.Equal(t, timeSince, *dc.TimeSince)
	assert.Equal(t, staleTime, *dc.StaleTime)
	assert.Equal(t, domain.CheckStatusLevelCRIT, *dc.Level)
	return dc
}

func validateTC(t *testing.T, check domain.Check, orgID string) *domain.ThresholdCheck {
	require.NotNil(t, check)
	require.Equal(t, "threshold", check.Type())
	tc := check.(*domain.ThresholdCheck)
	require.NotNil(t, tc)
	assert.Equal(t, "ThresholdCheck test", tc.Name)
	assert.Equal(t, orgID, tc.OrgID)
	assert.Equal(t, msg, *tc.StatusMessageTemplate)
	assert.Equal(t, flux, *tc.Query.Text)
	assert.Equal(t, every, *tc.Every)
	assert.Equal(t, offset, *tc.Offset)
	assert.Equal(t, domain.TaskStatusTypeActive, tc.Status)
	assert.Len(t, *tc.Thresholds, 3)
	require.Equal(t, "greater", (*tc.Thresholds)[0].Type())
	gt := (*tc.Thresholds)[0].(*domain.GreaterThreshold)
	require.NotNil(t, gt)
	assert.Equal(t, float32(10.0), gt.Value)
	assert.Equal(t, domain.CheckStatusLevelCRIT, *gt.Level)
	assert.Equal(t, true, *gt.AllValues)
	require.Equal(t, "lesser", (*tc.Thresholds)[1].Type())
	lt := (*tc.Thresholds)[1].(*domain.LesserThreshold)
	require.NotNil(t, lt)
	assert.Equal(t, float32(1.0), lt.Value)
	assert.Equal(t, domain.CheckStatusLevelOK, *lt.Level)
	require.Equal(t, "range", (*tc.Thresholds)[2].Type())
	rt := (*tc.Thresholds)[2].(*domain.RangeThreshold)
	require.NotNil(t, rt)
	assert.Equal(t, float32(3.0), rt.Min)
	assert.Equal(t, float32(8.0), rt.Max)
	assert.Equal(t, domain.CheckStatusLevelWARN, *rt.Level)
	return tc
}
