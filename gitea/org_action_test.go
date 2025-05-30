// Copyright 2024 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gitea

import (
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateOrgActionSecret(t *testing.T) {
	log.Println("== TestCreateOrgActionSecret ==")
	c := newTestClient()

	user := createTestUser(t, "org_action_user", c)
	c.SetSudo(user.UserName)
	newOrg, _, err := c.CreateOrg(CreateOrgOption{Name: "ActionOrg"})
	assert.NoError(t, err)
	assert.NotNil(t, newOrg)

	// create secret
	resp, err := c.CreateOrgActionSecret(newOrg.UserName, CreateSecretOption{Name: "test", Data: "test"})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// update secret
	resp, err = c.CreateOrgActionSecret(newOrg.UserName, CreateSecretOption{Name: "test", Data: "test2"})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// list secrets
	secrets, _, err := c.ListOrgActionSecret(newOrg.UserName, ListOrgActionSecretOption{})
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
}

func TestUpdateOrgActionVariable(t *testing.T) {
	log.Println("== TestUpdateOrgActionVariable ==")
	c := newTestClient()

	user := createTestUser(t, "org_action_update_var_user", c)
	c.SetSudo(user.UserName)
	org, _, err := c.CreateOrg(CreateOrgOption{Name: "ActionUpdateVarOrg"})
	assert.NoError(t, err)
	assert.NotNil(t, org)

	// Create variable
	resp, err := c.CreateOrgActionVariable(org.UserName, CreateOrgActionVariableOption{
		Name:  "UPDATE_VAR",
		Value: "before",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Update variable (success)
	resp, err = c.UpdateOrgActionVariable(org.UserName, "UPDATE_VAR", UpdateOrgActionVariableOption{
		Value: "after",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Get variable and check updated value
	v, resp, err := c.GetOrgActionVariable(org.UserName, "UPDATE_VAR")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "UPDATE_VAR", v.Name)
	assert.Equal(t, "after", v.Data)

	// Update variable (empty value, should error)
	_, err = c.UpdateOrgActionVariable(org.UserName, "UPDATE_VAR", UpdateOrgActionVariableOption{
		Value: "",
	})
	assert.Error(t, err)

	// Update variable (not found)
	resp, err = c.UpdateOrgActionVariable(org.UserName, "NOT_EXIST_VAR", UpdateOrgActionVariableOption{
		Value: "something",
	})
	assert.Error(t, err)
	if resp != nil {
		assert.True(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden)
	}
}

func TestOrgActionVariable(t *testing.T) {
	log.Println("== TestOrgActionVariable ==")
	c := newTestClient()

	user := createTestUser(t, "org_action_var_user", c)
	c.SetSudo(user.UserName)
	org, _, err := c.CreateOrg(CreateOrgOption{Name: "ActionVarOrg"})
	assert.NoError(t, err)
	assert.NotNil(t, org)

	// Create variable (success)
	resp, err := c.CreateOrgActionVariable(org.UserName, CreateOrgActionVariableOption{
		Name:  "TEST_VAR",
		Value: "initial",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Create duplicate variable (should return Conflict)
	resp, err = c.CreateOrgActionVariable(org.UserName, CreateOrgActionVariableOption{
		Name:  "TEST_VAR",
		Value: "updated",
	})
	assert.Error(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	// List variables
	vars, _, err := c.ListOrgActionVariable(org.UserName, ListOrgActionVariableOption{})
	assert.NoError(t, err)
	assert.NotNil(t, vars)
	assert.True(t, len(vars) >= 1)
	found := false
	for _, v := range vars {
		if v.Name == "TEST_VAR" {
			found = true
			assert.Equal(t, "initial", v.Data)
		}
	}
	assert.True(t, found, "TEST_VAR should be in the list")

	// Get variable (success)
	v, resp, err := c.GetOrgActionVariable(org.UserName, "TEST_VAR")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "TEST_VAR", v.Name)
	assert.Equal(t, "initial", v.Data)

	// Get variable (not found)
	_, resp, err = c.GetOrgActionVariable(org.UserName, "NOT_EXIST")
	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Create variable (invalid cases)
	cases := []struct {
		name        string
		opt         CreateOrgActionVariableOption
		expectError bool
	}{
		{
			name:        "empty name",
			opt:         CreateOrgActionVariableOption{Name: "", Value: "v"},
			expectError: true,
		},
		{
			name:        "name too long",
			opt:         CreateOrgActionVariableOption{Name: "THIS_NAME_IS_WAY_TOO_LONG_FOR_THE_LIMIT", Value: "v"},
			expectError: true,
		},
		{
			name:        "empty value",
			opt:         CreateOrgActionVariableOption{Name: "EMPTY_VALUE", Value: ""},
			expectError: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := c.CreateOrgActionVariable(org.UserName, tc.opt)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if resp != nil {
				assert.True(t, resp.StatusCode >= 400)
			}
		})
	}
}
