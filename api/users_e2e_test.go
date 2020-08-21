// +build e2e

// Copyright 2020 InfluxData, Inc. All rights reserved.
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

	// it fails, https://github.com/influxdata/influxdb/pull/15981
	//err = usersAPI.MeUpdatePassword(ctx, "my-password", "my-new-password")
	//assert.Nil(t, err)

	//err = usersAPI.MeUpdatePassword(ctx, "my-new-password", "my-password")
	//assert.Nil(t, err)
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

	err = usersAPI.MeUpdatePassword(ctx, "my-password", "my-new-password")
	assert.NotNil(t, err)

	err = usersAPI.SignIn(ctx, "user", "my-password")
	assert.NotNil(t, err)

	err = usersAPI.SignOut(ctx)
	assert.NotNil(t, err)
}

func TestSignInOut(t *testing.T) {
	ctx := context.Background()
	client := influxdb2.NewClient("http://localhost:9999", "")

	usersAPI := client.UsersAPI()

	err := usersAPI.SignIn(ctx, "my-user", "my-password")
	require.Nil(t, err)

	// try authorized calls
	orgs, err := client.OrganizationsAPI().GetOrganizations(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, orgs)

	// try authorized calls
	buckets, err := client.BucketsAPI().GetBuckets(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, buckets)

	// try authorized calls
	err = client.WriteAPIBlocking("my-org", "my-bucket").WriteRecord(ctx, "test,a=rock,b=local f=1.2,i=-5i")
	assert.Nil(t, err)

	res, err := client.QueryAPI("my-org").QueryRaw(context.Background(), `from(bucket:"my-bucket")|> range(start: -24h) |> filter(fn: (r) => r._measurement == "test")`, influxdb2.DefaultDialect())
	assert.Nil(t, err)
	assert.NotNil(t, res)

	err = usersAPI.SignOut(ctx)
	assert.Nil(t, err)

	// unauthorized signout
	err = usersAPI.SignOut(ctx)
	assert.NotNil(t, err)

	// Unauthorized call
	_, err = client.OrganizationsAPI().GetOrganizations(ctx)
	assert.NotNil(t, err)

	// test wrong credentials
	err = usersAPI.SignIn(ctx, "my-user", "password")
	assert.NotNil(t, err)

	client.HTTPService().SetAuthorization("Token my-token")

	user, err := usersAPI.CreateUserWithName(ctx, "user-01")
	require.Nil(t, err)
	require.NotNil(t, user)

	// 2nd client to use for new user auth
	client2 := influxdb2.NewClient("http://localhost:9999", "")

	err = usersAPI.UpdateUserPassword(ctx, user, "123password")
	assert.Nil(t, err)

	err = client2.UsersAPI().SignIn(ctx, "user-01", "123password")
	assert.Nil(t, err)

	err = client2.UsersAPI().SignOut(ctx)
	assert.Nil(t, err)

	status := domain.UserStatusInactive
	user.Status = &status
	u, err := usersAPI.UpdateUser(ctx, user)
	assert.Nil(t, err)
	assert.NotNil(t, u)

	// log in inactive user,
	//err = client2.SignIn(ctx, "user-01", "123password")
	//assert.NotNil(t, err)

	err = usersAPI.DeleteUser(ctx, user)
	assert.Nil(t, err)

	client.Close()
	client2.Close()
}
