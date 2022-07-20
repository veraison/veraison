// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"encoding/json"
	"fmt"
	"time"
)

type Status uint8

const (
	StatusWaiting Status = iota
	StatusProcessing
	StatusComplete
	StatusFailed
)

func (o Status) String() string {
	switch o {
	case StatusWaiting:
		return "waiting"
	case StatusProcessing:
		return "processing"
	case StatusComplete:
		return "complete"
	case StatusFailed:
		return "failed"
	}
	return "unknown"
}

func (o *Status) FromString(s string) error {
	switch s {
	case "waiting":
		*o = StatusWaiting
	case "processing":
		*o = StatusProcessing
	case "complete":
		*o = StatusComplete
	case "failed":
		*o = StatusFailed
	default:
		return fmt.Errorf("unknown status %s", s)
	}
	return nil
}

func (o Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

func (o *Status) UnmarshalJSON(b []byte) error {
	var s string

	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	return o.FromString(s)
}

type Blob struct {
	Type  string `json:"type"`
	Value []byte `json:"value"`
}

type ChallengeResponseSession struct {
	id       string
	Status   Status    `json:"status"`
	Nonce    []byte    `json:"nonce"`
	Expiry   time.Time `json:"expiry"`
	Accept   []string  `json:"accept"`
	Evidence *Blob     `json:"evidence,omitempty"`
	Result   *[]byte   `json:"result,omitempty"`
}

func (o *ChallengeResponseSession) SetEvidence(mt string, evidence []byte) {
	o.Evidence = &Blob{Type: mt, Value: evidence}
}

func (o *ChallengeResponseSession) SetStatus(status Status) {
	o.Status = status
}

func (o *ChallengeResponseSession) SetResult(result []byte) {
	tmp := make([]byte, len(result))
	copy(tmp, result)
	o.Result = &tmp
}
