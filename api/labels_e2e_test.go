// +build e2e

// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api_test

import (
	"context"
	"testing"

	influxdb2 "github.com/influxdata/influxdb-client-go"
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

	label2, err = labelsAPI.FindLabelByID(ctx, "000000000000000")
	require.NotNil(t, err, err)
	require.Nil(t, label2)

	label2, err = labelsAPI.FindLabelByName(ctx, *myorg.Id, labelName)
	require.Nil(t, err, err)
	require.NotNil(t, label2)
	assert.Equal(t, labelName, *label2.Name)

	label2, err = labelsAPI.FindLabelByName(ctx, *myorg.Id, "wrong label")
	require.NotNil(t, err, err)
	require.Nil(t, label2)

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

	labels, err = labelsAPI.FindLabelsByOrgID(ctx, "000000000000000")
	require.NotNil(t, err, err)
	require.Nil(t, labels)

	// duplicate label
	label2, err = labelsAPI.CreateLabelWithName(ctx, myorg, labelName, nil)
	require.NotNil(t, err, err)
	require.Nil(t, label2)

	labels, err = orgAPI.GetLabels(ctx, myorg)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, *labels, 0)

	org, err := orgAPI.CreateOrganizationWithName(ctx, "org1")
	require.Nil(t, err, err)
	require.NotNil(t, org)

	labels, err = orgAPI.GetLabels(ctx, org)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, *labels, 0)

	labelx, err := orgAPI.AddLabel(ctx, org, label)
	require.Nil(t, err, err)
	require.NotNil(t, labelx)

	labels, err = orgAPI.GetLabels(ctx, org)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, *labels, 1)

	err = orgAPI.RemoveLabel(ctx, org, label)
	require.Nil(t, err, err)

	labels, err = orgAPI.GetLabels(ctx, org)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, *labels, 0)

	labels, err = orgAPI.GetLabelsWithID(ctx, "000000000000000")
	require.NotNil(t, err, err)
	require.Nil(t, labels)

	label2, err = orgAPI.AddLabelWithID(ctx, *org.Id, "000000000000000")
	require.NotNil(t, err, err)
	require.Nil(t, label2)

	label2, err = orgAPI.AddLabelWithID(ctx, "000000000000000", "000000000000000")
	require.NotNil(t, err, err)
	require.Nil(t, label2)

	err = orgAPI.RemoveLabelWithID(ctx, *org.Id, "000000000000000")
	require.NotNil(t, err, err)
	require.Nil(t, label2)

	err = orgAPI.RemoveLabelWithID(ctx, "000000000000000", "000000000000000")
	require.NotNil(t, err, err)
	require.Nil(t, label2)

	err = orgAPI.DeleteOrganization(ctx, org)
	assert.Nil(t, err, err)

	err = labelsAPI.DeleteLabel(ctx, label)
	require.Nil(t, err, err)

	err = labelsAPI.DeleteLabel(ctx, label)
	require.NotNil(t, err, err)
}
