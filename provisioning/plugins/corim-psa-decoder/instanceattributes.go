// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"errors"

	"github.com/veraison/corim/comid"
	"github.com/veraison/eat"
)

type PSAInstanceAttributes struct {
	InstID eat.UEID
}

func (o *PSAInstanceAttributes) FromEnvironment(e comid.Environment) error {
	var err error

	if e.Instance == nil {
		return errors.New("expecting instance in environment")
	}

	o.InstID, err = e.Instance.GetUEID()

	return err
}
