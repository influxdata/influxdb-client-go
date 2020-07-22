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

	me, err := usersAPI.Me(context.Background())
	require.Nil(t, err)
	require.NotNil(t, me)

	users, err := usersAPI.GetUsers(context.Background())
	require.Nil(t, err)
	require.NotNil(t, users)
	assert.Len(t, *users, 1)

	user, err := usersAPI.CreateUserWithName(context.Background(), "user-01")
	require.Nil(t, err)
	require.NotNil(t, user)

	users, err = usersAPI.GetUsers(context.Background())
	require.Nil(t, err)
	require.NotNil(t, users)
	assert.Len(t, *users, 2)

	status := domain.UserStatusInactive
	user.Status = &status
	user, err = usersAPI.UpdateUser(context.Background(), user)
	require.Nil(t, err)
	require.NotNil(t, user)
	assert.Equal(t, status, *user.Status)

	user, err = usersAPI.FindUserByID(context.Background(), *user.Id)
	require.Nil(t, err)
	require.NotNil(t, user)

	err = usersAPI.UpdateUserPassword(context.Background(), user, "my-password")
	require.Nil(t, err)

	err = usersAPI.DeleteUser(context.Background(), user)
	require.Nil(t, err)

	users, err = usersAPI.GetUsers(context.Background())
	require.Nil(t, err)
	require.NotNil(t, users)
	assert.Len(t, *users, 1)
}
