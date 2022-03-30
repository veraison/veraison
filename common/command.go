// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"fmt"
	"strings"
)

type CommandArgs map[string]string

func (c *CommandArgs) Add(argText string) error {
	parts := strings.SplitN(argText, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("no \"=\" found in command argument: %s", argText)
	}
	key, value := parts[0], parts[1]

	if _, ok := (*c)[key]; ok {
		return fmt.Errorf("command argument \"%s\" already set", key)
	}

	(*c)[key] = value

	return nil
}
