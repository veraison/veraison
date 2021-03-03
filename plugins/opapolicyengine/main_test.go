package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

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

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	endorsementsPath := filepath.Join(wd, "test", "endorsements.json")
	evidencePath := filepath.Join(wd, "test", "evidence.json")
	policyPath := filepath.Join(wd, "test", "policy.rego")

	endorsements, err := readJSONObjectFromFile(endorsementsPath)
	if err != nil {
		t.Fatalf("%v", err)
	}

	evidence, err := readJSONObjectFromFile(evidencePath)
	if err != nil {
		t.Fatalf("%v", err)
	}

	policyData, err := ioutil.ReadFile(policyPath)
	if err != nil {
		t.Fatalf("%v", err)
	}

	var pe OpaPolicyEngine
	var peParams common.PolicyEngineParams

	if err = pe.Init(peParams); err != nil {
		t.Fatalf("%v", err)
	}

	if err = pe.LoadPolicy(policyData); err != nil {
		t.Fatalf("%v", err)
	}

	isValid, err := pe.CheckValid(evidence, endorsements)
	require.Nil(err)
	require.True(isValid)
}
