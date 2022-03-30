// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spf13/viper"
)

func TestParamStore_Values(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	params := NewParamStore("test")

	require.Nil(params.SetInt("one", 1))
	assert.Equal(1, params.GetInt("one"))

	v, err := params.TryGetInt("two")
	assert.Equal(0, v)
	assert.Equal("key not found: \"two\"", err.Error())
	assert.Equal(0, params.GetInt("two"))

	s, err := params.TryGetString("one")
	assert.Equal("", s)
	assert.Equal("cannot get value for \"one\": cannot convert to string: 1 (float64)", err.Error())

	require.Nil(params.SetString("yes", "true"))
	assert.True(params.GetBool("yes"))

	require.Nil(params.SetStringSlice("horses", []string{"bay", "chesnut", "sorrel"}))
	assert.Equal([]string{"bay", "chesnut", "sorrel"}, params.GetStringSlice("horses"))
	assert.Equal([]string(nil), params.GetStringSlice("ponies"))

	serenityCrew := map[string]string{
		"captain":   "Mal",
		"firstMate": "Zoe",
		"pilot":     "Wash",
		"mechanic":  "Kaylee",
		"muscle":    "Jayne",
	}
	require.Nil(params.SetStringMapString("serenity", serenityCrew))
	assert.Equal(serenityCrew, params.GetStringMapString("serenity"))
	assert.Equal(map[string]string(nil), params.GetStringMapString("beepop"))
}

func TestParamStore_Constraints(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	params := NewParamStore("test")

	require.Nil(params.DefineParam("name", reflect.String, "", ParamNecessity_REQUIRED))
	require.Nil(params.DefineParam("numCats", reflect.Int, "", ParamNecessity_OPTIONAL))

	require.Nil(params.SetString("name", "Eleanor Abernathy"))
	require.Nil(params.SetInt("numCats", 12))
	require.Nil(params.SetString("firstCatName", "Buster"))

	assert.Nil(params.Validate(false))

	err := params.Validate(true)
	require.NotNil(err)
	assert.Equal("unexpected parameter: \"firstCatName\"", err.Error())

	require.Nil(params.Clear())
	require.Nil(params.SetString("name", "Eleanor Abernathy"))
	require.Nil(params.SetString("numCats", "twelve"))

	err = params.Validate(false)
	require.NotNil(err)
	assert.Equal(
		"constraint failed for \"numCats\": expected type \"int\", but found \"string\"",
		err.Error(),
	)

	require.Nil(params.Clear())

	err = params.Validate(false)
	require.NotNil(err)
	assert.Equal("missing required parameter(s): name", err.Error())

	// overrides previsous definition
	require.Nil(params.DefineParam("numCats", reflect.Int, "", ParamNecessity_REQUIRED))

	err = params.Validate(false)
	require.NotNil(err)
	assert.Equal("missing required parameter(s): name, numCats", err.Error())
}

func TestParamStore_Viper_Map(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	v := viper.GetViper()
	v.SetConfigType("yaml")

	var configText = []byte(`
trust_anchor:
  store_name: memory
  store_params:
    store-type: memory
`)
	require.Nil(v.ReadConfig(bytes.NewBuffer(configText)))
	assert.Nil(nil)

	params := NewParamStore("kvstore")
	require.Nil(params.DefineParam("name", reflect.String, "trust_anchor.store_name", ParamNecessity_REQUIRED))
	require.Nil(params.DefineParam("params", reflect.Map, "trust_anchor.store_params", ParamNecessity_OPTIONAL))
	require.Nil(params.PopulateFromViper(v))
	require.Nil(params.Validate(false))

	assert.Equal(map[string]string{"store-type": "memory"}, params.GetStringMapString("params"))
}

func TestParamStore_Viper(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	params := NewParamStore("test")

	v := viper.GetViper()

	v.SetConfigType("yaml")

	var configText = []byte(`
crazy-cat-lady:
   name: Eleanor Abernathy
   cats:
     number: 30
     first:
       name:
          Buster
`)
	require.Nil(v.ReadConfig(bytes.NewBuffer(configText)))

	require.Nil(params.DefineParam("name", reflect.String, "crazy-cat-lady.name", ParamNecessity_REQUIRED))
	require.Nil(params.DefineParam("numCats", reflect.Int, "crazy-cat-lady.cats.number", ParamNecessity_OPTIONAL))

	require.Nil(params.PopulateFromViper(v))
	require.Nil(params.Validate(false))
	assert.Equal("Eleanor Abernathy", params.GetString("name"))

	require.Nil(params.DefineParam("DoesNotExist", reflect.String, "nope", ParamNecessity_REQUIRED))
	require.Nil(params.PopulateFromViper(v))

	err := params.Validate(false)
	require.NotNil(err)
	assert.Equal("missing required parameter(s): DoesNotExist", err.Error())
}

func TestParamStore_Map(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	params := NewParamStore("test")

	paramMap := map[string]interface{}{
		"name":    "Eleanor Abernathy",
		"numCats": 30,
	}

	err := params.PopulateFromMap(paramMap)
	require.NotNil(err)
	assert.Equal("no parameters have been defined", err.Error())

	require.Nil(params.DefineParam("name", reflect.String, "crazy-cat-lady.name", ParamNecessity_REQUIRED))
	require.Nil(params.DefineParam("numCats", reflect.Int, "crazy-cat-lady.cats.number", ParamNecessity_OPTIONAL))

	require.Nil(params.PopulateFromMap(paramMap))

	require.Nil(params.Validate(false))
	assert.Equal("Eleanor Abernathy", params.GetString("name"))

	require.Nil(params.PopulateFromMap(map[string]interface{}{}))
	assert.Equal(30, params.GetInt("numCats"))
}

func TestParamStore_Freeze(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	params := NewParamStore("test")

	require.Nil(params.AddParamDefinitions(map[string]*ParamDescription{
		"name": {
			Kind:     uint32(reflect.String),
			Path:     "crazy-cat-lady.name",
			Required: ParamNecessity_REQUIRED,
		},
		"numCats": {
			Kind:     uint32(reflect.Int),
			Path:     "crazy-cat-lady.cats.number",
			Required: ParamNecessity_OPTIONAL},
	}))
	names := params.GetParamNames()
	sort.Strings(names)
	assert.Equal([]string{"name", "numCats"}, names)
	assert.Equal("", params.GetString("name"))

	params.Freeze()

	// freezeing prefents new param definitions from being added
	err := params.DefineParam("firstCatName", reflect.String, "crazy-cat-lady.cats.first.name", ParamNecessity_OPTIONAL)
	require.NotNil(err)
	assert.Equal("cannot define parameter \"firstCatName\" -- store \"test\" is frozen", err.Error())

	err = params.AddParamDefinitions(map[string]*ParamDescription{
		"firstCatName": {
			Kind:     uint32(reflect.String),
			Path:     "crazy-cat-lady.cats.first.name",
			Required: ParamNecessity_OPTIONAL,
		},
	})
	require.NotNil(err)
	assert.Equal("cannot define parameter \"firstCatName\" -- store \"test\" is frozen", err.Error())

	// freezing does not prevent values from being set.
	require.Nil(params.SetString("name", "Eleanor Abernathy"))
	assert.Equal("Eleanor Abernathy", params.GetString("name"))

}

func TestParamStore_JSON(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	params := NewParamStore("test")
	paramMap := map[string]interface{}{
		"name":    "Eleanor Abernathy",
		"numCats": 30,
	}
	require.Nil(params.DefineParam("name", reflect.String, "crazy-cat-lady.name", ParamNecessity_REQUIRED))
	require.Nil(params.DefineParam("numCats", reflect.Int, "crazy-cat-lady.cats.number", ParamNecessity_OPTIONAL))
	require.Nil(params.PopulateFromMap(paramMap))

	data, err := json.Marshal(params)
	require.Nil(err)
	assert.JSONEq("{\"name\":\"test\",\"data\":{\"name\":\"Eleanor Abernathy\",\"numCats\":30},\"params\":{\"name\":{\"kind\":24,\"path\":\"crazy-cat-lady.name\",\"required\":\"REQUIRED\"},\"numCats\":{\"kind\":2,\"path\":\"crazy-cat-lady.cats.number\"}},\"required\":[\"name\"]}", string(data))

	newParams := NewParamStore("test2")
	err = json.Unmarshal(data, &newParams)
	require.Nil(err)

	assert.Equal(newParams.GetString("name"), params.GetString("name"))
	assert.Equal(newParams.GetInt("numCats"), params.GetInt("numCats"))
}
