package main

import (
	"fmt"

	"github.com/veraison/corim/comid"
)

type SwCompAttributes struct {
	AlgID  uint64
	Digest []byte
}

func (o *SwCompAttributes) FromMeasurement(m comid.Measurement) error {
	// extract digest and alg-id from mval
	d := m.Val.Digests

	if d == nil {
		return fmt.Errorf("measurement value has no digests")
	}

	if len(*d) != 1 {
		return fmt.Errorf("expecting exactly one digest")
	}

	o.AlgID = (*d)[0].HashAlgID
	o.Digest = (*d)[0].HashValue

	return nil
}
