// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package tests. seed_test_utils provides integration test seeder(s).
package tests

//
//import (
//	"fmt"
//	"github.com/casdoor/casdoor/conf"
//	"github.com/casdoor/casdoor/object"
//	"github.com/casdoor/casdoor/util"
//)
//
//// CleanupFunc is a function that cleans up test data. It returns an error if the cleanup fails.
//// This is useful when we want to write an integration test against actual database instead
//// of a test container.
//// It is returned as a cleanup function that should be deferred to remove the seeded data.
//type CleanupFunc func() error
//
//// -------------------------------------------- Public Functions --------------------------------------------
//
//// SeedBuiltInTestData seeds organization, users and application for integration testing.
//func SeedBuiltInTestData(extraUsers ...*object.User) (cleanup CleanupFunc, err error) {
//	var createdUsers []*object.User
//
//	organization := object.InitBuiltInOrganization()
//	if organization == nil {
//		return cleanup, globalTypes.ErrNil
//	}
//
//	application := object.InitBuiltInApplication()
//	if application == nil {
//		return cleanup, globalTypes.ErrNil
//	}
//
//	builtinUser := object.InitBuiltInUser()
//	if builtinUser == nil {
//		return cleanup, globalTypes.ErrNil
//	}
//
//	createdUsers = append(createdUsers, builtinUser)
//
//	// create more test users
//	users := []*object.User{
//		{
//			Owner:               organization.Name,
//			Name:                "admin0",
//			Id:                  util.GenerateId(),
//			CreatedTime:         util.GetCurrentTime(),
//			Type:                "normal-user",
//			Password:            "123",
//			DisplayName:         "Admin0",
//			Avatar:              fmt.Sprintf("%s/img/casbin.svg", conf.GetConfigString("staticBaseUrl")),
//			Email:               "admin0@example.com",
//			Phone:               "12345678911",
//			CountryCode:         "IR",
//			Address:             []string{},
//			Affiliation:         "Example Inc.",
//			Tag:                 "staff",
//			Score:               2000,
//			Ranking:             1,
//			IsAdmin:             true,
//			IsForbidden:         false,
//			IsDeleted:           false,
//			SignupApplication:   application.Name,
//			CreatedIp:           "127.0.0.1",
//			Properties:          make(map[string]string),
//			Identifier:          "admin0",
//			NationalId:          "0000000000",
//			IsRegistering:       false,
//			CompletedSmartLogin: true,
//		},
//		{
//			Owner:               organization.Name,
//			Name:                "testu",
//			Id:                  util.GenerateId(),
//			CreatedTime:         util.GetCurrentTime(),
//			Type:                "normal-user",
//			Password:            "123",
//			DisplayName:         "TestU",
//			Avatar:              fmt.Sprintf("%s/img/casbin.svg", conf.GetConfigString("staticBaseUrl")),
//			Email:               "testu@example.com",
//			Phone:               "12345678912",
//			CountryCode:         "IR",
//			Address:             []string{},
//			Affiliation:         "Example Inc.",
//			Tag:                 "staff",
//			Score:               2000,
//			Ranking:             1,
//			IsAdmin:             true,
//			IsForbidden:         false,
//			IsDeleted:           false,
//			SignupApplication:   application.Name,
//			CreatedIp:           "127.0.0.1",
//			Properties:          make(map[string]string),
//			Identifier:          "testu",
//			NationalId:          "0000000000",
//			IsRegistering:       false,
//			CompletedSmartLogin: true,
//		},
//	}
//	organization.HasPrivilegeConsent = true
//	if _, err = object.UpdateOrganization(organization.GetId(), organization, true); err != nil {
//		return nil, err
//	}
//	for _, user := range users {
//		affected, err := object.AddUser(user, "en")
//		if err != nil || !affected {
//			return nil, err
//		}
//		createdUsers = append(createdUsers, user)
//	}
//	for _, extraUser := range extraUsers {
//		affected, err := object.AddUser(extraUser, "en")
//		if err != nil || !affected {
//			return nil, err
//		}
//		createdUsers = append(createdUsers, extraUser)
//	}
//
//	cleanup = func() error {
//		for _, user := range createdUsers {
//			if _, err = object.DeleteUser(user); err != nil {
//				return err
//			}
//		}
//		if _, err = object.DeleteApplication(application); err != nil {
//			return err
//		}
//		if _, err = object.DeleteOrganization(organization); err != nil {
//			return err
//		}
//		return nil
//	}
//	return cleanup, nil
//}
//
//// SeedTestGroup creates a test group in the database.
//func SeedTestGroup(owner, name string) (user *object.Group, cleanup CleanupFunc, err error) {
//	group := &object.Group{
//		Owner: owner,
//		Name:  name,
//		// Add other required fields.
//	}
//	affected, err := object.AddGroup(group)
//	if err != nil || !affected {
//		return nil, nil, err
//	}
//
//	cleanup = func() error {
//		if _, err := object.DeleteGroup(group); err != nil {
//			return err
//		}
//		return nil
//	}
//	return group, cleanup, nil
//}
//
//// SeedPasswordPoliciesTestData seeds groups and users with password policies for integration testing.
//func SeedPasswordPoliciesTestData(extraUsers ...*object.User) (cleanup CleanupFunc, err error) {
//	var createdGroups []*object.Group
//	var createdUsers []*object.User
//
//	// Define groups with password policies.
//	groupsData := []struct {
//		owner        string
//		name         string
//		parentID     string
//		isTopGroup   bool
//		expireUnit   globalTypes.ExpireUnit
//		expireValue  int
//		historyCount int
//	}{
//		{"built-in", "client", "", true, globalTypes.ExpireMinutes, 1, 2},
//		{"built-in", "staff", "employee", false, globalTypes.ExpireDays, 1, 1},
//		{"built-in", "developer", "staff", false, globalTypes.ExpireMinutes, 3, 0},
//		{"built-in", "devops", "staff", false, globalTypes.ExpireDays, 60, 8},
//		{"built-in", "employee", "", true, globalTypes.ExpireMinutes, 1, 2},
//		{"built-in", "temporary", "employee", false, globalTypes.ExpireDays, 1, 3},
//	}
//
//	// Create groups.
//	for _, g := range groupsData {
//		group := &object.Group{
//			Owner:       g.owner,
//			Name:        g.name,
//			CreatedTime: util.GetCurrentTime(),
//			DisplayName: g.name,
//			Type:        "Default",
//			IsTopGroup:  g.isTopGroup,
//			PasswordPolicy: &object.PasswordPolicy{
//				PasswordExpireUnit:  g.expireUnit,
//				PasswordExpireValue: g.expireValue,
//				MaxPasswordHistory:  g.historyCount,
//			},
//		}
//		if g.parentID != "" {
//			group.ParentId = g.parentID
//		}
//
//		affected, err := object.AddGroup(group)
//		if err != nil || !affected {
//			return cleanup, err
//		}
//		createdGroups = append(createdGroups, group)
//	}
//
//	// Define users with group memberships.
//	usersData := []struct {
//		owner             string
//		name              string
//		groups            []string
//		signupApplication string
//	}{
//		{"built-in", "admin", []string{"built-in/employee"}, "app-built-in"},
//		{"built-in", "Admin0", []string{"built-in/devops", "built-in/staff", "built-in/employee"}, "app-built-in"},
//		{"built-in", "Seyed Mohammad Ali", []string{"built-in/developer", "built-in/staff", "built-in/employee", "built-in/temporary", "built-in/client"}, "app-built-in"},
//	}
//
//	// Create users.
//	for _, u := range usersData {
//		user := &object.User{
//			Owner:                  u.owner,
//			Name:                   u.name,
//			CreatedTime:            util.GetCurrentTime(),
//			Id:                     util.GenerateId(),
//			Type:                   "normal-user",
//			Password:               "password",
//			DisplayName:            u.name,
//			Avatar:                 "https://example.com/avatar.jpg",
//			Email:                  u.name + "@example.com",
//			Phone:                  "+1234567890",
//			CountryCode:            "US",
//			Groups:                 u.groups,
//			LastChangePasswordTime: util.GetCurrentTime(),
//			SignupApplication:      u.signupApplication,
//		}
//
//		foundUser, err := object.GetUser(user.GetId())
//		if err != nil {
//			return cleanup, err
//		}
//		if foundUser != nil {
//			continue
//		}
//		affected, err := object.AddUser(user, "en")
//		if err != nil || !affected {
//			return cleanup, err
//		}
//		createdUsers = append(createdUsers, user)
//	}
//	for _, extraUser := range extraUsers {
//		affected, err := object.AddUser(extraUser, "en")
//		if err != nil || !affected {
//			return cleanup, err
//		}
//		createdUsers = append(createdUsers, extraUser)
//	}
//
//	// Define cleanup function to remove seeded data.
//	cleanup = func() error {
//		for _, user := range createdUsers {
//			if _, err = object.DeleteUser(user); err != nil {
//				return err
//			}
//		}
//
//		for _, group := range createdGroups {
//			if _, err = object.DeleteGroup(group); err != nil {
//				return err
//			}
//		}
//		return nil
//	}
//
//	return cleanup, nil
//}
