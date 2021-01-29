// +build e2e

// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api_test

import (
	"context"
	"testing"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLabelsAPI(t *testing.T) {
	client := influxdb2.NewClientWithOptions(serverURL, authToken, influxdb2.DefaultOptions().SetLogLevel(3))
	labelsAPI := client.LabelsAPI()
	orgAPI := client.OrganizationsAPI()

	ctx := context.Background()

	myorg, err := orgAPI.FindOrganizationByName(ctx, "my-org")
	require.Nil(t, err, err)
	require.NotNil(t, myorg)

	labels, err := labelsAPI.GetLabels(ctx)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, *labels, 0)

	labelName := "Active State"
	props := map[string]string{"color": "#33ffddd", "description": "Marks org active"}
	label, err := labelsAPI.CreateLabelWithName(ctx, myorg, labelName, props)
	require.Nil(t, err, err)
	require.NotNil(t, label)
	assert.Equal(t, labelName, *label.Name)
	require.NotNil(t, label.Properties)
	assert.Equal(t, props, label.Properties.AdditionalProperties)

	//remove properties
	label.Properties.AdditionalProperties = map[string]string{"color": "", "description": ""}
	label2, err := labelsAPI.UpdateLabel(ctx, label)
	require.Nil(t, err, err)
	require.NotNil(t, label2)
	assert.Equal(t, labelName, *label2.Name)
	assert.Nil(t, label2.Properties)

	label2, err = labelsAPI.FindLabelByID(ctx, *label.Id)
	require.Nil(t, err, err)
	require.NotNil(t, label2)
	assert.Equal(t, labelName, *label2.Name)

	label2, err = labelsAPI.FindLabelByName(ctx, *myorg.Id, labelName)
	require.Nil(t, err, err)
	require.NotNil(t, label2)
	assert.Equal(t, labelName, *label2.Name)

	label2, err = labelsAPI.FindLabelByName(ctx, *myorg.Id, "wrong label")
	assert.NotNil(t, err, err)
	assert.Nil(t, label2)

	labels, err = labelsAPI.GetLabels(ctx)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, *labels, 1)

	labels, err = labelsAPI.FindLabelsByOrg(ctx, myorg)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, *labels, 1)

	labels, err = labelsAPI.FindLabelsByOrgID(ctx, *myorg.Id)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, *labels, 1)

	// duplicate label
	label2, err = labelsAPI.CreateLabelWithName(ctx, myorg, labelName, nil)
	assert.NotNil(t, err)
	assert.Nil(t, label2)

	err = labelsAPI.DeleteLabel(ctx, label)
	require.Nil(t, err, err)
	//
	err = labelsAPI.DeleteLabel(ctx, label)
	assert.NotNil(t, err, err)
}

func TestLabelsAPI_failing(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)
	clientUnAuth := influxdb2.NewClient(serverURL, "invalid_token")
	labelsAPI := client.LabelsAPI()
	orgAPI := client.OrganizationsAPI()
	ctx := context.Background()

	invalidID := "xyz"
	wrongID := "1000000000000000"

	var label = &domain.Label{
		Id: &wrongID,
	}

	org, err := orgAPI.FindOrganizationByName(ctx, "my-org")
	require.Nil(t, err, err)
	require.NotNil(t, org)

	label, err = labelsAPI.UpdateLabel(ctx, label)
	assert.NotNil(t, err)
	assert.Nil(t, label)

	label, err = labelsAPI.FindLabelByID(ctx, wrongID)
	assert.NotNil(t, err)
	assert.Nil(t, label)

	labels, err := clientUnAuth.LabelsAPI().FindLabelsByOrgID(ctx, invalidID)
	assert.NotNil(t, err)
	assert.Nil(t, labels)

	err = labelsAPI.DeleteLabelWithID(ctx, invalidID)
	assert.NotNil(t, err)
}

func TestLabelsAPI_requestFailing(t *testing.T) {
	client := influxdb2.NewClient("serverURL", authToken)
	labelsAPI := client.LabelsAPI()
	ctx := context.Background()

	anID := "1000000000000000"

	label := &domain.Label{Id: &anID}

	_, err := labelsAPI.GetLabels(ctx)
	assert.NotNil(t, err)

	_, err = labelsAPI.FindLabelByName(ctx, anID, "name")
	assert.NotNil(t, err)

	_, err = labelsAPI.FindLabelByID(ctx, anID)
	assert.NotNil(t, err)

	_, err = labelsAPI.CreateLabelWithNameWithID(ctx, anID, "name", nil)
	assert.NotNil(t, err)

	_, err = labelsAPI.UpdateLabel(ctx, label)
	assert.NotNil(t, err)

	err = labelsAPI.DeleteLabel(ctx, label)
	assert.NotNil(t, err)
}
