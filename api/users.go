package api

import (
	"context"
	"github.com/influxdata/influxdb-client-go/domain"
	ihttp "github.com/influxdata/influxdb-client-go/internal/http"
)

// UsersApi provides methods for managing users in a InfluxDB server
type UsersApi interface {
	// FindUsers returns all users
	FindUsers(ctx context.Context) (*[]domain.User, error)
	// FindUserByID returns user with userID
	FindUserByID(ctx context.Context, userID string) (*domain.User, error)
	// CreateUser creates new user
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	// CreateUserWithName creates new user with userName
	CreateUserWithName(ctx context.Context, userName string) (*domain.User, error)
	// UpdateUser updates user
	UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	// UpdateUserPassword sets password for an user
	UpdateUserPassword(ctx context.Context, user *domain.User, password string) error
	// UpdateUserPasswordWithID sets password for an user with userId
	UpdateUserPasswordWithID(ctx context.Context, userID string, password string) error
	// DeleteUserWithId deletes an user with userId
	DeleteUserWithID(ctx context.Context, userID string) error
	// DeleteUser deletes an user
	DeleteUser(ctx context.Context, user *domain.User) error
	// Me returns actual user
	Me(ctx context.Context) (*domain.User, error)
	// MeUpdatePassword set password of actual user
	MeUpdatePassword(ctx context.Context, password string) error
}

type usersApiImpl struct {
	apiClient *domain.ClientWithResponses
}

func NewUsersApi(service ihttp.Service) UsersApi {

	apiClient := domain.NewClientWithResponses(service)
	return &usersApiImpl{
		apiClient: apiClient,
	}
}

func (u *usersApiImpl) FindUsers(ctx context.Context) (*[]domain.User, error) {
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

func (u *usersApiImpl) FindUserByID(ctx context.Context, userID string) (*domain.User, error) {
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

func (u *usersApiImpl) CreateUserWithName(ctx context.Context, userName string) (*domain.User, error) {
	user := &domain.User{Name: userName}
	return u.CreateUser(ctx, user)
}

func (u *usersApiImpl) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
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

func (u *usersApiImpl) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
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

func (u *usersApiImpl) UpdateUserPassword(ctx context.Context, user *domain.User, password string) error {
	return u.UpdateUserPasswordWithID(ctx, *user.Id, password)
}

func (u *usersApiImpl) UpdateUserPasswordWithID(ctx context.Context, userID string, password string) error {
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

func (u *usersApiImpl) DeleteUser(ctx context.Context, user *domain.User) error {
	return u.DeleteUserWithID(ctx, *user.Id)
}

func (u *usersApiImpl) DeleteUserWithID(ctx context.Context, userID string) error {
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

func (u *usersApiImpl) Me(ctx context.Context) (*domain.User, error) {
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

func (u *usersApiImpl) MeUpdatePassword(ctx context.Context, password string) error {
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
