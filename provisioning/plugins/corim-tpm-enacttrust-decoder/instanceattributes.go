// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"

	"github.com/veraison/corim/comid"
)

type InstanceAttributes struct {
	NodeID string
}

func (o *InstanceAttributes) FromEnvironment(e comid.Environment) error {
	inst := e.Instance

	if inst == nil {
		return fmt.Errorf("expecting instance in environment")
	}

	nodeID, err := e.Instance.GetUUID()
	if err != nil {
		return fmt.Errorf("could not extract node-id (UUID) from instance-id")
	}

	o.NodeID = nodeID.String()

	return nil
}
