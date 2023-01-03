// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient

import (
	"context"
	"fmt"

	"github.com/influxdata/influxdb-client-go/v3/influxclient/model"
)

// UsersAPI holds methods related to user, as found under
// the /users endpoint.
type UsersAPI struct {
	client *model.Client
}

// newUsersAPI returns new UsersPI instance
func newUsersAPI(client *model.Client) *UsersAPI {
	return &UsersAPI{client: client}
}

// Find returns all users matching the given filter.
// Supported filters:
//   - ID
//   - Name
func (a *UsersAPI) Find(ctx context.Context, filter *Filter) ([]model.User, error) {
	return a.getUsers(ctx, filter)
}

// FindOne returns one user matching the given filter.
// Supported filters:
//   - ID
//   - Name
func (a *UsersAPI) FindOne(ctx context.Context, filter *Filter) (*model.User, error) {
	users, err := a.getUsers(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(users) > 0 {
		return &(users[0]), nil
	}
	return nil, fmt.Errorf("user not found")
}

// Create creates a user. The user.Name must not be empty.
func (a *UsersAPI) Create(ctx context.Context, user *model.User) (*model.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}
	if user.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	params := &model.PostUsersAllParams{
		Body: model.PostUsersJSONRequestBody{
			Name: user.Name,
		},
	}
	if user.Status != nil {
		params.Body.Status = user.Status
	}
	response, err := a.client.PostUsers(ctx, params)
	if err != nil {
		return nil, err
	}
	return userResponseToUser(response), nil
}

// Update updates a user. The user.ID field must be specified.
// The complete user information is returned.
func (a *UsersAPI) Update(ctx context.Context, user *model.User) (*model.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}
	if user.Id == nil {
		return nil, fmt.Errorf("user ID is required")
	}
	if user.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	params := &model.PatchUsersIDAllParams{
		UserID: *user.Id,
		Body: model.PatchUsersIDJSONRequestBody{
			Name: user.Name,
		},
	}
	if user.Status != nil {
		params.Body.Status = user.Status
	}
	response, err := a.client.PatchUsersID(ctx, params)
	if err != nil {
		return nil, err
	}
	return userResponseToUser(response), nil
}

// Delete deletes the user with the given ID.
func (a *UsersAPI) Delete(ctx context.Context, userID string) error {
	params := &model.DeleteUsersIDAllParams{
		UserID: userID,
	}
	return a.client.DeleteUsersID(ctx, params)
}

// SetPassword sets the password for the user with the given ID.
func (a *UsersAPI) SetPassword(ctx context.Context, userID, password string) error {
	params := &model.PostUsersIDPasswordAllParams{
		UserID: userID,
		Body: model.PostUsersIDPasswordJSONRequestBody{
			Password: password,
		},
	}
	return a.client.PostUsersIDPassword(ctx, params)
}

// SetMyPassword sets the password associated with the current user.
// The oldPassword parameter must match the previously set password
// for the user.
func (a *UsersAPI) SetMyPassword(ctx context.Context, oldPassword, newPassword string) error {
	_, err := a.getMe(ctx)
	if err != nil {
		return err
	}
	params := &model.PutMePasswordAllParams{
		Body: model.PutMePasswordJSONRequestBody{
			Password: newPassword,
		},
	}
	return a.client.PutMePassword(ctx, params)
}

// getUsers create request for GET on /users according to the filter and validates returned structure
func (a *UsersAPI) getUsers(ctx context.Context, filter *Filter) ([]model.User, error) {
	params := &model.GetUsersParams{}
	if filter != nil {
		if filter.ID != "" {
			params.Id = &filter.ID
		}
		if filter.Name != "" {
			params.Name = &filter.Name
		}
		if filter.Limit > 0 {
			limit := model.Limit(filter.Limit)
			params.Limit = &limit
		}
		if filter.Offset > 0 {
			offset := model.Offset(filter.Offset)
			params.Offset = &offset
		}
	}
	response, err := a.client.GetUsers(ctx, params)
	if err != nil {
		return nil, err
	}
	return userResponsesToUsers(response.Users), nil
}

// getMe retrieves currently authenticated user information.
func (a *UsersAPI) getMe(ctx context.Context) (*model.User, error) {
	params := &model.GetMeParams{}
	response, err := a.client.GetMe(ctx, params)
	if err != nil {
		return nil, err
	}
	return userResponseToUser(response), nil
}

func userResponseToUser(ur *model.UserResponse) *model.User {
	if ur == nil {
		return nil
	}
	user := &model.User{
		Id:     ur.Id,
		Name:   ur.Name,
		Status: userResponseStatusToUserStatus(ur.Status),
	}
	return user
}

func userResponseStatusToUserStatus(urs *model.UserResponseStatus) *model.UserStatus {
	if urs == nil {
		return nil
	}
	us := model.UserStatus(*urs)
	return &us
}

func userResponsesToUsers(urs *[]model.UserResponse) []model.User {
	if urs == nil {
		return nil
	}
	us := make([]model.User, len(*urs))
	for i, ur := range *urs {
		us[i] = *userResponseToUser(&ur)
	}
	return us
}
