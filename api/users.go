// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/domain"
)

// UsersAPI provides methods for managing users in a InfluxDB server
type UsersAPI interface {
	// GetUsers returns all users
	GetUsers(ctx context.Context) (*[]domain.User, error)
	// FindUserByID returns user with userID
	FindUserByID(ctx context.Context, userID string) (*domain.User, error)
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
	// UpdateUserPasswordWithID sets password for an user with userID
	UpdateUserPasswordWithID(ctx context.Context, userID string, password string) error
	// DeleteUserWithID deletes an user with userID
	DeleteUserWithID(ctx context.Context, userID string) error
	// DeleteUser deletes an user
	DeleteUser(ctx context.Context, user *domain.User) error
	// Me returns actual user
	Me(ctx context.Context) (*domain.User, error)
	// MeUpdatePassword set password of actual user
	MeUpdatePassword(ctx context.Context, password string) error
}

type usersAPI struct {
	apiClient *domain.ClientWithResponses
}

func NewUsersAPI(apiClient *domain.ClientWithResponses) UsersAPI {
	return &usersAPI{
		apiClient: apiClient,
	}
}

func (u *usersAPI) GetUsers(ctx context.Context) (*[]domain.User, error) {
	params := &domain.GetUsersParams{}
	response, err := u.apiClient.GetUsersWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON200.Users, nil
}

func (u *usersAPI) FindUserByID(ctx context.Context, userID string) (*domain.User, error) {
	params := &domain.GetUsersIDParams{}
	response, err := u.apiClient.GetUsersIDWithResponse(ctx, userID, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON200, nil
}

func (u *usersAPI) FindUserByName(ctx context.Context, userName string) (*domain.User, error) {
	users, err := u.GetUsers(ctx)
	if err != nil {
		return nil, err
	}
	var user *domain.User
	for _, u := range *users {
		if u.Name == userName {
			user = &u
			break
		}
	}
	if user == nil {
		return nil, fmt.Errorf("user '%s' not found", userName)
	}
	return user, nil
}

func (u *usersAPI) CreateUserWithName(ctx context.Context, userName string) (*domain.User, error) {
	user := &domain.User{Name: userName}
	return u.CreateUser(ctx, user)
}

func (u *usersAPI) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	params := &domain.PostUsersParams{}
	response, err := u.apiClient.PostUsersWithResponse(ctx, params, domain.PostUsersJSONRequestBody(*user))
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON201, nil
}

func (u *usersAPI) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	params := &domain.PatchUsersIDParams{}
	response, err := u.apiClient.PatchUsersIDWithResponse(ctx, *user.Id, params, domain.PatchUsersIDJSONRequestBody(*user))
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON200, nil
}

func (u *usersAPI) UpdateUserPassword(ctx context.Context, user *domain.User, password string) error {
	return u.UpdateUserPasswordWithID(ctx, *user.Id, password)
}

func (u *usersAPI) UpdateUserPasswordWithID(ctx context.Context, userID string, password string) error {
	params := &domain.PostUsersIDPasswordParams{}
	body := &domain.PasswordResetBody{Password: password}
	response, err := u.apiClient.PostUsersIDPasswordWithResponse(ctx, userID, params, domain.PostUsersIDPasswordJSONRequestBody(*body))
	if err != nil {
		return err
	}
	if response.JSONDefault != nil {
		return domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return nil
}

func (u *usersAPI) DeleteUser(ctx context.Context, user *domain.User) error {
	return u.DeleteUserWithID(ctx, *user.Id)
}

func (u *usersAPI) DeleteUserWithID(ctx context.Context, userID string) error {
	params := &domain.DeleteUsersIDParams{}
	response, err := u.apiClient.DeleteUsersIDWithResponse(ctx, userID, params)
	if err != nil {
		return err
	}
	if response.JSONDefault != nil {
		return domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return nil
}

func (u *usersAPI) Me(ctx context.Context) (*domain.User, error) {
	params := &domain.GetMeParams{}
	response, err := u.apiClient.GetMeWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON200, nil
}

func (u *usersAPI) MeUpdatePassword(ctx context.Context, password string) error {
	params := &domain.PutMePasswordParams{}
	body := &domain.PasswordResetBody{Password: password}
	response, err := u.apiClient.PutMePasswordWithResponse(ctx, params, domain.PutMePasswordJSONRequestBody(*body))
	if err != nil {
		return err
	}
	if response.JSONDefault != nil {
		return domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return nil
}
