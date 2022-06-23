//go:build e2e
// +build e2e

// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient_test

import (
	"testing"

	. "github.com/influxdata/influxdb-client-go/influxclient"
	"github.com/influxdata/influxdb-client-go/influxclient/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersAPI(t *testing.T) {
	client, ctx := newClient(t)
	usersAPI := client.UsersAPI()

	users, err := usersAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, users)
	assert.Len(t, users, 1)

	// create user
	user, err := usersAPI.Create(ctx, &model.User{
		Name: "user-01",
	})
	require.NoError(t, err)
	require.NotNil(t, user)
	defer usersAPI.Delete(ctx, safeId(user.Id))

	// try to create duplicate user
	user2, err := usersAPI.Create(ctx, &model.User{
		Name: "user-01",
	})
	assert.Error(t, err)
	assert.Nil(t, user2)

	// find
	users, err = usersAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, users)
	assert.Len(t, users, 2)

	// update
	status := model.UserStatusInactive // default is active
	user.Status = &status
	user, err = usersAPI.Update(ctx, user)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, status, *user.Status)

	user2 = &model.User{
		Id:   user.Id,
		Name: "my-user2",
	}

	// update username
	user2, err = usersAPI.Update(ctx, user2)
	assert.NoError(t, err)
	assert.NotNil(t, user2)

	// find by ID
	user, err = usersAPI.FindOne(ctx, &Filter{
		ID: *user2.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, user)

	// find by name
	user, err = usersAPI.FindOne(ctx, &Filter{
		Name: user.Name,
	})
	require.NoError(t, err)
	require.NotNil(t, user)

	// find multiple
	users, err = usersAPI.Find(ctx, &Filter{
		Offset: 1,
		Limit: 100,
	})
	require.NoError(t, err)
	require.NotNil(t, users)
	require.Equal(t, 1, len(users))

	// update password
	err = usersAPI.SetPassword(ctx, *user.Id, "my-password2")
	require.NoError(t, err)

	// delete user
	err = usersAPI.Delete(ctx, *user.Id)
	require.NoError(t, err)

	// verify there's just onboarded user
	users, err = usersAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, users)
	assert.Len(t, users, 1)

	// it fails, https://github.com/influxdata/influxdb/pull/15981
	//err = usersAPI.MeUpdatePassword(ctx, "my-password", "my-new-password")
	//assert.NoError(t, err)

	//err = usersAPI.MeUpdatePassword(ctx, "my-new-password", "my-password")
	//assert.NoError(t, err)
}

func TestUsersAPI_failing(t *testing.T) {
	client, ctx := newClient(t)
	usersAPI := client.UsersAPI()

	// user cannot be nil
	user, err := usersAPI.Create(ctx, nil)
	require.Error(t, err)
	assert.Nil(t, user)

	// user name must not be empty
	user, err = usersAPI.Create(ctx, &model.User{})
	require.Error(t, err)
	assert.Nil(t, user)

	user, err = usersAPI.Update(ctx, &model.User{
		Id:   nil,
		Name: "john doe",
	})
	require.Error(t, err)
	assert.Nil(t, user)

	user, err = usersAPI.Update(ctx, nil)
	require.Error(t, err)
	assert.Nil(t, user)

	user, err = usersAPI.Update(ctx, &model.User{
		Id:   &notExistingID,
		Name: "",
	})
	require.Error(t, err)
	assert.Nil(t, user)

	user, err = usersAPI.Update(ctx, &model.User{
		Id:   &notExistingID,
		Name: "john doe",
	})
	require.Error(t, err)
	assert.Nil(t, user)

	user, err = usersAPI.FindOne(ctx, &Filter{
		ID: invalidID,
	})
	assert.Error(t, err)
	assert.Nil(t, user)

	user, err = usersAPI.FindOne(ctx, &Filter{
		Name: "not-existing-name",
	})
	assert.Error(t, err)
	assert.Nil(t, user)

	err = usersAPI.Delete(ctx, invalidID)
	assert.Error(t, err)

	err = usersAPI.SetPassword(ctx, invalidID, "pass")
	assert.Error(t, err)

	err = usersAPI.SetMyPassword(ctx, "wrong", "better")
	assert.Error(t, err)
}
