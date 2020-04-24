// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package api provides clients for InfluxDB server APIs
//
// Examples
//
// Users API
//
//	// Create influxdb client
//	client := influxdb2.NewClient("http://localhost:9999", "my-token")
//
//	// Find organization
//	org, err := client.OrganizationsApi().FindOrganizationByName(context.Background(), "my-org")
//	if err != nil {
//		panic(err)
//	}
//
//	// Get users API client
//	usersApi := client.UsersApi()
//
//	// Create new user
//	user, err := usersApi.CreateUserWithName(context.Background(), "user-01")
//	if err != nil {
//		panic(err)
//	}
//
//	// Set user password
//	err = usersApi.UpdateUserPassword(context.Background(), user, "pass-at-least-8-chars")
//	if err != nil {
//		panic(err)
//	}
//
//	// Add user to organization
//	_, err = client.OrganizationsApi().AddMember(context.Background(), org, user)
//	if err != nil {
//		panic(err)
//	}
//
// Organizations API
//
//	// Create influxdb client
//	client := influxdb2.NewClient("http://localhost:9999", "my-token")
//
//	// Get Organizations API client
//	orgApi := client.OrganizationsApi()
//
//	// Create new organization
//	org, err := orgApi.CreateOrganizationWithName(context.Background(), "org-2")
//	if err != nil {
//		panic(err)
//	}
//
//	orgDescription := "My second org "
//	org.Description = &orgDescription
//
//	org, err = orgApi.UpdateOrganization(context.Background(), org)
//	if err != nil {
//		panic(err)
//	}
//
//	// Find user to set owner
//	user, err := client.UsersApi().FindUserByName(context.Background(), "user-01")
//	if err != nil {
//		panic(err)
//	}
//
//	// Add another owner (first owner is the one who create organization
//	_, err = orgApi.AddOwner(context.Background(), org, user)
//	if err != nil {
//		panic(err)
//	}
//
//	// Create new user to add to org
//	newUser, err := client.UsersApi().CreateUserWithName(context.Background(), "user-02")
//	if err != nil {
//		panic(err)
//	}
//
//	// Add new user to organization
//	_, err = orgApi.AddMember(context.Background(), org, newUser)
//	if err != nil {
//		panic(err)
//	}
//
// Authorizations API
//
//	// Create influxdb client
//	client := influxdb2.NewClient("http://localhost:9999", "my-token")
//
//	// Find user to grant permission
//	user, err := client.UsersApi().FindUserByName(context.Background(), "user-01")
//	if err != nil {
//		panic(err)
//	}
//
//	// Find organization
//	org, err := client.OrganizationsApi().FindOrganizationByName(context.Background(), "my-org")
//	if err != nil {
//		panic(err)
//	}
//
//	// create write permission for buckets
//	permissionWrite := &domain.Permission{
//		Action: domain.PermissionActionWrite,
//		Resource: domain.Resource{
//			Type: domain.ResourceTypeBuckets,
//		},
//	}
//
//	// create read permission for buckets
//	permissionRead := &domain.Permission{
//		Action: domain.PermissionActionRead,
//		Resource: domain.Resource{
//			Type: domain.ResourceTypeBuckets,
//		},
//	}
//
//	// group permissions
//	permissions := []domain.Permission{*permissionWrite, *permissionRead}
//
//	// create authorization object using info above
//	auth := &domain.Authorization{
//		OrgID:       org.Id,
//		Permissions: &permissions,
//		User:        &user.Name,
//		UserID:      user.Id,
//	}
//
//	// grant permission and create token
//	authCreated, err := client.AuthorizationsApi().CreateAuthorization(context.Background(), auth)
//	if err != nil {
//		panic(err)
//	}
//
//	// Use token
//	fmt.Println("Token: ", *authCreated.Token)
package api
