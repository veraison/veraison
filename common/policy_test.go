// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolicyReadWrite(t *testing.T) {
	assert := assert.New(t)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	policyFile := filepath.Join(wd, "test", "policy.zip")
	policies, err := ReadPoliciesFromPath(policyFile)
	if err != nil {
		t.Fatalf("%v", err)
	}

	assert.Equal(1, len(policies))

	policy := policies[0]
	assert.Equal(TokenFormat_PSA, policy.TokenFormat)
	assert.Equal("$.implementation_id", policy.QueryMap["hardware_id"]["platform_id"])
	assert.Equal("$.sw_components[*].measurement_value",
		policy.QueryMap["software_components"]["measurements"])
	assert.True(strings.HasPrefix(string(policy.Rules), "package iat\n"))

	tFile, err := ioutil.TempFile(os.TempDir(), "policy")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if err = policy.Write(tFile); err != nil {
		t.Fatalf("%v", err)
	}
	tFile.Close()
	defer os.Remove(tFile.Name())

	content, err := ioutil.ReadFile(tFile.Name())
	if err != nil {
		t.Fatalf("%v", err)
	}
	assert.Equal("PK", string(content[:2]))
}
