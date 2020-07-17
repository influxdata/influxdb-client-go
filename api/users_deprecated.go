// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"github.com/influxdata/influxdb-client-go/domain"
)

//lint:file-ignore ST1003 This is deprecated API to be removed in next release.

// UsersApi provides methods for managing users in a InfluxDB server.
// Deprecated: use UsersAPI instead.
type UsersApi interface {
	// GetUsers returns all users
	GetUsers(ctx context.Context) (*[]domain.User, error)
	// FindUserById returns user with userID
	FindUserById(ctx context.Context, userID string) (*domain.User, error)
	// FindUserByName returns user with name userName
	FindUserByName(ctx context.Context, userName string) (*domain.User, error)
	// CreateUser creates new user
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	// CreateUserWithName creates new user with userName
	CreateUserWithName(ctx context.Context, userName string) (*domain.User, error)
	// UpdateUser updates user
	UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	// UpdateUserPassword sets password for an user
	UpdateUserPassword(ctx context.Context, user *domain.User, password string) error
	// UpdateUserPasswordWithId sets password for an user with userId
	UpdateUserPasswordWithId(ctx context.Context, userID string, password string) error
	// DeleteUserWithId deletes an user with userId
	DeleteUserWithId(ctx context.Context, userID string) error
	// DeleteUser deletes an user
	DeleteUser(ctx context.Context, user *domain.User) error
	// Me returns actual user
	Me(ctx context.Context) (*domain.User, error)
	// MeUpdatePassword set password of actual user
	MeUpdatePassword(ctx context.Context, password string) error
}

type usersApiImpl struct {
	usersAPI UsersAPI
}

// NewUsersApi creates instance of UsersApi
// Deprecated: use NewUsersAPI instead.
func NewUsersApi(apiClient *domain.ClientWithResponses) UsersApi {
	return &usersApiImpl{
		usersAPI: NewUsersAPI(apiClient),
	}
}

func (u *usersApiImpl) GetUsers(ctx context.Context) (*[]domain.User, error) {
	return u.usersAPI.GetUsers(ctx)
}

func (u *usersApiImpl) FindUserById(ctx context.Context, userID string) (*domain.User, error) {
	return u.usersAPI.FindUserByID(ctx, userID)
}

func (u *usersApiImpl) FindUserByName(ctx context.Context, userName string) (*domain.User, error) {
	return u.usersAPI.FindUserByName(ctx, userName)
}

func (u *usersApiImpl) CreateUserWithName(ctx context.Context, userName string) (*domain.User, error) {
	return u.usersAPI.CreateUserWithName(ctx, userName)
}

func (u *usersApiImpl) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	return u.usersAPI.CreateUser(ctx, user)
}

func (u *usersApiImpl) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	return u.usersAPI.UpdateUser(ctx, user)
}

func (u *usersApiImpl) UpdateUserPassword(ctx context.Context, user *domain.User, password string) error {
	return u.usersAPI.UpdateUserPassword(ctx, user, password)
}

func (u *usersApiImpl) UpdateUserPasswordWithId(ctx context.Context, userID string, password string) error {
	return u.usersAPI.UpdateUserPasswordWithID(ctx, userID, password)
}

func (u *usersApiImpl) DeleteUser(ctx context.Context, user *domain.User) error {
	return u.usersAPI.DeleteUser(ctx, user)
}

func (u *usersApiImpl) DeleteUserWithId(ctx context.Context, userID string) error {
	return u.usersAPI.DeleteUserWithID(ctx, userID)
}

func (u *usersApiImpl) Me(ctx context.Context) (*domain.User, error) {
	return u.usersAPI.Me(ctx)
}

func (u *usersApiImpl) MeUpdatePassword(ctx context.Context, password string) error {
	return u.usersAPI.MeUpdatePassword(ctx, password)
}
