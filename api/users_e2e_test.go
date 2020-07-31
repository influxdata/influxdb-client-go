// +build e2e

// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api_test

import (
	"context"
	"testing"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersAPI(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)
	usersAPI := client.UsersAPI()
	ctx := context.Background()

	me, err := usersAPI.Me(ctx)
	require.Nil(t, err)
	require.NotNil(t, me)

	users, err := usersAPI.GetUsers(ctx)
	require.Nil(t, err)
	require.NotNil(t, users)
	assert.Len(t, *users, 1)

	user, err := usersAPI.CreateUserWithName(ctx, "user-01")
	require.Nil(t, err)
	require.NotNil(t, user)

	// create duplicate user
	user2, err := usersAPI.CreateUserWithName(ctx, "user-01")
	assert.NotNil(t, err)
	assert.Nil(t, user2)

	users, err = usersAPI.GetUsers(ctx)
	require.Nil(t, err)
	require.NotNil(t, users)
	assert.Len(t, *users, 2)

	status := domain.UserStatusInactive
	user.Status = &status
	user, err = usersAPI.UpdateUser(ctx, user)
	require.Nil(t, err)
	require.NotNil(t, user)
	assert.Equal(t, status, *user.Status)

	user2 = &domain.User{
		Id:   user.Id,
		Name: "my-user",
	}
	//update username to existing user
	user2, err = usersAPI.UpdateUser(ctx, user2)
	assert.NotNil(t, err)
	assert.Nil(t, user2)

	user, err = usersAPI.FindUserByID(ctx, *user.Id)
	require.Nil(t, err)
	require.NotNil(t, user)

	err = usersAPI.UpdateUserPassword(ctx, user, "my-password")
	require.Nil(t, err)

	err = usersAPI.DeleteUser(ctx, user)
	require.Nil(t, err)

	users, err = usersAPI.GetUsers(ctx)
	require.Nil(t, err)
	require.NotNil(t, users)
	assert.Len(t, *users, 1)

	// this fails now
	err = usersAPI.MeUpdatePassword(ctx, "new-password")
	assert.NotNil(t, err)
}

func TestUsersAPI_failing(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)
	usersAPI := client.UsersAPI()
	ctx := context.Background()

	invalidID := "aaaaaa"

	user, err := usersAPI.FindUserByID(ctx, invalidID)
	assert.NotNil(t, err)
	assert.Nil(t, user)

	user, err = usersAPI.FindUserByName(ctx, "not-existing-name")
	assert.NotNil(t, err)
	assert.Nil(t, user)

	err = usersAPI.DeleteUserWithID(ctx, invalidID)
	assert.NotNil(t, err)

	err = usersAPI.UpdateUserPasswordWithID(ctx, invalidID, "pass")
	assert.NotNil(t, err)
}

func TestUsersAPI_requestFailing(t *testing.T) {
	client := influxdb2.NewClient("serverURL", authToken)
	usersAPI := client.UsersAPI()
	ctx := context.Background()

	invalidID := "aaaaaa"

	user := &domain.User{
		Id: &invalidID,
	}
	_, err := usersAPI.GetUsers(ctx)
	assert.NotNil(t, err)

	_, err = usersAPI.FindUserByID(ctx, invalidID)
	assert.NotNil(t, err)

	_, err = usersAPI.FindUserByName(ctx, "not-existing-name")
	assert.NotNil(t, err)

	_, err = usersAPI.CreateUserWithName(ctx, "not-existing-name")
	assert.NotNil(t, err)

	_, err = usersAPI.UpdateUser(ctx, user)
	assert.NotNil(t, err)

	err = usersAPI.UpdateUserPasswordWithID(ctx, invalidID, "pass")
	assert.NotNil(t, err)

	err = usersAPI.DeleteUserWithID(ctx, invalidID)
	assert.NotNil(t, err)

	_, err = usersAPI.Me(ctx)
	assert.NotNil(t, err)

	err = usersAPI.MeUpdatePassword(ctx, "pass")
	assert.NotNil(t, err)
}
