// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"encoding/json"
	"errors"
	"fmt"
)

func sanitizeKV(key, val string) error {
	if err := sanitizeK(key); err != nil {
		return err
	}

	if err := sanitizeV(val); err != nil {
		return err
	}

	return nil
}

func sanitizeK(key string) error {
	if key == "" {
		return errors.New("the supplied key is empty")
	}

	return nil
}

func sanitizeV(val string) error {
	var tmp interface{}

	if err := json.Unmarshal([]byte(val), &tmp); err != nil {
		return fmt.Errorf("the supplied val contains invalid JSON: %w", err)
	}

	return nil
}
