// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/veraison/common"
)

func readJSONObjectFromFile(path string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	v := make(map[string]interface{})

	err = json.Unmarshal(data, &v)
	return v, err
}

func Test_Validate(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	wd, err := os.Getwd()
	require.Nil(err)

	endorsementsPath := filepath.Join(wd, "test", "endorsements.json")
	evidencePath := filepath.Join(wd, "test", "evidence.json")
	policyPath := filepath.Join(wd, "test", "policy.rego")

	endorsements, err := readJSONObjectFromFile(endorsementsPath)
	require.Nil(err)

	evidence, err := readJSONObjectFromFile(evidencePath)
	require.Nil(err)

	policyData, err := ioutil.ReadFile(policyPath)
	require.Nil(err)

	var pe OpaPolicyEngine

	err = pe.Init(nil)
	require.Nil(err)

	err = pe.LoadPolicy(policyData)
	require.Nil(err)

	status, err := pe.CheckValid(evidence, endorsements)
	assert.Nil(err)
	assert.Equal(common.Status_SUCCESS, status)
}
