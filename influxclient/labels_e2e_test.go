//go:build e2e
// +build e2e

// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/influxdata/influxdb-client-go/v3/influxclient"
	"github.com/influxdata/influxdb-client-go/v3/influxclient/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLabelsAPI(t *testing.T) {
	client, ctx := newClient(t)
	labelsAPI := client.LabelsAPI()
	orgAPI := client.OrganizationsAPI()

	org, err := orgAPI.FindOne(ctx, &Filter{
		OrgName: orgName,
	})
	require.Nil(t, err, err)
	require.NotNil(t, org)

	// find without filter
	labels, err := labelsAPI.Find(ctx, nil)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, labels, 0)

	// find
	labels, err = labelsAPI.Find(ctx, &Filter{
		OrgID: *org.Id,
	})
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, labels, 0)

	// create label 1 without properties
	label1Name := "Basic State"
	label1, err := labelsAPI.Create(ctx, &model.Label{
		Name:  &label1Name,
		OrgID: org.Id,
	})
	require.Nil(t, err, err)
	require.NotNil(t, label1)
	defer labelsAPI.Delete(ctx, safeId(label1.Id))
	assert.Equal(t, label1Name, *label1.Name)
	require.Nil(t, label1.Properties)

	// create label 2 with properties
	label2Name := "Active State"
	props2 := map[string]string{"color": "#33ffddd", "description": "Marks state active"}
	label2, err := labelsAPI.Create(ctx, &model.Label{
		Name:       &label2Name,
		OrgID:      org.Id,
		Properties: &model.Label_Properties{AdditionalProperties: props2},
	})
	require.Nil(t, err, err)
	require.NotNil(t, label2)
	defer labelsAPI.Delete(ctx, safeId(label2.Id))
	assert.Equal(t, label2Name, *label2.Name)
	require.NotNil(t, label2.Properties)
	diff := cmp.Diff([]model.Label{}, labels)
	assert.Equal(t, props2, label2.Properties.AdditionalProperties, "diff: %s", diff)

	// find
	labels, err = labelsAPI.Find(ctx, &Filter{OrgID: org.Name})
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, labels, 2)

	// remove properties
	label2.Properties.AdditionalProperties = map[string]string{"color": "", "description": ""}
	label2, err = labelsAPI.Update(ctx, label2)
	require.Nil(t, err, err)
	require.NotNil(t, label2)
	assert.Equal(t, label2Name, *label2.Name)
	assert.Nil(t, label2.Properties)

	// delete label 2
	err = labelsAPI.Delete(ctx, *label2.Id)
	require.Nil(t, err, err)

	// find
	labels, err = labelsAPI.Find(ctx, &Filter{OrgID: org.Name})
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, labels, 1)

	// find one
	label, err := labelsAPI.FindOne(ctx, &Filter{OrgID: org.Name})
	require.Nil(t, err, err)
	require.NotNil(t, label)
	assert.Equal(t, label1.Name, label.Name)
	assert.Equal(t, label1.Properties, label.Properties)

	// try to create label with existing name
	label3, err := labelsAPI.Create(ctx, &model.Label{
		Name:  label1.Name,
		OrgID: org.Id,
	})
	assert.Error(t, err)
	assert.Nil(t, label3)

	// delete label
	err = labelsAPI.Delete(ctx, *label1.Id)
	require.Nil(t, err, err)

	// find
	labels, err = labelsAPI.Find(ctx, &Filter{OrgID: org.Name})
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, labels, 0)
}

func TestLabelsAPI_failing(t *testing.T) {
	client, ctx := newClient(t)
	labelsAPI := client.LabelsAPI()

	label, err := labelsAPI.FindOne(ctx, &Filter{
		OrgID: invalidID,
	})
	assert.Error(t, err)
	assert.Nil(t, label)

	label, err = labelsAPI.Create(ctx, nil)
	assert.Error(t, err)
	assert.Nil(t, label)

	label, err = labelsAPI.Create(ctx, &model.Label{})
	assert.Error(t, err)
	assert.Nil(t, label)

	name := "a label"
	label, err = labelsAPI.Create(ctx, &model.Label{
		Name:  &name,
		OrgID: nil,
	})
	assert.Error(t, err)
	assert.Nil(t, label)

	label, err = labelsAPI.Create(ctx, &model.Label{
		Name:  &name,
		OrgID: &invalidID,
	})
	assert.Error(t, err)
	assert.Nil(t, label)

	label, err = labelsAPI.Update(ctx, nil)
	assert.Error(t, err)
	assert.Nil(t, label)

	label, err = labelsAPI.Update(ctx, &model.Label{})
	assert.Error(t, err)
	assert.Nil(t, label)

	label, err = labelsAPI.Update(ctx, &model.Label{
		Id: &invalidID,
	})
	assert.Error(t, err)
	assert.Nil(t, label)

	err = labelsAPI.Delete(ctx, notExistingID)
	require.Error(t, err)

	err = labelsAPI.Delete(ctx, invalidID)
	assert.Error(t, err)
}
