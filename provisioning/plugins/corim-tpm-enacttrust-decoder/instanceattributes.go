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

	nodeID := inst.String()

	if nodeID == "" {
		return fmt.Errorf("could not extract node-id (UUID) from instance-id")
	}

	o.NodeID = nodeID

	return nil
}
